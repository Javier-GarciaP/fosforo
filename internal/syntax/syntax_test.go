package syntax_test

import (
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// kinds devuelve, para una línea, «texto:categoría» de cada tramo coloreado.
func kinds(src string, spans []syntax.Span) []string {
	var out []string
	for _, sp := range spans {
		out = append(out, src[sp.Start:sp.End]+":"+sp.Kind.String())
	}
	return out
}

func tokenizeLine(t *testing.T, lang, text string, line int) []string {
	t.Helper()
	lines, err := syntax.Tokenize(lang, text)
	if err != nil {
		t.Fatal(err)
	}
	src := strings.Split(text, "\n")
	if line >= len(lines) {
		t.Fatalf("línea %d fuera de %d", line, len(lines))
	}
	return kinds(src[line], lines[line])
}

func has(got []string, want ...string) bool {
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

const goSrc = "package keymap\n\n// Lookup resuelve una secuencia.\nfunc (n *Node) Lookup(seq []rune) (string, bool) {\n\tx := \"hola\\n\"\n\tif y := 12; y > 3.5 {\n\t\treturn \"\", false\n\t}\n\treturn nil, true\n}\n"

func TestTokenizeGoKinds(t *testing.T) {
	for line, want := range map[int][]string{
		0: {"package:keyword"},
		2: {"// Lookup resuelve una secuencia.:comment"},
		3: {"func:keyword", "Lookup:function", "string:type", "bool:type", "rune:type"},
		4: {"\"hola\\n\":string"},
		5: {"if:keyword", "12:number", "3.5:number"},
		6: {"return:keyword", "false:constant"},
		8: {"return:keyword", "nil:constant", "true:constant"},
	} {
		got := tokenizeLine(t, "Go", goSrc, line)
		if !has(got, want...) {
			t.Errorf("línea %d (%q): %v no contiene %v", line, strings.Split(goSrc, "\n")[line], got, want)
		}
	}
}

func TestTokenizeReturnsOneEntryPerLine(t *testing.T) {
	for _, text := range []string{"", "a", "a\n", "a\n\nb", goSrc, "\n\n\n"} {
		lines, err := syntax.Tokenize("Go", text)
		if err != nil {
			t.Fatal(err)
		}
		if want := len(strings.Split(text, "\n")); len(lines) != want {
			t.Errorf("%q: %d líneas, quería %d", text, len(lines), want)
		}
	}
}

func TestSpansAreOrderedInBoundsAndMerged(t *testing.T) {
	lines, _ := syntax.Tokenize("Go", goSrc)
	src := strings.Split(goSrc, "\n")
	for i, spans := range lines {
		prev := 0
		for _, sp := range spans {
			if sp.Start < prev || sp.End <= sp.Start || sp.End > len(src[i]) {
				t.Errorf("línea %d: tramo inválido %+v (longitud %d, anterior termina en %d)", i, sp, len(src[i]), prev)
			}
			if sp.Kind == syntax.Plain {
				t.Errorf("línea %d: los tramos sin categoría no se guardan: %+v", i, sp)
			}
			prev = sp.End
		}
	}
}

func TestMultilineBlockComment(t *testing.T) {
	src := "int x;\n/* uno\n   dos\n   tres */\nint y;\n"
	lines, _ := syntax.Tokenize("C", src)
	parts := strings.Split(src, "\n")
	for _, i := range []int{1, 2, 3} {
		if got := kinds(parts[i], lines[i]); len(got) != 1 || !strings.HasSuffix(got[0], ":comment") {
			t.Errorf("línea %d debe ser comentario: %v", i, got)
		}
	}
	if got := kinds(parts[4], lines[4]); strings.Contains(strings.Join(got, ","), "comment") {
		t.Errorf("tras el cierre ya no es comentario: %v", got)
	}
	// al abrir un comentario antes, todo lo siguiente cambia
	open := "/*\n" + goSrc + "*/\n"
	lines2, _ := syntax.Tokenize("Go", open)
	p2 := strings.Split(open, "\n")
	for i := 1; i < len(p2)-2; i++ {
		for _, sp := range lines2[i] {
			if sp.Kind != syntax.Comment {
				t.Errorf("línea %d (%q): dentro de un comentario abierto todo es comentario: %+v", i, p2[i], sp)
			}
		}
	}
}

func TestMultilineRawString(t *testing.T) {
	src := "x := `uno\ndos\ntres`\ny := 1\n"
	lines, _ := syntax.Tokenize("Go", src)
	parts := strings.Split(src, "\n")
	for _, i := range []int{1, 2} {
		if got := kinds(parts[i], lines[i]); len(got) != 1 || !strings.HasSuffix(got[0], ":string") {
			t.Errorf("línea %d debe ser cadena: %v", i, got)
		}
	}
	if got := kinds(parts[3], lines[3]); !has(got, "1:number") {
		t.Errorf("tras la cadena cruda: %v", got)
	}
}

func TestSupportedLanguages(t *testing.T) {
	for _, tc := range []struct {
		file, src string
		line      int
		want      string
	}{
		{"main.go", "func main() {}\n", 0, "func:keyword"},
		{"README.md", "# Título\n\ntexto **fuerte**\n", 0, ":function"},
		{"datos.json", "{\"clave\": 12, \"x\": true}\n", 0, "12:number"},
		{"config.yaml", "clave: valor\nlista:\n  - 1\n", 0, "clave:tag"},
		{"Cargo.toml", "[package]\nname = \"fos\"\nversion = 1\n", 1, "\"fos\":string"},
		{"app.js", "const x = 5; // nota\n", 0, "const:keyword"},
		{"app.ts", "let n: number = 3;\n", 0, "let:keyword"},
		{"s.py", "def f(x):\n    return 1\n", 0, "def:keyword"},
		{"lib.rs", "fn main() { let x = 1; }\n", 0, "fn:keyword"},
		{"run.sh", "if [ -f x ]; then echo hola; fi\n", 0, "if:keyword"},
		{"Makefile", "all:\n\tgo build\n", 0, ":"},
		{"Dockerfile", "FROM golang:1.27\n", 0, "FROM:keyword"},
		{"style.css", "a { color: red; }\n", 0, ":"},
		{"index.html", "<p class=\"a\">hola</p>\n", 0, ":"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			lang := syntax.Detect(tc.file)
			if lang == "" {
				t.Fatalf("no se detecta el lenguaje de %s", tc.file)
			}
			lines, err := syntax.Tokenize(lang, tc.src)
			if err != nil || len(lines) == 0 {
				t.Fatalf("tokenizar: %v", err)
			}
			if tc.want == ":" || strings.HasPrefix(tc.want, ":") {
				if len(lines[tc.line]) == 0 {
					t.Errorf("%s (%s): la línea %d debe tener color", tc.file, lang, tc.line)
				}
				return
			}
			got := tokenizeLine(t, lang, tc.src, tc.line)
			if !has(got, tc.want) {
				t.Errorf("%s (%s) línea %d: %v no contiene %q", tc.file, lang, tc.line, got, tc.want)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	for file, want := range map[string]string{
		"trie.go": "Go", "TRIE.GO": "Go", "dir/sub/main.go": "Go", "go.mod": "Go", "go.work": "Go", "go.sum": "",
		"Makefile": "Makefile", "README.md": "markdown", "a.json": "JSON", "a.yml": "YAML", "a.toml": "TOML",
		"a.py": "Python", "a.rs": "Rust", "a.ts": "TypeScript", "a.sh": "Bash",
		"notas.xyz": "", "sinextension": "", "": "", ".": "", "notas.txt": "", ".gitignore": "",
	} {
		if got := syntax.Detect(file); got != want {
			t.Errorf("Detect(%q) = %q, quería %q", file, got, want)
		}
	}
}

func TestUnknownLanguageAndPlain(t *testing.T) {
	for _, lang := range []string{"", "LenguajeInventado"} {
		lines, err := syntax.Tokenize(lang, "texto cualquiera\n")
		if err != nil || lines != nil {
			t.Errorf("lang %q: debe devolver nil sin error: %v %v", lang, lines, err)
		}
	}
}

func TestSafetyLimits(t *testing.T) {
	// archivo mayor de 2 MiB: sin resaltado
	big := strings.Repeat("func main() {}\n", syntax.MaxFileBytes/15+10)
	if lines, err := syntax.Tokenize("Go", big); err != nil || lines != nil {
		t.Errorf("un archivo de %d bytes no se resalta", len(big))
	}
	// justo debajo del límite sí
	ok := strings.Repeat("x := 1\n", 1000)
	if lines, _ := syntax.Tokenize("Go", ok); len(lines) != 1001 {
		t.Errorf("líneas = %d", len(lines))
	}

	// una línea gigantesca no tiene color; el resto sí
	long := "var a = \"" + strings.Repeat("x", 100000) + "\"\nfunc main() {}\n"
	lines, err := syntax.Tokenize("Go", long)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || len(lines[0]) != 0 {
		t.Errorf("la línea gigante queda plana: %d tramos", len(lines[0]))
	}
	if got := kinds("func main() {}", lines[1]); !has(got, "func:keyword") {
		t.Errorf("el resto del archivo sigue coloreado: %v", got)
	}
	// el límite es por línea: 2000 exactos sí, 2001 no
	at := "// " + strings.Repeat("a", syntax.MaxLineBytes-3) + "\n"
	over := "// " + strings.Repeat("a", syntax.MaxLineBytes-2) + "\n"
	l1, _ := syntax.Tokenize("Go", at)
	l2, _ := syntax.Tokenize("Go", over)
	if len(l1[0]) != 1 || len(l2[0]) != 0 {
		t.Errorf("límite de línea: %v %v", l1[0], l2[0])
	}
}

func TestKindString(t *testing.T) {
	if syntax.Keyword.String() != "keyword" || syntax.Kind(200).String() != "?" || syntax.Plain.String() != "plain" {
		t.Error("String")
	}
}

func BenchmarkTokenize2MiB(b *testing.B) {
	src := strings.Repeat("func (n *Node) Lookup(seq []rune) (string, bool) { return \"x\", 12 > 3 } // c\n", 25000)
	if len(src) > syntax.MaxFileBytes {
		src = src[:syntax.MaxFileBytes]
	}
	b.SetBytes(int64(len(src)))
	for i := 0; i < b.N; i++ {
		if _, err := syntax.Tokenize("Go", src); err != nil {
			b.Fatal(err)
		}
	}
}
