package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// fosforoTree reproduce el árbol de la maqueta 01.
func fosforoTree(icons ui.IconMode) *ui.ExplorerView {
	return &ui.ExplorerView{
		Rows: []ui.ExplorerRow{
			{Name: "fosforo", Root: true, IsDir: true, Expanded: true},
			{Name: "cmd", Depth: 1, IsDir: true, Expanded: true},
			{Name: "fosforo", Depth: 2, IsDir: true, Expanded: true},
			{Name: "main.go", Depth: 3},
			{Name: "internal", Depth: 1, IsDir: true, Expanded: true},
			{Name: "buffer", Depth: 2, IsDir: true},
			{Name: "editor", Depth: 2, IsDir: true},
			{Name: "keymap", Depth: 2, IsDir: true, Expanded: true},
			{Name: "trie.go", Depth: 3},
			{Name: "trie_test.go", Depth: 3},
			{Name: "ui", Depth: 2, IsDir: true},
			{Name: "docs", Depth: 1, IsDir: true, Expanded: true},
			{Name: "design", Depth: 2, IsDir: true},
			{Name: "README.md", Depth: 2},
			{Name: "openspec", Depth: 1, IsDir: true},
			{Name: ".gitignore", Depth: 1},
			{Name: "CONTRIBUTING.md", Depth: 1},
			{Name: "go.mod", Depth: 1},
			{Name: "LICENSE", Depth: 1},
			{Name: "Makefile", Depth: 1},
		},
		Selected: 8, Focused: true, Icons: icons,
	}
}

func explorerScreen(icons ui.IconMode) (string, string) {
	b := buffer.FromString("", trieGo)
	v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "EXPLORER", PathLabel: "internal/keymap/trie.go",
		Explorer: fosforoTree(icons)}
	raw := ui.RenderEditor(v, 100, 28, ui.Dark())
	return raw, plain(v, 100, 28)
}

func TestGoldenExplorerNerdIcons(t *testing.T) {
	_, out := explorerScreen(ui.IconsNerd)
	golden(t, "explorer_nerd", out)
}

func TestGoldenExplorerNoIcons(t *testing.T) {
	_, out := explorerScreen(ui.IconsNone)
	golden(t, "explorer_none", out)
	for _, want := range []string{"cmd/", "internal/", "keymap/", "buffer/", "trie.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q", want)
		}
	}
	for _, r := range out {
		if r >= 0xE000 && r <= 0xF8FF || r >= 0xF0000 {
			t.Fatalf("con icons=none no se imprime ningún icono: %U", r)
		}
	}
}

func TestExplorerColors(t *testing.T) {
	raw, _ := explorerScreen(ui.IconsNerd)
	rows := strings.Split(raw, "\n")
	// la raíz en negrita; las carpetas en azul; los archivos con el texto por defecto; trie.go (Go) con icono teal
	if !strings.Contains(rows[0], "\x1b[0;1;") || !strings.Contains(ansi.Strip(rows[0]), "fosforo") {
		t.Errorf("la raíz va en negrita: %q", rows[0])
	}
	const blue, teal, fg = "38;2;127;193;254", "38;2;74;209;211", "38;2;207;215;235"
	if !strings.Contains(rows[1], blue) {
		t.Errorf("una carpeta va en azul: %q", rows[1])
	}
	if sgr := lastSGRBefore(rows[3], "main.go"); strings.Contains(sgr, blue) || !strings.Contains(sgr, fg) {
		t.Errorf("el nombre de un archivo va con el color de texto, no en azul: %q", sgr)
	}
	if !strings.Contains(rows[3], teal) || !strings.Contains(rows[3], fg) {
		t.Errorf("el icono de Go va en teal y el nombre con el color de texto: %q", rows[3])
	}
	// fila seleccionada: fondo de selección; resto del panel: fondo del panel
	if !strings.Contains(rows[8], "48;2;35;44;66") {
		t.Errorf("la fila seleccionada se resalta: %q", rows[8])
	}
	if strings.Contains(rows[7], "48;2;35;44;66") && !strings.Contains(rows[7], "48;2;15;19;31") {
		t.Errorf("solo la fila seleccionada se resalta: %q", rows[7])
	}
}

