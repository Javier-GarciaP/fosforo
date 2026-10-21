package editor_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

const wait = 3 * time.Second

var (
	keyEnter     = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyBackspace = tea.KeyPressMsg{Code: tea.KeyBackspace}
	keyDelete    = tea.KeyPressMsg{Code: tea.KeyDelete}
	keyLeft      = tea.KeyPressMsg{Code: tea.KeyLeft}
	keyRight     = tea.KeyPressMsg{Code: tea.KeyRight}
	keyUp        = tea.KeyPressMsg{Code: tea.KeyUp}
	keyDown      = tea.KeyPressMsg{Code: tea.KeyDown}
	keyHome      = tea.KeyPressMsg{Code: tea.KeyHome}
	keyEnd       = tea.KeyPressMsg{Code: tea.KeyEnd}
	keyPgUp      = tea.KeyPressMsg{Code: tea.KeyPgUp}
	keyPgDown    = tea.KeyPressMsg{Code: tea.KeyPgDown}
	keyEsc       = tea.KeyPressMsg{Code: tea.KeyEscape}
	keyTab       = tea.KeyPressMsg{Code: tea.KeyTab}
	keyCtrlS     = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	keyCtrlQ     = tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}
	keyCtrlC     = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
)

// ed envuelve un modelo para encadenar mensajes en las pruebas sin ejecutar el bucle de Bubble Tea.
type ed struct {
	t    *testing.T
	m    editor.Model
	quit bool
}

func newEd(t *testing.T, content string) *ed {
	t.Helper()
	e := &ed{t: t, m: editor.New(buffer.FromString("", content), "")}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24})
	return e
}

func (e *ed) send(msgs ...tea.Msg) *ed {
	for _, msg := range msgs {
		nm, cmd := e.m.Update(msg)
		e.m = nm.(editor.Model)
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				e.quit = true
			}
		}
	}
	return e
}

func (e *ed) typ(s string) *ed {
	for _, r := range s {
		e.send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return e
}

func (e *ed) text() string { return e.m.Buffer().Text() }

func (e *ed) at(line, col int) {
	e.t.Helper()
	if got := e.m.Cursor(); got != (buffer.Pos{Line: line, Col: col}) {
		e.t.Errorf("cursor = %+v, quería {%d %d}", got, line, col)
	}
}

func TestTypeAndNewline(t *testing.T) {
	e := newEd(t, "").typ("ab").send(keyEnter).typ("c")
	if e.m.Buffer().LineCount() != 2 || e.m.Buffer().Line(0) != "ab" || e.m.Buffer().Line(1) != "c" {
		t.Fatalf("líneas: %q %q", e.m.Buffer().Line(0), e.m.Buffer().Line(1))
	}
	e.at(1, 1)
	if !e.m.Buffer().Modified() {
		t.Error("debe quedar modificado")
	}
}

func TestEnterKeepsIndentation(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		col           int
		want1, want2  string
		cursorCol     int
	}{
		{"espacios", "    foo", 7, "    foo", "    ", 4},
		{"tabuladores", "\t\tfoo", 5, "\t\tfoo", "\t\t", 2},
		{"cursor dentro del sangrado", "      foo", 2, "  ", "      foo", 0},
		{"sin sangrado", "foo", 3, "foo", "", 0},
		{"parte media", "  abcd", 4, "  ab", "  cd", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content)
			for i := 0; i < tc.col; i++ {
				e.send(keyRight)
			}
			e.send(keyEnter)
			if got := e.m.Buffer().Line(0); got != tc.want1 {
				t.Errorf("línea 1 = %q, quería %q", got, tc.want1)
			}
			if got := e.m.Buffer().Line(1); got != tc.want2 {
				t.Errorf("línea 2 = %q, quería %q", got, tc.want2)
			}
			if tc.name == "cursor dentro del sangrado" {
				e.at(1, 2) // el sangrado insertado se limita a lo que hay antes del cursor
			}
		})
	}
}

func TestBackspaceJoinsLines(t *testing.T) {
	e := newEd(t, "ab\ncd\n").send(keyDown, keyHome, keyBackspace)
	if e.m.Buffer().LineCount() != 1 || e.m.Buffer().Line(0) != "abcd" {
		t.Fatalf("buffer: %q", e.text())
	}
	e.at(0, 2) // el cursor queda en el punto de unión
}

