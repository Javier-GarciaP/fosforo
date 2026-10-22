package ui_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

var update = flag.Bool("update", false, "reescribe los archivos golden")

// golden compara got con testdata/<name>.golden (texto sin colores, una fila por línea).
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falta %s (ejecuta go test ./internal/ui -update): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("la salida difiere de %s:\n--- obtenido ---\n%s\n--- esperado ---\n%s", path, got, want)
	}
}

// plain dibuja la vista y quita los colores, recortando los espacios finales de cada fila.
func plain(v ui.EditorView, w, h int) string {
	raw := ui.RenderEditor(v, w, h, ui.Dark())
	rows := strings.Split(ansi.Strip(raw), "\n")
	for i := range rows {
		rows[i] = strings.TrimRight(rows[i], " ")
	}
	return strings.Join(rows, "\n") + "\n"
}

const trieGo = `package keymap

type Node struct {
	children map[rune]*Node
	command  string
}

// Lookup resuelve una secuencia de teclas.
func (n *Node) Lookup(seq []rune) (string, bool) {
	cur := n
	for _, r := range seq {
		next, ok := cur.children[r]
		if !ok { return "", false }
		cur = next
	}
	return cur.command, cur.command != ""
}
`

func TestGoldenShortFile(t *testing.T) {
	// Maqueta 01: 17 líneas en una terminal de 36 filas; las filas siguientes muestran ~ y la penúltima es la statusline.
	b := buffer.FromString("", trieGo)
	v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "INSERT", PathLabel: "internal/keymap/trie.go"}
	out := plain(v, 100, 36)
	golden(t, "short_file", out)

	rows := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(rows) != 36 {
		t.Fatalf("filas = %d, quería 36", len(rows))
	}
	if !strings.HasPrefix(rows[0], "  1 package keymap") || !strings.HasPrefix(rows[16], " 17 }") {
		t.Errorf("las 17 líneas deben aparecer numeradas:\n%s", strings.Join(rows[:18], "\n"))
	}
	for _, r := range rows[17:34] {
		if strings.TrimSpace(r) != "~" {
			t.Errorf("las filas sobrantes muestran ~, y se obtuvo %q", r)
		}
	}
	if !strings.Contains(rows[34], "INSERT") || !strings.Contains(rows[34], "12:35") {
		t.Errorf("la penúltima fila es la statusline con el modo y la posición: %q", rows[34])
	}
}

func TestStatuslineContent(t *testing.T) {
	b := buffer.FromString("", strings.Repeat("x\n", 22))
	base := ui.EditorView{Buf: b, Mode: "NORMAL", PathLabel: "a/b.go"}
	statusRow := func(v ui.EditorView) string {
		rows := strings.Split(plain(v, 80, 12), "\n")
		return rows[10]
	}

	v := base
	v.Cursor = buffer.Pos{Line: 11, Col: 34}
	if got := statusRow(v); !strings.Contains(got, "12:35") || !strings.Contains(got, "54%") || !strings.Contains(got, "NORMAL") || !strings.Contains(got, "a/b.go") {
		t.Errorf("posición y porcentaje: %q", got)
	}
	if got := statusRow(base); strings.Contains(got, "[+]") || strings.Contains(got, "solo lectura") {
		t.Errorf("sin cambios no hay marcas: %q", got)
	}

	b.Insert(buffer.Pos{}, "!")
	if got := statusRow(base); !strings.Contains(got, "[+]") {
		t.Errorf("modificado debe mostrar [+]: %q", got)
	}

	ro := ui.EditorView{Buf: buffer.FromString("", "x\n"), Mode: "INSERT", PathLabel: "f"}
	ro.Buf, _ = openReadOnly(t)
	if got := statusRow(ui.EditorView{Buf: ro.Buf, Mode: "INSERT", PathLabel: "bin"}); !strings.Contains(got, "solo lectura") {
		t.Errorf("solo lectura: %q", got)
	}

	if got := statusRow(ui.EditorView{Buf: buffer.New(""), Mode: "INSERT"}); !strings.Contains(got, "[Sin nombre]") {
		t.Errorf("sin nombre: %q", got)
	}
	crlf := buffer.FromString("", "a\r\nb\r\n")
	if got := statusRow(ui.EditorView{Buf: crlf, Mode: "INSERT", PathLabel: "w"}); !strings.Contains(got, "CRLF") {
		t.Errorf("CRLF: %q", got)
	}
}

