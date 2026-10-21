package editor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

// files crea archivos con contenido «nombre\n» en un directorio temporal y devuelve sus rutas.
func files(t *testing.T, names ...string) []string {
	t.Helper()
	dir := t.TempDir()
	var out []string
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte("contenido "+n+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

// openMany abre un editor con el primer archivo y los demás en buffers adicionales; queda activo el último.
func openMany(t *testing.T, paths ...string) *ed {
	t.Helper()
	m, err := editor.Open(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24})
	for _, p := range paths[1:] {
		e.ex("e " + p)
	}
	return e
}

// touchLater cambia un archivo en disco desde «fuera» de forma que mtime y tamaño difieran.
func touchLater(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(3 * time.Second)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}
}

func TestReopenReusesBufferAndRestoresPosition(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.txt"), filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte(strings.Repeat("línea de a\n", 60)), 0o644)
	os.WriteFile(b, []byte("bbb\n"), 0o644)
	e := openMany(t, a)
	e.typ("40G").typ("w")
	pos, top := e.m.Cursor(), e.m.Top()
	if pos.Line != 39 || top == 0 {
		t.Fatalf("preparación: cursor %v top %d", pos, top)
	}
	e.ex("e " + b)
	if e.m.BufferCount() != 2 {
		t.Fatalf("buffers = %d", e.m.BufferCount())
	}
	e.wantText("bbb\n")
	e.ex("e " + a) // ya estaba abierto: se reutiliza
	if e.m.BufferCount() != 2 {
		t.Errorf("no deben duplicarse los buffers: %d", e.m.BufferCount())
	}
	if e.m.Cursor() != pos || e.m.Top() != top {
		t.Errorf("se restauran cursor y scroll: %v top %d, quería %v top %d", e.m.Cursor(), e.m.Top(), pos, top)
	}
}

func TestReopenWithDifferentSpellingOfThePath(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.ex("e " + filepath.Join(filepath.Dir(p[0]), ".", "sub", "..", "a.txt"))
	if e.m.BufferCount() != 2 || e.m.ActiveBuffer() != 0 {
		t.Errorf("la misma ruta escrita distinto reutiliza el buffer: %d, activo %d", e.m.BufferCount(), e.m.ActiveBuffer())
	}
}

func TestBuffersKeepIndependentState(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ixx").send(keyEsc) // cambios en b
	e.ex("bp")
	e.wantText("contenido a.txt\n")
	if e.m.Buffer().Modified() {
		t.Error("a no se ve afectado por los cambios de b")
	}
	e.ex("bn")
	e.wantText("xxcontenido b.txt\n")
	if !e.m.Buffer().Modified() {
		t.Error("b conserva sus cambios al volver")
	}
	e.at(0, 1) // el cursor de b quedó donde estaba
}

func TestSwitchBuffers(t *testing.T) {
	p := files(t, "a.txt", "b.txt", "c.txt")
	e := openMany(t, p[0], p[1], p[2])
	if e.m.ActiveBuffer() != 2 || e.m.BufferCount() != 3 {
		t.Fatal(e.m.ActiveBuffer())
	}
	e.space("bn")
	if e.m.ActiveBuffer() != 0 {
		t.Errorf("Space b n da la vuelta: %d", e.m.ActiveBuffer())
	}
	e.space("bp")
	if e.m.ActiveBuffer() != 2 {
		t.Errorf("Space b p da la vuelta hacia atrás: %d", e.m.ActiveBuffer())
	}
	e.ex("bp")
	if e.m.ActiveBuffer() != 1 || !strings.Contains(e.screen(), "2/3") {
		t.Errorf(":bp y el indicador 2/3: activo %d\n%s", e.m.ActiveBuffer(), e.screen())
	}
	e.ex("bn")
	if e.m.ActiveBuffer() != 2 {
		t.Errorf(":bn: %d", e.m.ActiveBuffer())
	}
	e.ex("bnext").ex("bprev")
	if e.m.ActiveBuffer() != 2 {
		t.Errorf("sinónimos: %d", e.m.ActiveBuffer())
	}
}