func TestDeleteJoinsLines(t *testing.T) {
	e := newEd(t, "ab\ncd\n").send(keyEnd, keyDelete)
	if e.m.Buffer().Line(0) != "abcd" || e.m.Buffer().LineCount() != 1 {
		t.Fatalf("buffer: %q", e.text())
	}
	e.at(0, 2)
	e.send(keyDelete)
	if e.m.Buffer().Line(0) != "abd" {
		t.Errorf("Delete dentro de la línea: %q", e.m.Buffer().Line(0))
	}
}

func TestBackspaceDeleteAtLimitsDoNothing(t *testing.T) {
	e := newEd(t, "ab\n")
	e.send(keyBackspace)
	if e.m.Buffer().Modified() {
		t.Error("Backspace al inicio del texto no debe modificar")
	}
	e.send(keyDown, keyEnd, keyDelete) // fin del texto
	if e.m.Buffer().Modified() {
		t.Error("Delete al final del texto no debe modificar")
	}
}

func TestBackspaceRemovesWholeEmoji(t *testing.T) {
	e := newEd(t, "").typ("a👨‍👩‍👧b")
	e.send(keyBackspace).send(keyBackspace)
	if got := e.m.Buffer().Line(0); got != "a" {
		t.Errorf("tras dos Backspace: %q, quería %q", got, "a")
	}
}

func TestTabInsertsTab(t *testing.T) {
	e := newEd(t, "").send(keyTab).typ("x")
	if got := e.m.Buffer().Line(0); got != "\tx" {
		t.Errorf("línea = %q", got)
	}
}

func TestPasteInsertsText(t *testing.T) {
	e := newEd(t, "").send(tea.PasteMsg{Content: "uno\r\ndos\ntres"})
	if e.m.Buffer().LineCount() != 3 || e.m.Buffer().Line(2) != "tres" {
		t.Errorf("buffer: %q", e.text())
	}
	e.at(2, 4)
}

func TestControlKeysDoNotInsertText(t *testing.T) {
	e := newEd(t, "").send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: 'x', Mod: tea.ModAlt}, keyEsc)
	if e.m.Buffer().Modified() {
		t.Errorf("las teclas con Ctrl/Alt no deben escribir: %q", e.text())
	}
}

func TestArrowsHomeEnd(t *testing.T) {
	e := newEd(t, "abc\n日本\n")
	e.send(keyRight, keyRight)
	e.at(0, 2)
	e.send(keyEnd)
	e.at(0, 3)
	e.send(keyRight) // cruza a la línea siguiente
	e.at(1, 0)
	e.send(keyLeft)
	e.at(0, 3)
	e.send(keyHome)
	e.at(0, 0)
	e.send(keyLeft, keyUp) // límites: no pasa nada
	e.at(0, 0)
}

func TestVerticalKeepsDesiredColumn(t *testing.T) {
	e := newEd(t, "01234567890123456789012\nabcde\n01234567890123456789012\n")
	for i := 0; i < 20; i++ {
		e.send(keyRight)
	}
	e.send(keyDown)
	e.at(1, 5)
	e.send(keyDown)
	e.at(2, 20)
}

func TestPageKeysClampToLimits(t *testing.T) {
	e := newEd(t, strings.Repeat("x\n", 100))
	e.send(keyPgDown)
	if got := e.m.Cursor().Line; got != 21 { // 24 filas - 2 reservadas - 1 de solapamiento
		t.Errorf("PageDown mueve %d líneas, quería 21", got)
	}
	for i := 0; i < 20; i++ {
		e.send(keyPgDown)
	}
	if got := e.m.Cursor().Line; got != 99 {
		t.Errorf("PageDown se detiene en la última línea: %d", got)
	}
	for i := 0; i < 20; i++ {
		e.send(keyPgUp)
	}
	e.at(0, 0)
}

