package editor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

// Reloj simulado para el doble clic.
type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }
func newClock() *clock                   { return &clock{t: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)} }

func (e *ed) click(x, y int) *ed {
	return e.send(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
}
func (e *ed) release(x, y int) *ed {
	return e.send(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft})
}
func (e *ed) drag(x, y int) *ed {
	return e.send(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft})
}
func (e *ed) wheel(x, y, notches int) *ed {
	b := tea.MouseWheelDown
	if notches < 0 {
		b, notches = tea.MouseWheelUp, -notches
	}
	for i := 0; i < notches; i++ {
		e.send(tea.MouseWheelMsg{X: x, Y: y, Button: b})
	}
	return e
}

// textX es la columna donde empieza el texto del editor sin explorador con ≤ 999 líneas.
const textX = 4

// mouseEd abre un editor de 100x30 sobre content (sin explorador), con el reloj simulado.
func mouseEd(t *testing.T, content string, clk *clock, opts ...editor.Option) *ed {
	t.Helper()
	if clk == nil {
		clk = newClock()
	}
	opts = append([]editor.Option{editor.WithClock(clk.now), editor.WithTimeout(20 * time.Millisecond)}, opts...)
	e := &ed{t: t, m: editor.New(buffer.FromString("", content), "", opts...)}
	return e.send(tea.WindowSizeMsg{Width: 100, Height: 30})
}

func nLines(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&sb, "línea %d\n", i)
	}
	return sb.String()
}

// ---------- cursor ----------

func TestClickMovesCursor(t *testing.T) {
	e := mouseEd(t, "uno\ndos tres\ncuatro cinco\n", nil)
	e.click(textX+5, 2)
	e.at(2, 5)
	e.click(textX+1, 0)
	e.at(0, 1)
	e.mode(editor.ModeNormal)
	if e.m.Buffer().Modified() {
		t.Error("un clic no modifica el texto")
	}
}

func TestClickOnGutterAndPastTheEndOfLine(t *testing.T) {
	clk := newClock()
	e := mouseEd(t, "uno\ndos tres\n", clk)
	e.click(1, 1) // sobre los números de línea
	e.at(1, 0)
	clk.advance(time.Second)
	e.click(80, 1) // más allá del final: en Normal, el último carácter
	e.at(1, 7)
	clk.advance(time.Second)
	e.typ("i").click(80, 1) // en Insert se puede llegar al final
	e.at(1, 8)
	e.mode(editor.ModeInsert)
}

func TestClickOnWideCharactersAndTabs(t *testing.T) {
	e := mouseEd(t, "日本語x\n\tabc\n", nil)
	for _, tc := range []struct{ x, line, col int }{
		{textX, 0, 0}, {textX + 1, 0, 0}, // las dos celdas del primer carácter ancho
		{textX + 2, 0, 1}, {textX + 3, 0, 1},
		{textX + 6, 0, 3},
		{textX, 1, 0}, {textX + 3, 1, 0}, // dentro del tabulador (4 celdas)
		{textX + 4, 1, 1}, {textX + 6, 1, 3},
	} {
		e.click(tc.x, tc.line)
		e.at(tc.line, tc.col)
	}
}

func TestClickBelowTheTextGoesToLastLine(t *testing.T) {
	e := mouseEd(t, "uno\ndos\ntres\n", nil)
	e.click(textX+2, 10) // una fila «~»
	e.at(2, 2)
	e.click(textX+1, 27)
	e.at(2, 1)
}

func TestClickWithScrollOffsets(t *testing.T) {
	e := mouseEd(t, nLines(200), nil).typ("100G")
	top := e.m.Top()
	if top == 0 {
		t.Fatal("preparación: hay scroll vertical")
	}
	e.click(textX+2, 5)
	e.at(top+5, 2)

	long := strings.Repeat("0123456789", 20)
	e = mouseEd(t, long+"\ncorta\n", nil).typ("$")
	left := e.m.Left()
	if left == 0 {
		t.Fatal("preparación: scroll horizontal")
	}
	e.click(textX+3, 0)
	e.at(0, left+3)
}