func TestBufferIndicatorOnlyWithSeveralBuffers(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0])
	if strings.Contains(e.screen(), "1/1") {
		t.Error("con un solo buffer no se muestra el indicador")
	}
	e.send(keyEsc).typ("x")
	e.ex("bn")
	if !strings.Contains(e.m.Message(), "No hay otros buffers") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.ex("e " + p[1])
	if !strings.Contains(e.screen(), "2/2") {
		t.Errorf("con dos buffers:\n%s", e.screen())
	}
}

func TestSwitchingResetsModeAndPending(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("v") // Visual en b
	e.send(keyEsc)
	e.ex("bn")
	e.mode(editor.ModeNormal)
	if e.m.Pending() != "" {
		t.Errorf("pendiente = %q", e.m.Pending())
	}
}

func TestCloseBuffers(t *testing.T) {
	p := files(t, "a.txt", "b.txt", "c.txt")
	e := openMany(t, p[0], p[1], p[2]) // activo c
	e.space("bd")
	if e.m.BufferCount() != 2 || e.m.ActiveBuffer() != 1 {
		t.Errorf("tras cerrar c: %d buffers, activo %d", e.m.BufferCount(), e.m.ActiveBuffer())
	}
	e.wantText("contenido b.txt\n")
	e.ex("bp").ex("bd") // cerrar a (el primero): pasa a b
	e.wantText("contenido b.txt\n")
	if e.m.BufferCount() != 1 {
		t.Fatalf("buffers = %d", e.m.BufferCount())
	}
	e.ex("bd") // el último: queda un buffer vacío sin nombre y no se sale
	if e.quit || e.m.BufferCount() != 1 || e.m.Buffer().Path() != "" || e.text() != "" {
		t.Errorf("cerrar el último deja un buffer vacío: quit=%v buffers=%d path=%q texto=%q",
			e.quit, e.m.BufferCount(), e.m.Buffer().Path(), e.text())
	}
	if !strings.Contains(e.screen(), "[Sin nombre]") {
		t.Error("la statusline muestra [Sin nombre]")
	}
}

func TestCloseModifiedBufferAsksFirst(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc)
	e.ex("bd")
	if !e.m.DialogOpen() || !strings.Contains(e.screen(), "Cambios sin guardar") || !strings.Contains(e.screen(), "b.txt") {
		t.Fatalf("diálogo de la maqueta 10:\n%s", e.screen())
	}
	e.send(keyEsc)
	if e.m.BufferCount() != 2 || e.m.DialogOpen() {
		t.Error("Cancelar no cierra")
	}
	e.ex("bd").typ("n") // descartar
	if e.m.BufferCount() != 1 {
		t.Fatal("Descartar cierra")
	}
	if got, _ := os.ReadFile(p[1]); string(got) != "contenido b.txt\n" {
		t.Error("descartar no escribe")
	}
	e.ex("e " + p[1])
	e.typ("iy").send(keyEsc).ex("bd").typ("y") // guardar y cerrar
	if e.m.BufferCount() != 1 {
		t.Error("Guardar cierra")
	}
	if got, _ := os.ReadFile(p[1]); string(got) != "ycontenido b.txt\n" {
		t.Errorf("guardar escribe: %q", got)
	}
	e.ex("e " + p[1]).typ("iz").send(keyEsc).ex("bd!")
	if e.m.BufferCount() != 1 || e.m.DialogOpen() {
		t.Error(":bd! descarta sin preguntar")
	}
}

func TestCloseModifiedBufferSaveFailureKeepsIt(t *testing.T) {
	e := newEd(t, "").typ("ix").send(keyEsc).ex("bd").typ("y") // sin nombre: no se puede guardar
	if !strings.HasPrefix(e.m.Message(), "E32") || e.text() != "x\n" {
		t.Errorf("msg=%q texto=%q", e.m.Message(), e.text())
	}
}

