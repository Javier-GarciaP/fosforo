package ui_test

import (
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

func highlightedView(t *testing.T) (ui.EditorView, [][]syntax.Span) {
	t.Helper()
	spans, err := syntax.Tokenize("Go", trieGo)
	if err != nil {
		t.Fatal(err)
	}
	b := buffer.FromString("", trieGo)
	return ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "NORMAL", PathLabel: "internal/keymap/trie.go", Highlight: spans}, spans
}

// visible hace legibles las secuencias de escape en los goldens de color.
func visible(s string) string { return strings.ReplaceAll(s, "\x1b", "␛") }

func TestGoldenSyntaxColorsDarkAndLight(t *testing.T) {
	v, _ := highlightedView(t)
	golden(t, "syntax_dark_ansi", visible(ui.RenderEditor(v, 80, 20, ui.Dark()))+"\n")
	golden(t, "syntax_light_ansi", visible(ui.RenderEditor(v, 80, 20, ui.Light()))+"\n")
}

func TestSyntaxColorsFollowThePalette(t *testing.T) {
	v, _ := highlightedView(t)
	raw := ui.RenderEditor(v, 100, 24, ui.Dark())
	rows := strings.Split(raw, "\n")
	// fila 0: «package»; fila 7: comentario; fila 8: func, Lookup, tipos; fila 15: return y constantes
	for _, tc := range []struct {
		row  int
		text string
		fg   string // r;g;b esperado
		what string
	}{
		{0, "package", "204;168;246", "palabra clave en mauve"},
		{7, "// Lookup", "117;128;156", "comentario en gris"},
		{8, "func", "204;168;246", "func en mauve"},
		{8, "Lookup", "127;193;254", "función en azul"},
		{8, "string", "213;184;94", "tipo en amarillo"},
		{15, "return", "204;168;246", "return en mauve"},
	} {
		sgr := lastSGRBefore(rows[tc.row], tc.text)
		if !strings.Contains(sgr, "38;2;"+tc.fg) {
			t.Errorf("%s: %q tiene el SGR %q", tc.what, tc.text, sgr)
		}
	}
	// el texto sin categoría (p. ej. «Node» en la fila 2) conserva el color de texto
	if sgr := lastSGRBefore(rows[2], "Node"); !strings.Contains(sgr, "38;2;207;215;235") {
		t.Errorf("texto sin categoría: %q", sgr)
	}
}

func TestSyntaxColorsKeepLineAndSelectionBackgrounds(t *testing.T) {
	v, _ := highlightedView(t)
	// la línea del cursor es la 12: «cur.children[r]» es texto plano, pero «next, ok» no tiene color; usamos la 16 con «return»
	v.Cursor = buffer.Pos{Line: 15, Col: 0}
	rows := strings.Split(ui.RenderEditor(v, 100, 24, ui.Dark()), "\n")
	sgr := lastSGRBefore(rows[15], "return")
	if !strings.Contains(sgr, "38;2;204;168;246") || !strings.Contains(sgr, "48;2;35;44;66") {
		t.Errorf("sobre la línea actual se conserva el color del token y el fondo de la línea: %q", sgr)
	}
	// selección Visual: mismo color de token sobre el fondo de selección
	v.Cursor = buffer.Pos{Line: 15, Col: 8}
	v.Selection = &ui.Selection{Start: buffer.Pos{Line: 15, Col: 1}, End: buffer.Pos{Line: 15, Col: 6}}
	rows = strings.Split(ui.RenderEditor(v, 100, 24, ui.Dark()), "\n")
	sgr = lastSGRBefore(rows[15], "return")
	if !strings.Contains(sgr, "38;2;204;168;246") || !strings.Contains(sgr, "48;2;50;61;92") {
		t.Errorf("sobre la selección se conserva el color del token: %q", sgr)
	}
}

func TestHighlightShorterThanBufferAndOffsets(t *testing.T) {
	src := "func f() { s := \"ñandú ✓\" + 12 }\nsegunda\n"
	b := buffer.FromString("", src)
	spans, _ := syntax.Tokenize("Go", src)
	// sin tramos o con tramos de menos líneas no debe fallar
	for _, hl := range [][][]syntax.Span{nil, spans[:1], {}} {
		out := plain(ui.EditorView{Buf: b, Mode: "NORMAL", Highlight: hl}, 60, 8)
		if !strings.Contains(out, "ñandú ✓") {
			t.Errorf("el texto se dibuja igual:\n%s", out)
		}
	}
	// los desplazamientos son en bytes: tras caracteres de varios bytes el número sigue coloreado donde toca
	for _, left := range []int{0, 6} {
		raw := ui.RenderEditor(ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 1}, Mode: "NORMAL", Highlight: spans, Left: left}, 60, 8, ui.Dark())
		row := strings.Split(raw, "\n")[0]
		if sgr := lastSGRBefore(row[strings.Index(row, "ñandú"):], "12"); !strings.Contains(sgr, "245;165;116") {
			t.Errorf("left=%d: el número 12 va en melocotón: %q", left, sgr)
		}
		if sgr := lastSGRBefore(row, "ñandú"); !strings.Contains(sgr, "129;208;150") && !strings.Contains(sgr, "127;208;150") {
			t.Errorf("left=%d: la cadena con caracteres multibyte va en verde: %q", left, sgr)
		}
	}
}

func TestSyntaxColorTableCoversEveryKind(t *testing.T) {
	th := ui.Dark()
	seen := map[string]bool{}
	for k := syntax.Plain; k <= syntax.Error; k++ {
		b := buffer.FromString("", "abc\n")
		v := ui.EditorView{Buf: b, Mode: "NORMAL", Highlight: [][]syntax.Span{{{Start: 0, End: 3, Kind: k}}}}
		row := strings.Split(ui.RenderEditor(v, 40, 8, th), "\n")[0]
		seen[lastSGRBefore(row, "abc")] = true
	}
	if len(seen) < 8 {
		t.Errorf("las categorías deben distinguirse por color: %d variantes", len(seen))
	}
}