func TestClickResetsPendingKeysAndLeavesVisual(t *testing.T) {
	e := mouseEd(t, "uno dos\ntres cuatro\n", nil).typ("d")
	if e.m.Pending() != "d" {
		t.Fatal("preparación")
	}
	e.click(textX+2, 1)
	if e.m.Pending() != "" {
		t.Errorf("el clic cancela el operador pendiente: %q", e.m.Pending())
	}
	e.at(1, 2)
	e.typ("v").typ("l")
	e.mode(editor.ModeVisual)
	e.click(textX+1, 0)
	e.mode(editor.ModeNormal)
	e.at(0, 1)
}

func TestClickMovesFocusToEditor(t *testing.T) {
	root := project(t, "a.txt=hola\n")
	e := openProj(t, root).space("e")
	e.ex("e " + filepath.Join(root, "a.txt")) // abre el archivo; el explorador sigue visible
	e.space("e").space("e")
	if e.m.Focus() != editor.FocusExplorer {
		t.Fatal("preparación")
	}
	e.click(25+4+2, 0)
	if e.m.Focus() != editor.FocusEditor {
		t.Error("un clic en el texto da el foco al editor")
	}
	e.at(0, 2)
}

func TestMouseIgnoredWhilePromptIsOpen(t *testing.T) {
	e := mouseEd(t, "uno dos\n", nil).typ(":wq")
	e.click(textX+3, 0)
	e.mode(editor.ModeCommand)
	e.at(0, 0)
}

func TestRightAndMiddleButtonsAreIgnored(t *testing.T) {
	e := mouseEd(t, "uno dos\n", nil)
	e.send(tea.MouseClickMsg{X: textX + 3, Y: 0, Button: tea.MouseRight}, tea.MouseClickMsg{X: textX + 3, Y: 0, Button: tea.MouseMiddle})
	e.at(0, 0)
	e.send(tea.MouseReleaseMsg{X: 1, Y: 1, Button: tea.MouseLeft}) // soltar sin haber pulsado
	e.mode(editor.ModeNormal)
}

// ---------- selección ----------

func TestDragSelectsAndYankCopies(t *testing.T) {
	var copied []string
	clip := editor.WithClipboard(func(s string) tea.Cmd { copied = append(copied, s); return nil })
	e := mouseEd(t, "func (n *Node) Lookup(seq []rune) bool\n", nil, clip)
	// «Lookup» ocupa las columnas 15..20
	e.click(textX+15, 0)
	e.mode(editor.ModeNormal) // un solo clic no entra en Visual
	e.drag(textX+16, 0).drag(textX+18, 0).drag(textX+20, 0)
	e.mode(editor.ModeVisual)
	e.release(textX+20, 0)
	e.typ("y")
	if len(copied) != 1 || copied[0] != "Lookup" {
		t.Errorf("la palabra seleccionada se copia con y: %q", copied)
	}
	e.mode(editor.ModeNormal)
}

func TestDragBackwardsAndAcrossLines(t *testing.T) {
	var copied []string
	clip := editor.WithClipboard(func(s string) tea.Cmd { copied = append(copied, s); return nil })
	e := mouseEd(t, "uno dos\ntres cuatro\n", nil, clip)
	e.click(textX+6, 0).drag(textX+4, 0).release(textX+4, 0) // hacia atrás: «dos» sin la «o» final... de col 4 a 6
	e.typ("y")
	if copied[0] != "dos" {
		t.Errorf("hacia atrás: %q", copied[0])
	}
	e.click(textX+4, 0).drag(textX+3, 1).release(textX+3, 1)
	e.typ("y")
	if copied[1] != "dos\ntres" {
		t.Errorf("varias líneas: %q", copied[1])
	}
	// arrastrar y soltar en la misma celda no selecciona nada
	e.click(textX+1, 0).drag(textX+1, 0).release(textX+1, 0)
	e.mode(editor.ModeNormal)
}

func TestDragSelectionIsHighlighted(t *testing.T) {
	e := mouseEd(t, "uno dos tres\n", nil)
	e.click(textX+4, 0).drag(textX+6, 0)
	if !strings.Contains(e.m.View().Content, "48;2;50;61;92") {
		t.Error("la selección se ve resaltada")
	}
	if !strings.Contains(e.screen(), "3 car.") {
		t.Errorf("la statusline indica el tamaño:\n%s", e.screen())
	}
}