func TestQuitListsAllModifiedBuffers(t *testing.T) {
	p := files(t, "a.txt", "b.txt", "c.txt")
	e := openMany(t, p[0], p[1], p[2])
	e.ex("bp").typ("iB").send(keyEsc).ex("bp").typ("iA").send(keyEsc) // modificados: b y a; c queda intacto
	e.ex("q")
	if e.quit || !e.m.DialogOpen() {
		t.Fatal("con buffers modificados no se sale")
	}
	scr := e.screen()
	for _, want := range []string{"a.txt", "b.txt", "Guardar todo", "Descartar todo", "Cancelar", "2 archivos"} {
		if !strings.Contains(scr, want) {
			t.Errorf("el diálogo debe mostrar %q:\n%s", want, scr)
		}
	}
	if strings.Contains(dialogText(scr), "c.txt") {
		t.Error("c.txt no está modificado y no debe listarse")
	}
	e.send(keyEsc)
	if e.quit || e.m.DialogOpen() {
		t.Error("Cancelar sigue editando")
	}
	e.ex("q").typ("y") // guardar todo y salir
	if !e.quit {
		t.Fatal("tras guardar todo se sale")
	}
	for i, want := range map[int]string{0: "Acontenido a.txt\n", 1: "Bcontenido b.txt\n"} {
		if got, _ := os.ReadFile(p[i]); string(got) != want {
			t.Errorf("%s = %q, quería %q", filepath.Base(p[i]), got, want)
		}
	}
}

func TestQuitDiscardAllAndBang(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc).ex("bp").typ("iy").send(keyEsc)
	e.ex("q").typ("n")
	if !e.quit {
		t.Fatal("Descartar todo sale")
	}
	for _, f := range p {
		if got, _ := os.ReadFile(f); !strings.HasPrefix(string(got), "contenido") {
			t.Errorf("descartar no escribe: %q", got)
		}
	}
	e = openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc).ex("q!")
	if !e.quit {
		t.Error(":q! sale sin preguntar")
	}
}

func TestWriteQuitWithOtherModifiedBuffers(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc).ex("bp").typ("iy").send(keyEsc) // activo a, ambos modificados
	e.ex("wq")
	if e.quit || !e.m.DialogOpen() {
		t.Fatal(":wq guarda el actual pero no sale mientras otro buffer tenga cambios")
	}
	if got, _ := os.ReadFile(p[0]); string(got) != "ycontenido a.txt\n" {
		t.Errorf("el actual se guardó: %q", got)
	}
	if !strings.Contains(dialogText(e.screen()), "b.txt") || strings.Contains(dialogText(e.screen()), "a.txt") {
		t.Errorf("el diálogo lista solo b.txt:\n%s", e.screen())
	}
}

func TestSaveAllConflictAbortsQuit(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc).ex("bp").typ("iy").send(keyEsc)
	touchLater(t, p[1], "cambiado fuera, más largo que antes\n")
	e.ex("q").typ("y")
	if e.quit {
		t.Fatal("un conflicto con el disco impide salir")
	}
	if got, _ := os.ReadFile(p[1]); string(got) != "cambiado fuera, más largo que antes\n" {
		t.Errorf("no se sobrescribe lo que cambió fuera: %q", got)
	}
	if !strings.HasPrefix(e.m.Message(), "E13") || !strings.Contains(e.m.Message(), "b.txt") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestExternalChangeReloadsUnmodifiedBufferOnActivate(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	touchLater(t, p[0], "nuevo contenido de a, escrito desde fuera\n")
	e.ex("bp")
	e.wantText("nuevo contenido de a, escrito desde fuera\n")
	if !strings.Contains(e.m.Message(), "cambió en disco") || !strings.Contains(e.m.Message(), "recargado") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	if e.m.Buffer().Modified() {
		t.Error("recargado no queda modificado")
	}
	e.ex("bn").ex("bp") // sin nuevos cambios no recarga ni avisa
	if e.m.Message() != "" && strings.Contains(e.m.Message(), "cambió") {
		t.Errorf("no hay cambios nuevos: %q", e.m.Message())
	}
}