func openReadOnly(t *testing.T) (*buffer.Buffer, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bin")
	if err := os.WriteFile(p, []byte("a\x00b"), 0o644); err != nil {
		t.Fatal(err)
	}
	return buffer.Open(p)
}

func TestCommandLine(t *testing.T) {
	b := buffer.FromString("", "x\n")
	last := func(v ui.EditorView) string {
		rows := strings.Split(strings.TrimSuffix(plain(v, 80, 12), "\n"), "\n")
		return rows[len(rows)-1]
	}
	if got := last(ui.EditorView{Buf: b, Mode: "INSERT"}); got != "-- INSERT --" {
		t.Errorf("sin mensaje: %q", got)
	}
	if got := last(ui.EditorView{Buf: b, Mode: "INSERT", Message: "E37: error", IsError: true}); got != "E37: error" {
		t.Errorf("con mensaje: %q", got)
	}
}

func TestGoldenHorizontalScrollTabsAndWide(t *testing.T) {
	b := buffer.FromString("", "\tinicio\n日本語のテキストが長い行です\nabc\tdef\n")
	v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 1, Col: 3}, Left: 3, Mode: "INSERT", PathLabel: "x"}
	golden(t, "scroll_wide", plain(v, 24, 8))
}

func TestWideCharCutAtBorderBecomesSpace(t *testing.T) {
	b := buffer.FromString("", "日本語\n")
	// left = 1 corta el primer carácter ancho por la mitad
	out := plain(ui.EditorView{Buf: b, Left: 1, Mode: "INSERT"}, 20, 6)
	row := strings.Split(out, "\n")[0]
	if !strings.Contains(row, "本語") || strings.Contains(row, "日") {
		t.Errorf("el carácter cortado no se dibuja: %q", row)
	}
	if strings.Contains(row, "\uFFFD") {
		t.Errorf("no debe aparecer el carácter de reemplazo: %q", row)
	}
}

func TestControlCharactersAreNeverSentToTerminal(t *testing.T) {
	b := buffer.FromString("", "a\x1b[31mROJO\x07\x00b\u009b\n")
	raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "INSERT"}, 40, 10, ui.Dark())
	stripped := ansi.Strip(raw)
	if strings.ContainsAny(stripped, "\x1b\x07\x00\u009b") {
		t.Errorf("la salida contiene caracteres de control del archivo: %q", stripped)
	}
	// las únicas secuencias de escape son las del propio dibujo (SGR de color)
	for _, seq := range strings.Split(raw, "\x1b")[1:] {
		if !strings.HasPrefix(seq, "[") || !strings.Contains(seq[:min(len(seq), 40)], "m") {
			t.Fatalf("secuencia inesperada: %q", seq[:min(len(seq), 20)])
		}
	}
	if !strings.Contains(stripped, "a·[31mROJO·") {
		t.Errorf("los controles se muestran como punto: %q", strings.Split(stripped, "\n")[0])
	}
}

func TestInvalidUTF8IsSanitized(t *testing.T) {
	b := buffer.FromString("", "ab\xffcd\n")
	raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "INSERT"}, 40, 10, ui.Dark())
	if strings.Contains(raw, "\xff") {
		t.Error("los bytes inválidos no deben llegar al terminal")
	}
}

func TestCurrentLineHighlighted(t *testing.T) {
	b := buffer.FromString("", "uno\ndos\n")
	raw := ui.RenderEditor(ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 1}, Mode: "INSERT"}, 40, 10, ui.Dark())
	rows := strings.Split(raw, "\n")
	if !strings.Contains(rows[1], "48;2;35;44;66") { // fondo «sel» (#232c42)
		t.Errorf("la línea del cursor debe tener fondo de selección: %q", rows[1])
	}
	if strings.Contains(rows[0], "48;2;35;44;66") {
		t.Error("solo la línea del cursor lleva fondo de selección")
	}
}