func TestDoubleClickSelectsWord(t *testing.T) {
	var copied []string
	clip := editor.WithClipboard(func(s string) tea.Cmd { copied = append(copied, s); return nil })
	clk := newClock()
	e := mouseEd(t, "func Lookup(seq []rune) bool\n", clk, clip)
	e.click(textX+7, 0)
	clk.advance(200 * time.Millisecond)
	e.click(textX+7, 0)
	e.mode(editor.ModeVisual)
	e.typ("y")
	if len(copied) != 1 || copied[0] != "Lookup" {
		t.Errorf("doble clic selecciona la palabra: %q", copied)
	}
	// puntuación y palabra son clases distintas (como w y b)
	clk.advance(time.Second)
	e.click(textX+11, 0)
	clk.advance(100 * time.Millisecond)
	e.click(textX+11, 0)
	e.typ("y")
	if copied[1] != "(" {
		t.Errorf("la puntuación se selecciona sola: %q", copied[1])
	}
}

func TestDoubleClickNeedsSameCellAndTime(t *testing.T) {
	clk := newClock()
	e := mouseEd(t, "uno dos tres\n", clk)
	e.click(textX+1, 0)
	clk.advance(DoubleClick + 50*time.Millisecond)
	e.click(textX+1, 0) // demasiado tarde
	e.mode(editor.ModeNormal)
	clk.advance(100 * time.Millisecond)
	e.click(textX+1, 0) // ya es doble respecto del anterior
	e.mode(editor.ModeVisual)

	e = mouseEd(t, "uno dos tres\n", clk)
	e.click(textX+1, 0)
	clk.advance(100 * time.Millisecond)
	e.click(textX+2, 0) // otra celda
	e.mode(editor.ModeNormal)
	// triple clic: el tercero empieza otra cuenta
	e = mouseEd(t, "uno dos tres\n", clk)
	e.click(textX+1, 0)
	clk.advance(50 * time.Millisecond)
	e.click(textX+1, 0)
	clk.advance(50 * time.Millisecond)
	e.mode(editor.ModeVisual)
	e.click(textX+1, 0)
	e.mode(editor.ModeNormal)
}

// DoubleClick es el intervalo de doble clic, repetido aquí para escribir las pruebas con claridad.
const DoubleClick = editor.DoubleClickInterval

func TestDoubleClickOnBlankAndEmptyLines(t *testing.T) {
	clk := newClock()
	e := mouseEd(t, "uno   dos\n\nfin\n", clk)
	e.click(textX+4, 0)
	clk.advance(10 * time.Millisecond)
	e.click(textX+4, 0)
	e.mode(editor.ModeVisual)
	e.send(keyEsc)
	clk.advance(time.Second)
	e.click(textX+1, 1)
	clk.advance(10 * time.Millisecond)
	e.click(textX+1, 1) // línea vacía: no falla
	e.send(keyEsc)
}

// ---------- rueda ----------

func TestWheelScrollsEditorByThreeLinesAndKeepsCursorVisible(t *testing.T) {
	e := mouseEd(t, nLines(300), nil)
	e.wheel(textX+5, 5, 1)
	if e.m.Top() != 3 {
		t.Errorf("una muesca baja 3 líneas: top=%d", e.m.Top())
	}
	if e.m.Cursor().Line < 3+3 { // el cursor se mantiene dentro de lo visible (margen de 3)
		t.Errorf("el cursor sigue visible: línea %d con top %d", e.m.Cursor().Line, e.m.Top())
	}
	e.wheel(textX+5, 5, 10)
	if e.m.Top() != 33 {
		t.Errorf("top=%d", e.m.Top())
	}
	if c := e.m.Cursor().Line; c < e.m.Top() || c >= e.m.Top()+28 {
		t.Errorf("el cursor queda dentro de la pantalla: %d (top %d)", c, e.m.Top())
	}
	e.wheel(textX+5, 5, -3)
	if e.m.Top() != 24 {
		t.Errorf("subir: top=%d", e.m.Top())
	}
	e.wheel(textX+5, 5, -100)
	if c := e.m.Cursor().Line; e.m.Top() != 0 || c > 24 { // sin scroll basta con que el cursor siga a la vista (hasta el margen inferior)
		t.Errorf("tope superior: top=%d cursor=%v", e.m.Top(), e.m.Cursor())
	}
	e.wheel(textX+5, 5, 1000)
	if want := 300 - 28; e.m.Top() != want {
		t.Errorf("tope inferior: top=%d, quería %d", e.m.Top(), want)
	}
	if e.m.Cursor().Line > 299 {
		t.Error("el cursor no sale del texto")
	}
	// después de la rueda, escribir sigue funcionando desde donde esté el cursor
	e.typ("x")
	if !e.m.Buffer().Modified() {
		t.Error("las teclas siguen funcionando")
	}
}