func TestScrollKeepsMargin(t *testing.T) {
	e := newEd(t, strings.Repeat("línea\n", 100)) // 22 filas de texto: líneas 0..21 visibles
	for i := 0; i < 18; i++ {
		e.send(keyDown)
	}
	if e.m.Top() != 0 {
		t.Fatalf("en la fila 18 aún no debe haber scroll (top=%d)", e.m.Top())
	}
	e.send(keyDown) // línea 19: a 3 filas del final visible (fila 21)
	if e.m.Top() != 1 {
		t.Fatalf("top = %d, quería 1 para conservar el margen de 3", e.m.Top())
	}
	for i := 0; i < 30; i++ {
		e.send(keyDown)
	}
	if got, want := e.m.Cursor().Line-e.m.Top(), 22-1-3; got != want {
		t.Errorf("fila del cursor = %d, quería %d (margen 3 abajo)", got, want)
	}
	// subiendo, el margen superior también se respeta
	for i := 0; i < 40; i++ {
		e.send(keyUp)
	}
	if got := e.m.Cursor().Line - e.m.Top(); got != 3 {
		t.Errorf("fila del cursor al subir = %d, quería 3", got)
	}
}

func TestScrollStopsAtEndOfFile(t *testing.T) {
	e := newEd(t, strings.Repeat("x\n", 30)) // 30 líneas, 22 filas
	e.send(keyPgDown, keyPgDown, keyPgDown)
	if e.m.Top() != 30-22 {
		t.Errorf("top = %d: la última línea debe quedar abajo sin dejar filas vacías", e.m.Top())
	}
}

func TestHorizontalScroll(t *testing.T) {
	long := strings.Repeat("0123456789", 20) // 200 columnas
	e := newEd(t, long+"\ncorta\n")
	e.send(keyEnd)
	// ancho de texto = 80 - 4 de gutter = 76
	if e.m.Left() != 200-76+1 {
		t.Errorf("left = %d, quería %d (cursor al final de la línea)", e.m.Left(), 200-76+1)
	}
	e.send(keyDown)
	if e.m.Left() != 0 {
		t.Errorf("en una línea corta la vista vuelve a la izquierda: left = %d", e.m.Left())
	}
	e.send(keyUp, keyHome)
	if e.m.Left() != 0 {
		t.Errorf("Home vuelve al inicio: left = %d", e.m.Left())
	}
}

func TestResizeKeepsCursorVisible(t *testing.T) {
	e := newEd(t, strings.Repeat("a", 70)+"\n")
	e.send(keyEnd)
	if e.m.Left() != 0 {
		t.Fatalf("con 80 columnas la línea cabe: left = %d", e.m.Left())
	}
	e.send(tea.WindowSizeMsg{Width: 40, Height: 24}) // ancho de texto 36
	if want := 70 - 36 + 1; e.m.Left() != want {
		t.Errorf("tras estrechar: left = %d, quería %d", e.m.Left(), want)
	}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24})
	if e.m.Left() != 0 && e.m.Cursor().Col-e.m.Left() >= 76 {
		t.Errorf("tras ensanchar el cursor debe seguir visible: left = %d", e.m.Left())
	}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 10}) // menos filas
	e.send(keyHome)
	if e.m.Top() != 0 {
		t.Errorf("top = %d", e.m.Top())
	}
}

func TestSaveWithCtrlS(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("uno\n"), 0o644)
	m, err := editor.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24}, keyEnd).typ("!").send(keyCtrlS)
	if got, _ := os.ReadFile(p); string(got) != "uno!\n" {
		t.Errorf("archivo = %q", got)
	}
	if e.m.Buffer().Modified() {
		t.Error("tras guardar no debe estar modificado")
	}
	if !strings.Contains(e.m.Message(), "escritas") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	if e.quit {
		t.Error("guardar no debe salir")
	}
}