func TestEveryRowHasExactWidth(t *testing.T) {
	b := buffer.FromString("", "日本語 mixto 🙂 con\ttab\n"+strings.Repeat("x", 200)+"\n")
	for _, w := range []int{40, 61, 100} {
		for _, left := range []int{0, 1, 2, 5} {
			out := ansi.Strip(ui.RenderEditor(ui.EditorView{Buf: b, Left: left, Mode: "INSERT"}, w, 12, ui.Dark()))
			for i, row := range strings.Split(out, "\n") {
				if got := ansi.StringWidth(row); got != w {
					t.Errorf("w=%d left=%d fila %d mide %d celdas: %q", w, left, i, got, row)
				}
			}
		}
	}
}

func TestCursorCell(t *testing.T) {
	b := buffer.FromString("", "日本x\n")
	x, y, ok := ui.CursorCell(ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 0, Col: 2}}, 80, 24)
	if !ok || x != ui.GutterWidth(1)+4 || y != 0 {
		t.Errorf("cursor en (%d,%d,%v)", x, y, ok)
	}
	if _, _, ok := ui.CursorCell(ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 0, Col: 2}, Dialog: &ui.Dialog{}}, 80, 24); ok {
		t.Error("con diálogo no hay cursor")
	}
	if _, _, ok := ui.CursorCell(ui.EditorView{Buf: b, Top: 1}, 80, 24); ok {
		t.Error("cursor fuera de la vista")
	}
}

func dialog() *ui.Dialog {
	return &ui.Dialog{
		Title:    "Cambios sin guardar",
		Lines:    []string{`"internal/keymap/trie.go" tiene cambios sin guardar.`, "¿Qué quieres hacer?"},
		Buttons:  []string{"Guardar (y)", "Descartar (n)", "Cancelar (Esc)"},
		Selected: 0,
		Hint:     "←→ elegir   ⏎ confirmar",
	}
}

func TestGoldenUnsavedDialog(t *testing.T) {
	// Maqueta 10: diálogo centrado sobre el editor.
	b := buffer.FromString("", trieGo)
	b.Insert(buffer.Pos{}, "")
	v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "INSERT", PathLabel: "internal/keymap/trie.go", Dialog: dialog()}
	golden(t, "dialog_wide", plain(v, 100, 36))
}

func TestGoldenUnsavedDialogNarrow(t *testing.T) {
	b := buffer.FromString("", trieGo)
	v := ui.EditorView{Buf: b, Mode: "INSERT", PathLabel: "trie.go", Dialog: dialog()}
	out := plain(v, 40, 16)
	golden(t, "dialog_narrow", out)
	for _, want := range []string{"Guardar (y)", "Descartar (n)", "Cancelar (Esc)"} {
		if !strings.Contains(out, want) {
			t.Errorf("en una terminal estrecha deben verse todos los botones; falta %q", want)
		}
	}
}

func TestDialogSelectedButtonIsHighlighted(t *testing.T) {
	b := buffer.FromString("", "x\n")
	for sel := 0; sel < 3; sel++ {
		d := dialog()
		d.Selected = sel
		raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "INSERT", Dialog: d}, 100, 24, ui.Dark())
		if got := strings.Count(raw, "48;2;127;193;254"); got == 0 { // fondo azul
			t.Errorf("seleccionado %d: no hay botón resaltado", sel)
		}
	}
}

func TestGoldenEmptyBuffer(t *testing.T) {
	golden(t, "empty", plain(ui.EditorView{Buf: buffer.New(""), Mode: "INSERT"}, 50, 12))
}

func TestStatuslineLongPathKeepsMarksVisible(t *testing.T) {
	b := buffer.FromString("", "x\n")
	b.Insert(buffer.Pos{}, "!")
	long := "/tmp/TestOpenEditAndSaveRealFile3102208925/001/un/directorio/muy/profundo/notas.txt"
	for _, w := range []int{40, 60, 80} {
		out := plain(ui.EditorView{Buf: b, Mode: "INSERT", PathLabel: long}, w, 12)
		row := strings.Split(out, "\n")[10]
		if !strings.Contains(row, "[+]") || !strings.Contains(row, "…") || !strings.Contains(row, "notas.txt") {
			t.Errorf("w=%d: la statusline debe recortar la ruta por la izquierda y conservar [+]: %q", w, row)
		}
		if ansi.StringWidth(row) > w {
			t.Errorf("w=%d: la fila se desborda: %q", w, row)
		}
		if !strings.Contains(row, "1:1") {
			t.Errorf("w=%d: falta la posición: %q", w, row)
		}
	}
}

