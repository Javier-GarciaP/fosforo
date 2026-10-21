package markdown_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Javier-GarciaP/fosforo/internal/markdown"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

var update = flag.Bool("update", false, "reescribe los archivos golden")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falta %s (go test ./internal/markdown -update): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("la salida difiere de %s:\n--- obtenido ---\n%s\n--- esperado ---\n%s", path, got, want)
	}
}

func dump(d markdown.Doc) string {
	var sb strings.Builder
	for _, l := range d.Lines {
		sb.WriteString(strings.TrimRight(l.Plain(), " "))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func render(src string, w int) markdown.Doc { return markdown.Render(src, w) }

func plain(src string, w int) []string {
	var out []string
	for _, l := range render(src, w).Lines {
		out = append(out, l.Plain())
	}
	return out
}

func read(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// findRun devuelve el primer fragmento que contiene text.
func findRun(t *testing.T, d markdown.Doc, text string) markdown.Run {
	t.Helper()
	for _, l := range d.Lines {
		for _, r := range l.Runs {
			if strings.Contains(r.Text, text) {
				return r
			}
		}
	}
	t.Fatalf("no hay ningún fragmento con %q", text)
	return markdown.Run{}
}

func TestGoldenMaqueta07(t *testing.T) {
	d := render(read(t, "maqueta07.md"), 54)
	golden(t, "maqueta07", dump(d))

	h := findRun(t, d, "fosforo — editor")
	if !h.Bold || h.Role != markdown.RoleHeading {
		t.Errorf("encabezado en negrita: %+v", h)
	}
	if r := d.Lines[1].Runs[0]; r.Role != markdown.RoleRule || r.Text != strings.Repeat("─", 54) {
		t.Errorf("regla bajo H1: %+v", r)
	}
	if r := findRun(t, d, "Neovim"); r.Role != markdown.RoleLink || !r.Underline {
		t.Errorf("enlace: %+v", r)
	}
	if r := findRun(t, d, "Space"); r.Role != markdown.RoleInlineCode {
		t.Errorf("código en línea: %+v", r)
	}
	if r := findRun(t, d, "☑ "); r.Role != markdown.RoleDone {
		t.Errorf("casilla marcada: %+v", r)
	}
	if r := findRun(t, d, "☐ "); r.Role != markdown.RoleTodo {
		t.Errorf("casilla pendiente: %+v", r)
	}
	if r := findRun(t, d, "• "); r.Role != markdown.RoleBullet {
		t.Errorf("viñeta: %+v", r)
	}
	if r := findRun(t, d, "┃ "); r.Role != markdown.RoleQuoteBar {
		t.Errorf("barra de cita: %+v", r)
	}
	if r := findRun(t, d, "Nota:"); r.Role != markdown.RoleQuote {
		t.Errorf("texto de cita: %+v", r)
	}
	if r := findRun(t, d, "┌"); r.Role != markdown.RoleTableLine {
		t.Errorf("borde de tabla: %+v", r)
	}
	// el bloque de código lleva fondo propio
	for _, l := range d.Lines {
		if strings.Contains(l.Plain(), "go install") && l.Fill != markdown.FillCode {
			t.Error("el bloque de código tiene fondo")
		}
	}
}

func TestTaskLists(t *testing.T) {
	got := plain("- [x] Hecho\n- [ ] Pendiente\n- Normal\n", 40)
	want := []string{"☑ Hecho", "☐ Pendiente", "• Normal"}
	for i, w := range want {
		if i >= len(got) || got[i] != w {
			t.Errorf("línea %d = %q, quería %q", i, got[i], w)
		}
	}
	// tareas anidadas
	n := plain("- [ ] a\n  - [x] b\n", 40)
	if n[0] != "☐ a" || n[1] != "  ☑ b" {
		t.Errorf("anidadas: %q", n)
	}
}

func TestListsOrderedNestedAndLoose(t *testing.T) {
	got := plain("3. tres\n4. cuatro\n   - sub a\n   - sub b\n5. cinco\n", 40)
	want := []string{"3. tres", "4. cuatro", "   • sub a", "   • sub b", "5. cinco"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("lista ordenada y anidada:\n%q\nquería\n%q", got, want)
	}
	loose := plain("- uno\n\n- dos\n", 40)
	if len(loose) != 3 || loose[1] != "" {
		t.Errorf("lista con párrafos: separada por una línea en blanco: %q", loose)
	}
	tight := plain("- uno\n- dos\n", 40)
	if len(tight) != 2 {
		t.Errorf("lista compacta: %q", tight)
	}
	// una viñeta larga sangra sus líneas de continuación
	long := plain("- "+strings.Repeat("palabra ", 10), 20)
	for i, l := range long[1:] {
		if !strings.HasPrefix(l, "  ") {
			t.Errorf("continuación %d sin sangría: %q", i, l)
		}
	}
}

func TestInlineStyles(t *testing.T) {
	d := render("a **fuerte** b *cursiva* c ~~tachado~~ d `cod` e ***ambos***\n", 80)
	if r := findRun(t, d, "fuerte"); !r.Bold || r.Italic {
		t.Errorf("negrita: %+v", r)
	}
	if r := findRun(t, d, "cursiva"); !r.Italic || r.Bold {
		t.Errorf("cursiva: %+v", r)
	}
	if r := findRun(t, d, "tachado"); !r.Strike {
		t.Errorf("tachado: %+v", r)
	}
	if r := findRun(t, d, "cod"); r.Role != markdown.RoleInlineCode {
		t.Errorf("código: %+v", r)
	}
	if r := findRun(t, d, "ambos"); !r.Bold || !r.Italic {
		t.Errorf("ambos: %+v", r)
	}
	// los marcadores no se muestran
	if p := d.Lines[0].Plain(); strings.ContainsAny(p, "*~`") {
		t.Errorf("no deben verse los marcadores: %q", p)
	}
}

func TestLinksAndAutolinks(t *testing.T) {
	d := render("ver [la guía](docs/README.md) y https://example.com o <https://go.dev>\n", 80)
	p := d.Lines[0].Plain()
	if !strings.Contains(p, "la guía") || strings.Contains(p, "docs/README.md") {
		t.Errorf("del enlace solo se muestra el texto: %q", p)
	}
	if !strings.Contains(p, "https://example.com") || !strings.Contains(p, "https://go.dev") {
		t.Errorf("las URL sueltas se muestran: %q", p)
	}
	if r := findRun(t, d, "https://example.com"); r.Role != markdown.RoleLink {
		t.Errorf("autoenlace: %+v", r)
	}
}

func TestHeadingsAndRules(t *testing.T) {
	d := render("# Uno\n\n## Dos\n\n### Tres\n\ntexto\n\n---\n\n#### Cuatro\n", 30)
	lines := dump(d)
	if strings.Count(lines, strings.Repeat("─", 30)) != 3 { // H1, H2 y la línea horizontal; H3 y H4 no llevan regla
		t.Errorf("reglas:\n%s", lines)
	}
	for _, name := range []string{"Uno", "Dos", "Tres", "Cuatro"} {
		if r := findRun(t, d, name); !r.Bold {
			t.Errorf("%s en negrita", name)
		}
	}
}

func TestBlockquotes(t *testing.T) {
	got := plain("> uno\n> dos\n>\n> > anidada\n", 40)
	want := []string{"┃ uno dos", "┃ ", "┃ ┃ anidada"}
	if len(got) != 3 {
		t.Fatalf("%q", got)
	}
	for i, w := range want {
		if strings.TrimRight(got[i], " ") != strings.TrimRight(w, " ") {
			t.Errorf("cita %d = %q, quería %q", i, got[i], w)
		}
	}
}

func TestTablesAlignmentAndBorders(t *testing.T) {
	src := "| Izq | Centro | Der |\n|:----|:------:|----:|\n| a | b | c |\n| largo texto | x | 12345 |\n"
	got := plain(src, 60)
	if !strings.HasPrefix(got[0], "┌") || !strings.HasPrefix(got[len(got)-1], "└") {
		t.Fatalf("bordes:\n%s", strings.Join(got, "\n"))
	}
	w := len([]rune(got[0]))
	for i, l := range got {
		if len([]rune(l)) != w {
			t.Errorf("la fila %d tiene otro ancho: %q", i, l)
		}
	}
	// alineación: «c» a la derecha, «b» centrada, «a» a la izquierda
	row := got[3]
	if !strings.Contains(row, "│ a ") || !strings.Contains(row, " c │") {
		t.Errorf("alineaciones izquierda y derecha: %q", row)
	}
	bi := strings.Index(row, "b")
	if bi < strings.Index(row, "│")+8 { // bien dentro de la columna central
		t.Errorf("centrado: %q", row)
	}
	golden(t, "table", strings.Join(got, "\n")+"\n")
}

func TestTablesWithWideCharactersStayAligned(t *testing.T) {
	d := render("| Nombre | Nota |\n|---|---|\n| 日本語 | 🙂 ok |\n| ab | 中文テキスト |\n", 50)
	w := d.Lines[0].Width()
	for i, l := range d.Lines {
		if l.Width() != w {
			t.Errorf("fila %d mide %d celdas, la primera %d:\n%s", i, l.Width(), w, dump(d))
		}
	}
}

func TestWideTablesAreTruncatedWithEllipsis(t *testing.T) {
	src := "| Tecla | Descripción larguísima de la acción | Notas adicionales extensas |\n|---|---|---|\n| Space f f | Buscar archivos del proyecto con vista previa | Muy útil |\n"
	for _, w := range []int{30, 40, 60} {
		d := render(src, w)
		for i, l := range d.Lines {
			if l.Width() > w {
				t.Errorf("w=%d fila %d se desborda: %d", w, i, l.Width())
			}
		}
		if w < 60 && !strings.Contains(dump(d), "…") {
			t.Errorf("w=%d: las celdas recortadas terminan en «…»:\n%s", w, dump(d))
		}
	}
	wide := render(src, 120)
	if strings.Contains(dump(wide), "…") {
		t.Error("con ancho suficiente no se recorta nada")
	}
}

func TestInlineCodeInsideTable(t *testing.T) {
	d := render("| A |\n|---|\n| `Space f f` |\n", 40)
	if r := findRun(t, d, "Space"); r.Role != markdown.RoleInlineCode {
		t.Errorf("%+v", r)
	}
}

func TestCodeBlocksUseTheSameColorsAsTheEditor(t *testing.T) {
	code := "package main\n\nfunc f(x int) string {\n\treturn \"a\" // c\n}"
	d := render("```go\n"+code+"\n```\n", 60)
	want, _ := syntax.Tokenize("Go", strings.ReplaceAll(code, "\t", "    "))
	_ = want
	// se comparan las categorías por texto: cada palabra clave, cadena y comentario de la salida tiene la misma categoría que Chroma
	check := map[string]syntax.Kind{"package": syntax.Keyword, "func": syntax.Keyword, "return": syntax.Keyword, "\"a\"": syntax.String, "// c": syntax.Comment, "int": syntax.Type, "string": syntax.Type}
	for text, kind := range check {
		found := false
		for _, l := range d.Lines {
			for _, r := range l.Runs {
				if strings.Contains(r.Text, text) && r.Syntax == kind {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("%q debe tener la categoría %v como en el editor:\n%s", text, kind, dump(d))
		}
	}
	for _, l := range d.Lines {
		if l.Fill != markdown.FillCode || !strings.HasPrefix(l.Plain(), " ") {
			t.Errorf("las líneas de código llevan fondo y relleno: %q", l.Plain())
		}
	}
}

func TestCodeBlockEdgeCases(t *testing.T) {
	// sin lenguaje, con tabuladores, línea larga, vacío, indentado
	d := render("```\na\tb\n"+strings.Repeat("x", 100)+"\n```\n\n    indentado\n\n```go\n```\n", 30)
	p := dump(d)
	if !strings.Contains(p, " a    b") {
		t.Errorf("los tabuladores se expanden: %q", p)
	}
	for _, l := range d.Lines {
		if l.Width() > 30 {
			t.Errorf("línea ancha: %d", l.Width())
		}
	}
	if !strings.Contains(p, "…") {
		t.Error("las líneas de código largas se recortan con …")
	}
	if !strings.Contains(p, "indentado") {
		t.Error("bloque indentado")
	}
}

func TestUnsupportedContentDegrades(t *testing.T) {
	d := render("![logo](logo.png) y ![](x.png)\n", 40)
	if p := d.Lines[0].Plain(); !strings.Contains(p, "[imagen: logo]") || !strings.Contains(p, "[imagen: ]") {
		t.Errorf("imagen: %q", p)
	}
	if r := findRun(t, d, "[imagen: logo]"); r.Role != markdown.RoleImage {
		t.Errorf("rol de imagen: %+v", r)
	}
	if d.Omitted {
		t.Error("las imágenes no cuentan como contenido omitido")
	}

	html := render("antes\n\n<div align=\"center\">\n<img src=\"a.png\">\n</div>\n\ndespués <b>negrita</b> fin<br>nueva línea\n", 50)
	p := dump(html)
	if strings.Contains(p, "<div") || strings.Contains(p, "<img") || strings.Contains(p, "<b>") {
		t.Errorf("el HTML crudo se omite:\n%s", p)
	}
	if !strings.Contains(p, "antes") || !strings.Contains(p, "después negrita fin") || !strings.Contains(p, "nueva línea") {
		t.Errorf("el texto alrededor se conserva y <br> salta de línea:\n%s", p)
	}
	if !html.Omitted || !strings.Contains(p, "omitido") {
		t.Errorf("se avisa del contenido omitido:\n%s", p)
	}

	mer := render("```mermaid\ngraph TD; A-->B\n```\n", 50)
	if p := dump(mer); strings.Contains(p, "graph TD") || !strings.Contains(p, "diagrama mermaid") || !mer.Omitted {
		t.Errorf("mermaid:\n%s", p)
	}
}

func TestWrapping(t *testing.T) {
	d := render("Un párrafo bastante largo que debe ajustarse al ancho del panel sin cortar las palabras por la mitad.\n", 24)
	for _, l := range d.Lines {
		if l.Width() > 24 || strings.HasSuffix(l.Plain(), " ") || strings.HasPrefix(l.Plain(), " ") {
			t.Errorf("línea mal ajustada: %q", l.Plain())
		}
	}
	if joined := strings.Join(plain("Un párrafo bastante largo que debe ajustarse.", 12), " "); joined != "Un párrafo bastante largo que debe ajustarse." {
		t.Errorf("no se pierde texto al ajustar: %q", joined)
	}
	// saltos de línea suaves y forzados
	got := plain("uno\ndos  \ntres\\\ncuatro\n", 40)
	if strings.Join(got, "|") != "uno dos|tres|cuatro" {
		t.Errorf("saltos: %q", got)
	}
	// una palabra más larga que la línea se parte
	long := plain(strings.Repeat("x", 50), 20)
	if len(long) != 3 || long[0] != strings.Repeat("x", 20) {
		t.Errorf("palabra larga: %q", long)
	}
	// ancho mínimo
	for _, l := range render("hola mundo\n", 1).Lines {
		if l.Width() > 10 {
			t.Error("el ancho mínimo es 10")
		}
	}
}

func TestWidthWithCJKAndEmoji(t *testing.T) {
	src := "日本語のテキストが長い文章として続きます。🙂 emoji 👨‍👩‍👧 compuesto y más texto normal que ajustar.\n\n- 項目 一二三四五六七八九十 の長いリスト項目です\n\n> 引用文もとても長くなります、折り返しが必要です。\n"
	for _, w := range []int{16, 21, 30, 40, 80} {
		d := render(src, w)
		for i, l := range d.Lines {
			if l.Width() > w {
				t.Errorf("w=%d línea %d mide %d: %q", w, i, l.Width(), l.Plain())
			}
		}
	}
	golden(t, "cjk", dump(render(src, 30)))
}

func TestRealWorldFixtures(t *testing.T) {
	for _, name := range []string{"fosforo_readme.md", "docs_readme.md", "contributing.md"} {
		src := read(t, name)
		for _, w := range []int{30, 40, 54, 80, 100, 160} {
			d := render(src, w)
			if len(d.Lines) == 0 {
				t.Fatalf("%s: documento vacío", name)
			}
			for i, l := range d.Lines {
				if l.Width() > w {
					t.Errorf("%s w=%d línea %d se desborda (%d): %q", name, w, i, l.Width(), l.Plain())
				}
			}
			prev := -1
			for _, a := range d.Anchors {
				if a.Out < prev || a.Out >= len(d.Lines) || a.Src < 0 || a.Src >= d.SrcLines {
					t.Errorf("%s w=%d ancla inválida %+v", name, w, a)
				}
				prev = a.Out
			}
		}
	}
	// discrepancias conocidas (anotadas en el diseño): el README de presentación usa HTML para el logo y las insignias
	if d := render(read(t, "fosforo_readme.md"), 80); !d.Omitted {
		t.Error("el README con <picture> y <p align> debe avisar de contenido omitido")
	}
	if d := render(read(t, "contributing.md"), 80); !strings.Contains(dump(d), "Contribuir a fosforo") {
		t.Error("el título de CONTRIBUTING")
	}
	golden(t, "contributing_54", dump(render(read(t, "contributing.md"), 54)))
}

func TestEmptyAndDegenerateDocuments(t *testing.T) {
	for _, src := range []string{"", "\n", "   \n\n", "<div></div>", "---", "\u00a0"} {
		d := render(src, 40)
		for _, l := range d.Lines {
			if l.Width() > 40 {
				t.Errorf("%q: %q", src, l.Plain())
			}
		}
	}
	if d := render("", 40); len(d.Lines) != 0 || d.SrcLines != 1 {
		t.Errorf("vacío: %d líneas, SrcLines=%d", len(d.Lines), d.SrcLines)
	}
}

func TestLargeDocumentRendersQuickly(t *testing.T) {
	src := strings.Repeat("# Título\n\nUn párrafo con **negrita**, `código` y [enlace](x) que ocupa bastante espacio en la línea.\n\n- a\n- b\n\n```go\nfunc f() {}\n```\n\n", 4000)
	start := time.Now()
	d := render(src, 80)
	if el := time.Since(start); el > 5*time.Second {
		t.Errorf("renderizar %d KiB tardó %v", len(src)/1024, el)
	}
	if len(d.Lines) < 10000 {
		t.Errorf("líneas = %d", len(d.Lines))
	}
}

func TestMaxBytesConstant(t *testing.T) {
	if markdown.MaxBytes != 2<<20 {
		t.Error("el límite es de 2 MiB")
	}
}

func TestAnchorsAndMapping(t *testing.T) {
	src := "# Uno\n\nparrafo\nde varias\nlineas\n\n## Dos\n\n- a\n- b\n- c\n\n## Tres\n\nfin\n"
	d := render(src, 40)
	heads := []int{}
	for _, a := range d.Anchors {
		if a.Heading {
			heads = append(heads, a.Src)
		}
	}
	if len(heads) != 3 || heads[0] != 0 || heads[1] != 6 || heads[2] != 12 {
		t.Errorf("anclas de encabezados = %v (líneas del fuente 0, 6 y 12)", heads)
	}
	// un encabezado del fuente cae exactamente en su línea del resultado
	for _, a := range d.Anchors {
		if got := d.OutLine(a.Src); got != a.Out {
			t.Errorf("OutLine(%d) = %d, quería %d", a.Src, got, a.Out)
		}
	}
	// el mapeo es creciente
	prev := -1
	for s := 0; s < d.SrcLines; s++ {
		o := d.OutLine(s)
		if o < prev {
			t.Errorf("OutLine no decrece: %d→%d en la línea %d", prev, o, s)
		}
		prev = o
	}
	// la inversa devuelve el bloque correcto
	for _, a := range d.Anchors {
		if got := d.SrcLine(a.Out); got != a.Src {
			t.Errorf("SrcLine(%d) = %d, quería %d", a.Out, got, a.Src)
		}
	}
	// encabezado más cercano por encima
	two := d.Anchors[2].Out // «Dos» (la regla va dentro del bloque)
	if src, ok := d.HeadingAt(two + 2); !ok || src != 6 {
		t.Errorf("HeadingAt dentro de la sección Dos = %d %v", src, ok)
	}
	if src, ok := d.HeadingAt(0); !ok || src != 0 {
		t.Errorf("HeadingAt(0) = %d %v", src, ok)
	}
	if src, ok := d.HeadingAt(9999); !ok || src != 12 {
		t.Errorf("HeadingAt tras el final = %d %v", src, ok)
	}
}

func TestMappingWithoutHeadings(t *testing.T) {
	src := strings.Repeat("un párrafo\n\n", 20)
	d := render(src, 40)
	if _, ok := d.HeadingAt(5); ok {
		t.Error("sin encabezados HeadingAt devuelve false")
	}
	mid := d.OutLine(d.SrcLines / 2)
	if mid < len(d.Lines)/3 || mid > 2*len(d.Lines)/3 {
		t.Errorf("la mitad del fuente cae hacia la mitad del resultado: %d de %d", mid, len(d.Lines))
	}
	// sin ningún bloque (solo HTML omitido) se usa la proporción global
	h := render("<div>\n\n\n\n</div>\n", 40)
	_ = h.OutLine(2)
	_ = h.SrcLine(0)
	empty := render("", 40)
	if empty.OutLine(3) != 0 || empty.SrcLine(3) != 0 {
		t.Error("documento vacío")
	}
}