func TestWheelKeepsScrolledPositionAfterKeys(t *testing.T) {
	e := mouseEd(t, nLines(300), nil)
	e.wheel(textX+5, 5, 10)
	top := e.m.Top()
	e.typ("l") // una tecla que no mueve de línea no debe deshacer el desplazamiento
	if e.m.Top() != top {
		t.Errorf("la vista se quedó donde la dejó la rueda: %d → %d", top, e.m.Top())
	}
}

func TestWheelOnShortDocumentDoesNothing(t *testing.T) {
	e := mouseEd(t, "uno\ndos\n", nil)
	e.wheel(textX, 2, 5)
	if e.m.Top() != 0 {
		t.Errorf("top=%d", e.m.Top())
	}
	e.at(0, 0)
	e.send(tea.MouseWheelMsg{X: textX, Y: 2, Button: tea.MouseWheelLeft}, tea.MouseWheelMsg{X: textX, Y: 2, Button: tea.MouseWheelRight})
	e.at(0, 0)
}

func manyFiles(t *testing.T, n int) string {
	var files []string
	for i := 0; i < n; i++ {
		files = append(files, fmt.Sprintf("f%02d.txt=contenido %d\n", i, i))
	}
	return project(t, files...)
}

func TestWheelOverExplorerScrollsItWithoutChangingFocus(t *testing.T) {
	root := manyFiles(t, 60)
	e := openProj(t, root).space("e")
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}) // foco en el editor
	if e.m.Focus() != editor.FocusEditor {
		t.Fatal("preparación")
	}
	if e.panelHas("f40.txt") {
		t.Fatal("el árbol no cabe")
	}
	sel := e.m.ExplorerSelection()
	e.wheel(5, 5, 10) // 30 filas
	if !e.panelHas("f30.txt") {
		t.Errorf("el explorador se desplazó:\n%s", strings.Join(e.panel(), "\n"))
	}
	if e.panelHas("f00.txt") {
		t.Error("las primeras filas salen de la vista")
	}
	if e.m.Focus() != editor.FocusEditor {
		t.Error("el foco sigue en el editor")
	}
	if e.m.ExplorerSelection() != sel {
		t.Error("la selección no cambia")
	}
	e.typ("i") // una tecla del editor no devuelve el explorador a la selección
	if e.panelHas("f00.txt") {
		t.Error("el desplazamiento del explorador se respeta")
	}
	e.wheel(5, 5, -100)
	if !e.panelHas("f00.txt") {
		t.Error("arriba del todo")
	}
	e.wheel(5, 5, 1000)
	if !e.panelHas("f59.txt") {
		t.Errorf("abajo del todo:\n%s", strings.Join(e.panel(), "\n"))
	}
}

func TestWheelOverPreviewScrollsIt(t *testing.T) {
	e := mdEditor(t, "doc.md", longDoc())
	e.wheel(50, 5, 2)
	if e.m.PreviewScroll() != 6 {
		t.Errorf("vista previa completa: %d", e.m.PreviewScroll())
	}
	e.wheel(50, 5, -5)
	if e.m.PreviewScroll() != 0 {
		t.Errorf("arriba: %d", e.m.PreviewScroll())
	}
	e.wheel(50, 5, 1000)
	if e.m.PreviewScroll() != e.m.PreviewScroll() || !strings.Contains(e.screen(), "FINAL_DEL_DOCUMENTO") {
		t.Errorf("abajo del todo muestra el final:\n%s", e.screen())
	}
	// en el modo dividido, la rueda sobre la vista previa mueve el fuente y la vista previa lo sigue
	e.typ("gg").space("ms")
	e.wheel(75, 5, 20)
	if e.m.Top() == 0 || e.m.PreviewScroll() == 0 {
		t.Errorf("dividido: fuente top=%d, vista previa %d", e.m.Top(), e.m.PreviewScroll())
	}
}

// ---------- explorador ----------