func TestGoldenVisualSelection(t *testing.T) {
	// Maqueta 03: selección por líneas de 3 líneas; la statusline muestra su tamaño.
	b := buffer.FromString("", trieGo)
	sel := &ui.Selection{Start: buffer.Pos{Line: 9}, End: buffer.Pos{Line: 11}, Linewise: true}
	v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "VISUAL", PathLabel: "internal/keymap/trie.go",
		Info: "3 líneas", Selection: sel}
	golden(t, "visual", plain(v, 100, 20))

	raw := ui.RenderEditor(v, 100, 20, ui.Dark())
	rows := strings.Split(raw, "\n")
	const visualBG = "48;2;50;61;92" // #323d5c
	for i, want := range map[int]bool{8: false, 9: true, 10: true, 11: true, 12: false} {
		if got := strings.Contains(rows[i], visualBG); got != want {
			t.Errorf("fila %d: fondo de selección = %v, quería %v", i, got, want)
		}
	}
}

func TestVisualCharSelectionHighlightsOnlySelectedCells(t *testing.T) {
	b := buffer.FromString("", "abcdef\nghijkl\n")
	sel := &ui.Selection{Start: buffer.Pos{Line: 0, Col: 2}, End: buffer.Pos{Line: 1, Col: 1}}
	raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "VISUAL", Selection: sel}, 40, 10, ui.Dark())
	rows := strings.Split(raw, "\n")
	// fila 0: «ab» sin selección, «cdef» y el salto de línea con selección
	if !strings.Contains(rows[0], "ab") {
		t.Fatal(rows[0])
	}
	idxAB := strings.Index(rows[0], "ab")
	idxCD := strings.Index(rows[0], "cdef")
	if strings.Contains(rows[0][:idxAB], "48;2;50;61;92") {
		t.Error("antes de la selección no debe haber fondo de selección")
	}
	if !strings.Contains(rows[0][idxAB:idxCD+1], "48;2;50;61;92") {
		t.Error("«cdef» debe estar seleccionado")
	}
	// fila 1: «gh» seleccionados (cols 0..1), «ijkl» no
	plainRow := ansi.Strip(rows[1])
	if !strings.Contains(plainRow, "ghijkl") {
		t.Fatal(plainRow)
	}
	idxGH := strings.Index(rows[1], "gh")
	idxIJ := strings.Index(rows[1], "ijkl")
	if !strings.Contains(rows[1][idxGH-30:idxIJ], "48;2;50;61;92") {
		t.Error("«gh» debe estar seleccionado")
	}
	if strings.Contains(rows[1][idxIJ:], "48;2;50;61;92") {
		t.Error("«ijkl» no debe estar seleccionado")
	}
}

func TestSelectionOnEmptyLineShowsOneCell(t *testing.T) {
	b := buffer.FromString("", "a\n\nb\n")
	sel := &ui.Selection{Start: buffer.Pos{Line: 0}, End: buffer.Pos{Line: 2}, Linewise: true}
	raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: "V-LINE", Selection: sel}, 40, 10, ui.Dark())
	rows := strings.Split(raw, "\n")
	if !strings.Contains(rows[1], "48;2;50;61;92") {
		t.Errorf("una línea vacía seleccionada se marca con una celda: %q", rows[1])
	}
}