func TestExplorerIconsNoneDirectoryColors(t *testing.T) {
	raw, _ := explorerScreen(ui.IconsNone)
	rows := strings.Split(raw, "\n")
	const blue, dim = "38;2;127;193;254", "38;2;117;128;156"
	if !strings.Contains(lastSGRBefore(rows[7], "keymap/"), blue) { // keymap/ expandida: azul
		t.Errorf("carpeta expandida en azul: %q", rows[7])
	}
	if !strings.Contains(lastSGRBefore(rows[5], "buffer/"), dim) { // buffer/ cerrada: atenuada
		t.Errorf("carpeta cerrada atenuada: %q", rows[5])
	}
}

func TestExplorerHiddenWhenTerminalIsNarrow(t *testing.T) {
	b := buffer.FromString("", "hola\n")
	v := ui.EditorView{Buf: b, Mode: "NORMAL", Explorer: fosforoTree(ui.IconsNone)}
	narrow := plain(v, 79, 12)
	if strings.Contains(narrow, "internal") || strings.Contains(narrow, "│") {
		t.Errorf("con menos de 80 columnas no hay panel:\n%s", narrow)
	}
	wide := plain(v, 80, 12)
	if !strings.Contains(wide, "internal") {
		t.Errorf("con 80 columnas sí:\n%s", wide)
	}
	if ui.PanelCols(80, true) != 25 || ui.PanelCols(79, true) != 0 || ui.PanelCols(200, false) != 0 {
		t.Error("PanelCols")
	}
}

func TestExplorerSeparatorIndicatesFocus(t *testing.T) {
	b := buffer.FromString("", "hola\n")
	for _, focused := range []bool{true, false} {
		ex := fosforoTree(ui.IconsNone)
		ex.Focused = focused
		raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "NORMAL", Explorer: ex}, 100, 12, ui.Dark())
		row := strings.Split(raw, "\n")[0]
		i := strings.Index(row, "│")
		if i < 0 {
			t.Fatal("falta el separador")
		}
		sgr := row[strings.LastIndex(row[:i], "\x1b["):i]
		isBlue := strings.Contains(sgr, "127;193;254")
		if isBlue != focused {
			t.Errorf("focused=%v: color del separador %q", focused, sgr)
		}
	}
}

func TestStatuslineShowsBufferPosition(t *testing.T) {
	b := buffer.FromString("", "x\n")
	row := func(idx, n int) string {
		rows := strings.Split(plain(ui.EditorView{Buf: b, Mode: "NORMAL", PathLabel: "a.go", BufIndex: idx, BufCount: n}, 80, 12), "\n")
		return rows[10]
	}
	if !strings.Contains(row(2, 3), "2/3") {
		t.Errorf("2/3: %q", row(2, 3))
	}
	if strings.Contains(row(1, 1), "1/1") || strings.Contains(row(0, 0), "0/0") {
		t.Error("con un solo buffer no se muestra")
	}
}

func TestExplorerScroll(t *testing.T) {
	for _, tc := range []struct{ scroll, sel, total, rows, want int }{
		{0, 0, 10, 5, 0},
		{0, 4, 10, 5, 0},
		{0, 5, 10, 5, 1},
		{3, 1, 10, 5, 1},
		{0, 9, 10, 5, 5},
		{9, 9, 10, 5, 5}, // no se pasa del final
		{0, 0, 3, 5, 0},  // cabe todo
	} {
		if got := ui.ExplorerScroll(tc.scroll, tc.sel, tc.total, tc.rows); got != tc.want {
			t.Errorf("%+v = %d", tc, got)
		}
	}
}

// lastSGRBefore devuelve la última secuencia de color que precede a la primera aparición de text en la fila.
func lastSGRBefore(row, text string) string {
	i := strings.Index(row, text)
	if i < 0 {
		return ""
	}
	pre := row[:i]
	return pre[strings.LastIndex(pre, "\x1b["):]
}