func TestExplorerClicksOpenAndToggle(t *testing.T) {
	root := project(t, "docs/guia.md=# G\n", "src/main.go=package main\n", "src/util.go=package main\n", "LEEME.txt=hola\n")
	e := openProj(t, root).space("e")
	// filas: 0 raíz, 1 docs/, 2 src/, 3 LEEME.txt
	e.click(5, 2)
	if !e.panelHas("main.go") || !e.panelHas("util.go") {
		t.Fatalf("un clic en una carpeta la expande:\n%s", strings.Join(e.panel(), "\n"))
	}
	if e.m.Focus() != editor.FocusExplorer {
		t.Error("el foco queda en el explorador")
	}
	e.click(5, 2)
	if e.panelHas("main.go") {
		t.Error("otro clic la cierra")
	}
	e.click(5, 0) // la raíz no se cierra
	if !e.panelHas("src/") {
		t.Error("la raíz sigue abierta")
	}
	e.click(5, 2).click(8, 3) // main.go
	if filepath.Base(e.m.Buffer().Path()) != "main.go" || e.m.Focus() != editor.FocusEditor {
		t.Errorf("un clic en un archivo lo abre y pasa el foco al editor: %q", e.m.Buffer().Path())
	}
	e.wantText("package main\n")
	// un clic debajo de la última fila solo da el foco al explorador
	e.click(5, 20)
	if e.m.Focus() != editor.FocusExplorer {
		t.Error("foco")
	}
	if filepath.Base(e.m.Buffer().Path()) != "main.go" {
		t.Error("no abre nada")
	}
}

func TestExplorerClickWithScroll(t *testing.T) {
	root := manyFiles(t, 60)
	e := openProj(t, root).space("e")
	e.wheel(5, 5, 10) // baja 30 filas
	e.click(6, 2)     // la fila visible 2 es la 32 del árbol: f31.txt (la fila 0 es la raíz)
	if got := filepath.Base(e.m.Buffer().Path()); got != "f31.txt" {
		t.Errorf("con scroll el clic abre la fila correcta: %q", got)
	}
}

func TestExplorerClickIsIgnoredWhenHidden(t *testing.T) {
	root := project(t, "a.txt=x\n")
	e := openProj(t, root)
	e.click(5, 1)
	if e.m.ExplorerVisible() || e.m.Buffer().Path() != "" {
		t.Error("sin explorador visible no hay filas")
	}
}

// ---------- ventanas superpuestas ----------

func TestPickerClickSelectsAndDoubleClickAccepts(t *testing.T) {
	root := sampleProject(t)
	clk := newClock()
	e := openProj(t, root, editor.WithClock(clk.now)).findFiles().typ("trie")
	scr := strings.Split(e.screen(), "\n")
	var rowY int
	for y, r := range scr {
		if strings.Contains(r, "internal/keymap/trie_test.go") {
			rowY = y
		}
	}
	rowX := strings.Index(scr[rowY], "internal/keymap/trie_test.go")
	e.click(rowX+3, rowY) // un clic selecciona
	if !e.m.PickerOpen() {
		t.Fatal("un clic no cierra el picker")
	}
	if !strings.Contains(strings.Split(e.screen(), "\n")[rowY], "▸") {
		t.Errorf("la fila clicada queda seleccionada:\n%s", e.screen())
	}
	clk.advance(100 * time.Millisecond)
	e.click(rowX+3, rowY) // doble clic acepta
	if e.m.PickerOpen() || !strings.HasSuffix(filepath.ToSlash(e.m.Buffer().Path()), "internal/keymap/trie_test.go") {
		t.Errorf("el doble clic abre el archivo: %q", e.m.Buffer().Path())
	}
}

func TestPickerClickOutsideClosesWithoutRunning(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles().typ("trie")
	e.click(1, 28) // esquina inferior izquierda: fuera de la ventana
	if e.m.PickerOpen() {
		t.Error("un clic fuera cierra el picker")
	}
	if e.m.Buffer().Path() != "" || e.m.BufferCount() != 1 {
		t.Error("sin ejecutar nada")
	}
	// un clic dentro pero fuera de las filas no cierra
	e = openProj(t, sampleProject(t)).findFiles().typ("trie")
	e.click(45, 4) // sobre la fila de entrada
	if !e.m.PickerOpen() {
		t.Error("clic en la entrada")
	}
}