func TestExternalChangeWarnsModifiedBufferAndAsksToOverwrite(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.ex("bp").typ("ix").send(keyEsc) // a modificado
	e.ex("bn")
	touchLater(t, p[0], "cambiado fuera\n")
	e.ex("bp")
	e.wantText("xcontenido a.txt\n") // el buffer modificado no se recarga
	if !strings.HasPrefix(e.m.Message(), "W12") {
		t.Errorf("aviso: %q", e.m.Message())
	}
	e.send(keyCtrlS) // guardar sobre cambios externos
	if !e.m.DialogOpen() || !strings.Contains(e.screen(), "cambió en disco") || !strings.Contains(e.screen(), "Sobrescribir") {
		t.Fatalf("pide confirmación:\n%s", e.screen())
	}
	e.send(keyEsc) // cancelar
	if got, _ := os.ReadFile(p[0]); string(got) != "cambiado fuera\n" {
		t.Errorf("cancelar no toca el disco: %q", got)
	}
	if !e.m.Buffer().Modified() {
		t.Error("el buffer sigue modificado")
	}
	e.ex("w")
	e.typ("y") // confirmar con y
	if got, _ := os.ReadFile(p[0]); string(got) != "xcontenido a.txt\n" {
		t.Errorf("confirmar sobrescribe: %q", got)
	}
	if e.m.Buffer().Modified() {
		t.Error("tras sobrescribir queda limpio")
	}
	e.send(keyCtrlS) // ya no hay conflicto: guarda sin preguntar
	if e.m.DialogOpen() {
		t.Error("sin conflicto no se pregunta")
	}
}

func TestExternalChangeOnTerminalFocus(t *testing.T) {
	p := files(t, "a.txt")
	e := openMany(t, p[0])
	touchLater(t, p[0], "versión nueva\n")
	e.send(tea.FocusMsg{})
	e.wantText("versión nueva\n")
	if !strings.Contains(e.m.Message(), "recargado") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestDeletedFileIsNotTreatedAsChanged(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	os.Remove(p[0])
	e.ex("bp")
	e.wantText("contenido a.txt\n") // sigue en memoria
	e.typ("ix").send(keyEsc).ex("w")
	if got, _ := os.ReadFile(p[0]); string(got) != "xcontenido a.txt\n" {
		t.Errorf("guardar recrea el archivo: %q", got)
	}
}

func TestSaveRefreshesDiskSnapshot(t *testing.T) {
	p := files(t, "a.txt")
	e := openMany(t, p[0])
	e.typ("ix").send(keyEsc).ex("w")
	e.send(tea.FocusMsg{}) // lo guardado por el propio editor no cuenta como cambio externo
	if strings.Contains(e.m.Message(), "cambió") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.typ("iy").send(keyEsc).send(keyCtrlS)
	if e.m.DialogOpen() {
		t.Error("tras guardar no hay conflicto")
	}
}

func TestReloadActiveWithEBang(t *testing.T) {
	p := files(t, "a.txt")
	e := openMany(t, p[0])
	e.typ("ix").send(keyEsc).ex("e!")
	e.wantText("contenido a.txt\n")
	if e.m.Buffer().Modified() {
		t.Error("recargar descarta los cambios")
	}
	touchLater(t, p[0], "otro\n")
	e.ex("e")
	e.wantText("otro\n")
}

// dialogText devuelve solo el texto del cuadro de diálogo de una pantalla (entre su borde superior e inferior).
func dialogText(screen string) string {
	i, j := strings.Index(screen, "┌"), strings.Index(screen, "└")
	if i < 0 || j < i {
		return ""
	}
	return screen[i:j]
}