func TestPromptAndPendingInCommandLine(t *testing.T) {
	b := buffer.FromString("", "x\n")
	last := func(v ui.EditorView) string {
		rows := strings.Split(strings.TrimSuffix(plain(v, 60, 12), "\n"), "\n")
		return rows[len(rows)-1]
	}
	if got := last(ui.EditorView{Buf: b, Mode: "COMMAND", Prompt: &ui.Prompt{Prefix: ":", Text: "wq", Cursor: 2}}); got != ":wq" {
		t.Errorf("prompt: %q", got)
	}
	if got := last(ui.EditorView{Buf: b, Mode: "NORMAL", Pending: "3d"}); got != "3d" {
		t.Errorf("pendiente: %q", got)
	}
	if got := last(ui.EditorView{Buf: b, Mode: "NORMAL"}); got != "" {
		t.Errorf("Normal no muestra banner: %q", got)
	}
	if got := last(ui.EditorView{Buf: b, Mode: "VISUAL"}); got != "-- VISUAL --" {
		t.Errorf("Visual: %q", got)
	}
	x, y, ok := ui.CursorCell(ui.EditorView{Buf: b, Prompt: &ui.Prompt{Prefix: ":", Text: "日本", Cursor: 1}}, 60, 12)
	if !ok || x != 3 || y != 11 {
		t.Errorf("cursor del prompt en (%d,%d,%v), quería (3,11)", x, y, ok)
	}
}

func TestModeColors(t *testing.T) {
	b := buffer.FromString("", "x\n")
	for mode, want := range map[string]string{
		"NORMAL":  "48;2;127;193;254", // azul
		"INSERT":  "48;2;127;208;150", // verde
		"VISUAL":  "48;2;204;168;246", // malva
		"V-LINE":  "48;2;204;168;246",
		"COMMAND": "48;2;245;165;116", // melocotón
	} {
		raw := ui.RenderEditor(ui.EditorView{Buf: b, Mode: mode}, 60, 12, ui.Dark())
		status := strings.Split(raw, "\n")[10]
		if !strings.Contains(status, want) {
			t.Errorf("%s: la etiqueta debe usar su color (%s): %q", mode, want, status[:60])
		}
	}
}

// Con Start y sitio suficiente, RenderEditor dibuja el logo y la lista y no los «~» del editor vacío.
func TestRenderEditorStartScreen(t *testing.T) {
	v := ui.EditorView{Buf: buffer.FromString("", ""), Start: true}
	out := plain(v, 100, 30)
	if !strings.Contains(out, "██╗") || !strings.Contains(out, "Space f f") {
		t.Errorf("falta el logo o la lista:\n%s", out)
	}
	if strings.Contains(out, "~") {
		t.Error("con la pantalla de inicio no debe haber «~»")
	}
	v.Start = false
	if strings.Contains(plain(v, 100, 30), "Space f f") {
		t.Error("sin Start no debe verse la pantalla de inicio")
	}
}

// Si la lista no cabe (terminal muy baja o estrecha) con Start se dibuja el editor vacío normal.
func TestRenderEditorStartFallsBackWhenSmall(t *testing.T) {
	v := ui.EditorView{Buf: buffer.FromString("", ""), Start: true}
	small := plain(v, 100, 9) // 7 filas de texto: menos que las 8 acciones
	if strings.Contains(small, "Space f f") || !strings.Contains(small, "~") {
		t.Errorf("con poco alto debe verse el editor vacío normal:\n%s", small)
	}
}

// El logo usa el color de la marca (Brand) en el tema oscuro y en el claro.
func TestRenderEditorStartLogoUsesBrand(t *testing.T) {
	v := ui.EditorView{Buf: buffer.FromString("", ""), Start: true}
	for name, th := range map[string]ui.Theme{"oscuro": ui.Dark(), "claro": ui.Light()} {
		raw := ui.RenderEditor(v, 100, 30, th)
		r, g, b := hexToRGB(t, th.Brand)
		if !strings.Contains(raw, fmt.Sprintf("38;2;%d;%d;%d", r, g, b)) {
			t.Errorf("tema %s: no se usa el color Brand %s", name, th.Brand)
		}
	}
}

func hexToRGB(t *testing.T, h string) (r, g, b int) {
	t.Helper()
	if _, err := fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatal(err)
	}
	return
}

// La pantalla de inicio no tiene cursor; si no cabe, el cursor vuelve a su sitio normal.
func TestCursorCellStartScreen(t *testing.T) {
	v := ui.EditorView{Buf: buffer.FromString("", ""), Start: true}
	if _, _, ok := ui.CursorCell(v, 100, 30); ok {
		t.Error("con la pantalla de inicio no hay cursor")
	}
	if _, _, ok := ui.CursorCell(v, 100, 9); !ok {
		t.Error("si la pantalla de inicio no cabe, el cursor se muestra")
	}
}