func TestPickerWheelMovesSelection(t *testing.T) {
	root := sampleProject(t)
	e := openProj(t, root).findFiles()
	res := e.results()
	e.wheel(50, 8, 1) // dentro de la ventana: baja 3 filas
	e.send(keyEnter)
	if got := filepath.ToSlash(e.m.Buffer().Path()); !strings.HasSuffix(got, res[3]) {
		t.Errorf("la rueda mueve la selección 3 filas: %q, quería %q", got, res[3])
	}
	// hacia arriba no pasa de la primera
	e = openProj(t, root).findFiles()
	e.wheel(50, 8, 2).wheel(50, 8, -9)
	e.send(keyEnter)
	if got := filepath.ToSlash(e.m.Buffer().Path()); !strings.HasSuffix(got, res[0]) {
		t.Errorf("tope superior: %q, quería %q", got, res[0])
	}
	// fuera de la ventana la rueda no actúa
	e = openProj(t, root).findFiles()
	e.wheel(1, 28, 5)
	e.send(keyEnter)
	if got := filepath.ToSlash(e.m.Buffer().Path()); !strings.HasSuffix(got, res[0]) {
		t.Errorf("fuera de la ventana: %q, quería %q", got, res[0])
	}
}

func TestPaletteClicks(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("abc\n"), 0o644)
	clk := newClock()
	m, _ := editor.Open(p, editor.WithClock(clk.now), editor.WithTimeout(20*time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 30}).typ("ix").send(keyEsc)
	e.space(" ").typ("guardar archivo")
	scr := strings.Split(e.screen(), "\n")
	var rowY int
	for y, r := range scr {
		if strings.Contains(r, "Guardar archivo") {
			rowY = y
		}
	}
	x := strings.Index(scr[rowY], "Guardar archivo")
	e.click(x+2, rowY)
	if got, _ := os.ReadFile(p); string(got) != "abc\n" {
		t.Error("un clic solo selecciona")
	}
	clk.advance(100 * time.Millisecond)
	e.click(x+2, rowY)
	if got, _ := os.ReadFile(p); string(got) != "xabc\n" {
		t.Errorf("el doble clic ejecuta el comando: %q", got)
	}
	if e.m.PickerOpen() {
		t.Error("la paleta se cierra")
	}
}

func TestSuggestionClicks(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("abc\n"), 0o644)
	clk := newClock()
	m, _ := editor.Open(p, editor.WithClock(clk.now), editor.WithTimeout(20*time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 30}).typ("ix").send(keyEsc)
	e.typ(":guardar archivo")
	scr := strings.Split(e.screen(), "\n")
	var rowY int
	for y, r := range scr {
		if strings.Contains(r, "Guardar archivo") && strings.Contains(r, "Ctrl+S") {
			rowY = y
		}
	}
	if rowY == 0 {
		t.Fatalf("sugerencias:\n%s", e.screen())
	}
	x := strings.Index(scr[rowY], "Guardar archivo")
	e.click(x+2, rowY).click(x+2, rowY) // el reloj no avanza: es un doble clic
	if got, _ := os.ReadFile(p); string(got) != "xabc\n" {
		t.Errorf("el doble clic en una sugerencia la ejecuta: %q", got)
	}
	if e.m.Mode() != editor.ModeNormal {
		t.Error("el prompt se cierra")
	}
}