func TestSaveWithoutNameShowsError(t *testing.T) {
	e := newEd(t, "").typ("x").send(keyCtrlS)
	if !strings.HasPrefix(e.m.Message(), "E32") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestSaveFailureShowsErrorAndKeepsModified(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root no se puede quitar el permiso de escritura")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("x")
	os.Chmod(dir, 0o555)
	defer os.Chmod(dir, 0o755)
	e.send(keyCtrlS)
	if !strings.HasPrefix(e.m.Message(), "E212") || !e.m.Buffer().Modified() {
		t.Errorf("mensaje = %q, modificado = %v", e.m.Message(), e.m.Buffer().Modified())
	}
	if got, _ := os.ReadFile(p); string(got) != "a\n" {
		t.Errorf("el original cambió: %q", got)
	}
}

func TestQuitWithoutChanges(t *testing.T) {
	for _, k := range []tea.KeyPressMsg{keyCtrlQ, keyCtrlC} {
		e := newEd(t, "hola\n").send(k)
		if !e.quit || e.m.DialogOpen() {
			t.Errorf("%v sin cambios debe salir directamente", k)
		}
	}
}

func TestQuitWithChangesAsksFirst(t *testing.T) {
	e := newEd(t, "hola\n").typ("x").send(keyCtrlQ)
	if e.quit || !e.m.DialogOpen() {
		t.Fatalf("con cambios debe abrir el diálogo y no salir (quit=%v)", e.quit)
	}
	e.send(keyEsc)
	if e.quit || e.m.DialogOpen() {
		t.Fatal("Esc cancela y se sigue editando")
	}
	if got := e.m.Buffer().Line(0); got != "xhola" {
		t.Errorf("el texto no debe cambiar: %q", got)
	}

	e.send(keyCtrlQ).typ("zzz") // el diálogo ignora el texto
	if got := e.m.Buffer().Line(0); got != "xhola" {
		t.Errorf("el diálogo no debe dejar escribir: %q", got)
	}
	e.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if !e.quit {
		t.Fatal("«Descartar» debe salir")
	}
}

func TestDialogSaveAndQuit(t *testing.T) {
	for _, how := range []string{"y", "enter"} {
		t.Run(how, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "f.txt")
			os.WriteFile(p, []byte("a\n"), 0o644)
			m, _ := editor.Open(p)
			e := &ed{t: t, m: m}
			e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("b").send(keyCtrlQ)
			if how == "y" {
				e.typ("y")
			} else {
				e.send(keyEnter) // «Guardar» está seleccionado por defecto
			}
			if !e.quit {
				t.Fatal("debe salir tras guardar")
			}
			if got, _ := os.ReadFile(p); string(got) != "ba\n" {
				t.Errorf("archivo = %q", got)
			}
		})
	}
}

func TestDialogNavigation(t *testing.T) {
	e := newEd(t, "").typ("x").send(keyCtrlQ)
	e.send(keyRight, keyEnter) // «Descartar»
	if !e.quit {
		t.Fatal("Descartar con Enter debe salir")
	}

	e = newEd(t, "").typ("x").send(keyCtrlQ)
	e.send(keyRight, keyRight, keyEnter) // «Cancelar»
	if e.quit || e.m.DialogOpen() {
		t.Fatal("Cancelar con Enter cierra el diálogo")
	}

	e = newEd(t, "").typ("x").send(keyCtrlQ)
	e.send(keyLeft, keyEnter) // vuelta circular a «Cancelar»
	if e.quit || e.m.DialogOpen() {
		t.Fatal("Left desde «Guardar» pasa a «Cancelar»")
	}
}

func TestDialogSaveFailureStaysOpenEditor(t *testing.T) {
	e := newEd(t, "").typ("x").send(keyCtrlQ, tea.KeyPressMsg{Code: 'y', Text: "y"}) // sin nombre: no se puede guardar
	if e.quit {
		t.Fatal("si falla el guardado no se sale")
	}
	if e.m.DialogOpen() || !strings.HasPrefix(e.m.Message(), "E32") {
		t.Errorf("dialogo abierto=%v mensaje=%q", e.m.DialogOpen(), e.m.Message())
	}
}

func TestReadOnlyBlocksEditing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bin")
	os.WriteFile(p, []byte("ab\x00cd"), 0o644)
	m, err := editor.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.Message(), "Solo lectura") {
		t.Errorf("al abrir debe avisar: %q", m.Message())
	}
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("x")
	if e.m.Buffer().Modified() || e.text() != "ab\x00cd" {
		t.Fatalf("el texto cambió: %q", e.text())
	}
	if !strings.Contains(e.m.Message(), "Solo lectura") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	for _, k := range []tea.Msg{keyEnter, keyBackspace, keyDelete, keyTab, tea.PasteMsg{Content: "z"}, keyCtrlS} {
		e.send(k)
	}
	if e.m.Buffer().Modified() || e.text() != "ab\x00cd" {
		t.Fatalf("el texto cambió: %q", e.text())
	}
	e.send(keyCtrlQ)
	if !e.quit {
		t.Error("un buffer de solo lectura sale sin diálogo")
	}
}

