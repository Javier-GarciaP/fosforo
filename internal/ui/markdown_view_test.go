package ui_test

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/markdown"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

func mdFixture(t *testing.T, width int) (string, markdown.Doc) {
	t.Helper()
	b, err := os.ReadFile("../markdown/testdata/maqueta07.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b), markdown.Render(string(b), width)
}

func TestGoldenMarkdownPreview(t *testing.T) {
	// Maqueta 07: vista previa a pantalla completa con el explorador a la izquierda.
	src, doc := mdFixture(t, 54)
	v := ui.EditorView{Buf: buffer.FromString("", src), Mode: "PREVIEW", PathLabel: "README.md",
		PathHint: "(e: editar   Space m s: dividir)", Right: "markdown  utf-8  LF  28 líneas  0%",
		Explorer: fosforoTree(ui.IconsNone), Markdown: &ui.MarkdownView{Doc: &doc}}
	golden(t, "markdown_preview", plain(v, 120, 36))

	raw := ui.RenderEditor(v, 120, 36, ui.Dark())
	rows := strings.Split(raw, "\n")
	find := func(s string) string {
		for _, r := range rows {
			if strings.Contains(ansi.Strip(r), s) {
				return r
			}
		}
		t.Fatalf("no hay una fila con %q", s)
		return ""
	}
	if sgr := lastSGRBefore(find("go install"), "go install"); !strings.Contains(sgr, "48;2;10;14;24") {
		t.Errorf("fondo del bloque de código: %q", sgr)
	}
	// una línea de código llena todo el ancho del documento con su fondo
	codeRow := find("go install")
	if strings.Count(codeRow, "48;2;10;14;24") < 1 {
		t.Error("el fondo se extiende")
	}
	if sgr := lastSGRBefore(find("Neovim"), "Neovim"); !strings.Contains(sgr, "38;2;127;193;254") || !strings.Contains(sgr, ";4") {
		t.Errorf("enlace en azul y subrayado: %q", sgr)
	}
	if sgr := lastSGRBefore(find("☑"), "☑"); !strings.Contains(sgr, "38;2;127;208;150") {
		t.Errorf("casilla marcada en verde: %q", sgr)
	}
	if sgr := lastSGRBefore(find("Nota:"), "Nota:"); !strings.Contains(sgr, "38;2;117;128;156") {
		t.Errorf("cita en gris: %q", sgr)
	}
	if sgr := lastSGRBefore(find("┃"), "┃"); !strings.Contains(sgr, "38;2;127;193;254") {
		t.Errorf("barra de cita en azul: %q", sgr)
	}
}

func TestGoldenMarkdownSplit(t *testing.T) {
	// Maqueta 08: fuente a la izquierda y vista previa a la derecha, con el cursor en el fuente.
	src, _ := mdFixture(t, 40)
	doc := markdown.Render(src, 40)
	spans, _ := syntax.Tokenize("markdown", src)
	v := ui.EditorView{Buf: buffer.FromString("", src), Cursor: buffer.Pos{Line: 9, Col: 6}, Mode: "NORMAL", PathLabel: "README.md",
		PathHint: "| vista previa sincronizada", Highlight: spans, Markdown: &ui.MarkdownView{Doc: &doc, Split: true}}
	out := plain(v, 120, 34)
	golden(t, "markdown_split", out)
	rows := strings.Split(out, "\n")
	left, right, ok := strings.Cut(rows[0], "│")
	if !ok || !strings.Contains(left, "# fosforo") || !strings.Contains(right, "fosforo — editor") {
		t.Errorf("fuente a la izquierda y vista previa a la derecha: %q", rows[0])
	}
	if x, y, ok := ui.CursorCell(v, 120, 34); !ok || x >= 60 || y != 9 {
		t.Errorf("el cursor está en el fuente: %d,%d,%v", x, y, ok)
	}
	// a pantalla completa no hay cursor
	v.Markdown.Split = false
	if _, _, ok := ui.CursorCell(v, 120, 34); ok {
		t.Error("sin cursor en la vista previa completa")
	}
}

func TestMarkdownAreas(t *testing.T) {
	for _, tc := range []struct {
		name           string
		w              int
		explorer       bool
		md             *ui.MarkdownView
		px, ew, mx, mw int
	}{
		{"sin vista previa", 100, false, nil, 0, 100, 0, 0},
		{"con explorador", 100, true, nil, 25, 75, 0, 0},
		{"pantalla completa", 100, false, &ui.MarkdownView{}, 0, 0, 0, 100},
		{"completa con explorador", 100, true, &ui.MarkdownView{}, 25, 0, 25, 75},
		{"dividida", 101, false, &ui.MarkdownView{Split: true}, 0, 50, 51, 50},
		{"dividida con explorador", 125, true, &ui.MarkdownView{Split: true}, 25, 50, 76, 49},
		{"explorador oculto por estrecho", 79, true, &ui.MarkdownView{Split: true}, 0, 39, 40, 39},
	} {
		px, ew, mx, mw := ui.Areas(tc.w, tc.explorer, tc.md)
		if px != tc.px || ew != tc.ew || mx != tc.mx || mw != tc.mw {
			t.Errorf("%s: %d,%d,%d,%d; quería %d,%d,%d,%d", tc.name, px, ew, mx, mw, tc.px, tc.ew, tc.mx, tc.mw)
		}
		if tc.md != nil && tc.md.Split && ew+1+mw != tc.w-px {
			t.Errorf("%s: el reparto no cubre todo el ancho", tc.name)
		}
	}
	if ui.MarkdownContentWidth(200) != 80 || ui.MarkdownContentWidth(50) != 47 || ui.MarkdownContentWidth(5) != 10 {
		t.Error("MarkdownContentWidth")
	}
}

func TestMarkdownRowsNeverOverflow(t *testing.T) {
	src, _ := mdFixture(t, 40)
	for _, w := range []int{40, 70, 100, 140} {
		for _, split := range []bool{false, true} {
			pw := ui.PreviewWidth(w, false, true, split)
			doc := markdown.Render(src, ui.MarkdownContentWidth(pw))
			v := ui.EditorView{Buf: buffer.FromString("", src), Mode: "PREVIEW", Markdown: &ui.MarkdownView{Doc: &doc, Split: split}}
			for i, row := range strings.Split(ansi.Strip(ui.RenderEditor(v, w, 20, ui.Dark())), "\n") {
				if rw := ansi.StringWidth(row); rw != w {
					t.Errorf("w=%d split=%v fila %d mide %d", w, split, i, rw)
				}
			}
		}
	}
}

func TestMarkdownStillRenderingAndScroll(t *testing.T) {
	v := ui.EditorView{Buf: buffer.FromString("", "x\n"), Mode: "PREVIEW", Markdown: &ui.MarkdownView{}}
	if out := plain(v, 60, 10); !strings.Contains(out, "Renderizando…") {
		t.Errorf("mientras no hay documento:\n%s", out)
	}
	doc := markdown.Render(strings.Repeat("línea\n\n", 30), 40)
	v.Markdown = &ui.MarkdownView{Doc: &doc, Scroll: 4}
	first := strings.Split(plain(v, 60, 10), "\n")[0]
	if !strings.Contains(first, "línea") {
		t.Errorf("con scroll: %q", first)
	}
	v.Markdown.Scroll = 10000 // fuera de rango: no debe fallar
	_ = plain(v, 60, 10)
}

func TestMarkdownLightThemeColors(t *testing.T) {
	_, doc := mdFixture(t, 54)
	v := ui.EditorView{Buf: buffer.FromString("", "x\n"), Mode: "PREVIEW", Markdown: &ui.MarkdownView{Doc: &doc}}
	raw := ui.RenderEditor(v, 100, 30, ui.Light())
	if strings.Contains(raw, "48;2;10;14;24") || strings.Contains(raw, "48;2;23;27;39") {
		t.Error("el tema claro no usa fondos oscuros")
	}
	if !strings.Contains(raw, "48;2;227;230;239") {
		t.Error("fondo de código claro")
	}
}