func TestDialogButtonClicks(t *testing.T) {
	open := func() (*ed, string) {
		p := filepath.Join(t.TempDir(), "f.txt")
		os.WriteFile(p, []byte("a\n"), 0o644)
		m, _ := editor.Open(p, editor.WithTimeout(20*time.Millisecond))
		e := &ed{t: t, m: m}
		e.send(tea.WindowSizeMsg{Width: 100, Height: 30}).typ("ix").send(keyEsc).ex("q")
		return e, p
	}
	buttonAt := func(e *ed, label string) (int, int) {
		for y, row := range strings.Split(e.screen(), "\n") {
			if i := strings.Index(row, label); i >= 0 {
				return len([]rune(row[:i])) + 1, y
			}
		}
		t.Fatalf("no se ve el botón %q:\n%s", label, e.screen())
		return 0, 0
	}
	e, p := open()
	x, y := buttonAt(e, "Guardar (y)")
	e.click(x, y)
	if !e.quit {
		t.Error("Guardar sale")
	}
	if got, _ := os.ReadFile(p); string(got) != "xa\n" {
		t.Errorf("archivo = %q", got)
	}
	e, p = open()
	x, y = buttonAt(e, "Descartar (n)")
	e.click(x, y)
	if !e.quit {
		t.Error("Descartar sale")
	}
	if got, _ := os.ReadFile(p); string(got) != "a\n" {
		t.Errorf("descartar no escribe: %q", got)
	}
	e, _ = open()
	x, y = buttonAt(e, "Cancelar (Esc)")
	e.click(x, y)
	if e.quit || e.m.DialogOpen() {
		t.Error("Cancelar cierra el diálogo")
	}
	// un clic fuera de los botones no hace nada
	e, _ = open()
	e.click(1, 1)
	if !e.m.DialogOpen() || e.quit {
		t.Error("clic fuera del diálogo: se ignora")
	}
	e.click(textX, 0)
	if !e.m.DialogOpen() {
		t.Error("el diálogo es modal: el editor no recibe clics")
	}
	e.wheel(50, 5, 3)
	if !e.m.DialogOpen() {
		t.Error("la rueda tampoco actúa")
	}
}

// ---------- activar y desactivar ----------

func TestMouseToggle(t *testing.T) {
	e := mouseEd(t, "uno dos tres\n", nil)
	if !e.m.MouseEnabled() || e.m.View().MouseMode != tea.MouseModeCellMotion {
		t.Fatal("el ratón se captura por defecto")
	}
	e.ex("mouse")
	if e.m.MouseEnabled() || e.m.View().MouseMode != tea.MouseModeNone {
		t.Fatal("tras mouse.toggle no se captura el ratón")
	}
	if !strings.Contains(e.m.Message(), "nativa") {
		t.Errorf("el aviso explica que se usa la selección nativa: %q", e.m.Message())
	}
	e.click(textX+5, 0).wheel(textX, 0, 3).drag(textX+8, 0)
	e.at(0, 0)
	e.mode(editor.ModeNormal)
	e.ex("mouse")
	if !e.m.MouseEnabled() || e.m.View().MouseMode != tea.MouseModeCellMotion {
		t.Fatal("se vuelve a activar")
	}
	e.click(textX+5, 0)
	e.at(0, 5)
	// también desde la paleta
	e.space(" ").typ("desactivar el ratón").send(keyEnter)
	if e.m.MouseEnabled() {
		t.Error("la paleta también alterna el ratón")
	}
	if off := mouseEd(t, "x\n", nil, editor.WithMouse(false)); off.m.View().MouseMode != tea.MouseModeNone {
		t.Error("WithMouse(false)")
	}
}

func TestMouseEventsBeforeFirstSizeAreIgnored(t *testing.T) {
	m := editor.New(buffer.FromString("", "uno\n"), "")
	nm, _ := m.Update(tea.MouseClickMsg{X: 3, Y: 0, Button: tea.MouseLeft})
	nm, _ = nm.Update(tea.MouseWheelMsg{X: 3, Y: 0, Button: tea.MouseWheelDown})
	if nm.(editor.Model).Cursor() != (buffer.Pos{}) {
		t.Error("sin tamaño conocido no se actúa")
	}
}

func TestMouseCommandsAreRegistered(t *testing.T) {
	for _, name := range []string{"cursor.move_to", "visual.select_range", "view.scroll", "explorer.click", "mouse.toggle"} {
		c, ok := editor.Registry().Get(name)
		if !ok || c.Desc == "" || len(c.Keys) == 0 {
			t.Errorf("%s: %+v %v", name, c, ok)
		}
	}
}

func TestMouseInPreviewAndSplit(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	e.click(40, 5) // vista previa completa: no hay cursor que mover
	e.wantView("preview")
	e.typ("e").space("ms")
	e.click(textX+3, 4) // en el modo dividido, el clic en el fuente coloca el cursor
	if e.m.Cursor().Line != 4 {
		t.Errorf("%v", e.m.Cursor())
	}
	e.click(80, 4) // un clic en la vista previa no mueve el cursor del fuente
	if e.m.Cursor().Line != 4 {
		t.Errorf("%v", e.m.Cursor())
	}
}