func TestOpenDirectoryFails(t *testing.T) {
	if _, err := editor.Open(t.TempDir()); err == nil {
		t.Fatal("abrir un directorio debía fallar")
	}
}

func TestMinimumSize(t *testing.T) {
	e := newEd(t, "hola\n").send(tea.WindowSizeMsg{Width: 30, Height: 8})
	if !e.m.TooSmall() || !strings.Contains(e.m.View().Content, "demasiado pequeña") {
		t.Fatalf("debe avisar: %q", e.m.View().Content)
	}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24})
	if e.m.TooSmall() || strings.Contains(e.m.View().Content, "demasiado pequeña") {
		t.Fatal("al agrandar debe volver la interfaz")
	}
	if e.m.Buffer().Line(0) != "hola" {
		t.Error("el contenido se conserva al redimensionar")
	}
}

func TestTooSmallBoundaries(t *testing.T) {
	for _, tc := range []struct {
		w, h int
		want bool
	}{
		{0, 0, false}, // tamaño aún desconocido
		{39, 24, true},
		{80, 9, true},
		{40, 10, false},
		{30, 8, true},
		{200, 60, false},
	} {
		m := editor.New(buffer.New(""), "")
		nm, _ := m.Update(tea.WindowSizeMsg{Width: tc.w, Height: tc.h})
		if got := nm.(editor.Model).TooSmall(); got != tc.want {
			t.Errorf("%dx%d: TooSmall() = %v, quería %v", tc.w, tc.h, got, tc.want)
		}
	}
}

func TestViewBeforeFirstSizeDoesNotPanic(t *testing.T) {
	v := editor.New(buffer.New(""), "").View()
	if !v.AltScreen {
		t.Error("debe usar la pantalla alternativa")
	}
}

func TestViewCursorFollowsText(t *testing.T) {
	e := newEd(t, "").typ("ab")
	v := e.m.View()
	if v.Cursor == nil {
		t.Fatal("la vista debe incluir el cursor")
	}
	// gutter de 4 columnas + 2 caracteres
	if v.Cursor.X != 6 || v.Cursor.Y != 0 {
		t.Errorf("cursor en (%d,%d), quería (6,0)", v.Cursor.X, v.Cursor.Y)
	}
	e.send(keyCtrlQ)
	e.typ("")
	if e.m.DialogOpen() && e.m.View().Cursor != nil {
		t.Error("con el diálogo abierto no se muestra el cursor de edición")
	}
}

// --- pruebas con el bucle real de Bubble Tea (teatest) ---

func contains(s string) func([]byte) bool {
	return func(b []byte) bool { return bytes.Contains(b, []byte(s)) }
}

func TestTeatestTypeSaveQuit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "doc.txt")
	m, _ := editor.Open(p)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	teatest.WaitFor(t, tm.Output(), contains("INSERT"), teatest.WithDuration(wait))
	tm.Type("hola")
	tm.Send(keyEnter)
	tm.Type("mundo")
	tm.Send(keyCtrlS)
	teatest.WaitFor(t, tm.Output(), contains("escritas"), teatest.WithDuration(wait))
	tm.Send(keyCtrlQ)
	tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
	if got, _ := os.ReadFile(p); string(got) != "hola\nmundo\n" {
		t.Errorf("archivo = %q", got)
	}
}

func TestTeatestUnsavedDialog(t *testing.T) {
	m, _ := editor.Open("")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	tm.Type("x")
	tm.Send(keyCtrlQ)
	teatest.WaitFor(t, tm.Output(), contains("Cambios sin guardar"), teatest.WithDuration(wait))
	tm.Type("n")
	tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
}

func TestTeatestResize(t *testing.T) {
	m, _ := editor.Open("")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(30, 8))
	teatest.WaitFor(t, tm.Output(), contains("demasiado pequeña"), teatest.WithDuration(wait))
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	teatest.WaitFor(t, tm.Output(), contains("INSERT"), teatest.WithDuration(wait))
	tm.Send(keyCtrlQ)
	tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
}

func ExampleModel_View() {
	m := editor.New(buffer.FromString("", "hola"), "demo.txt")
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	fmt.Println(nm.(editor.Model).View().WindowTitle)
	// Output: demo.txt - fos
}
