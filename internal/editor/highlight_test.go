package editor_test

import (
	"bytes"
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

const (
	darkMauve  = "38;2;204;168;246"
	darkGray   = "38;2;117;128;156"
	darkGreen  = "38;2;127;208;150"
	darkBlue   = "38;2;127;193;254"
	darkText   = "38;2;207;215;235"
	lightMauve = "38;2;125;77;173"
)

// run ejecuta un comando de Bubble Tea con un tiempo máximo y devuelve los mensajes que produce
// (los lotes se despliegan). Los temporizadores largos (p. ej. las teclas pendientes) se abandonan.
func run(c tea.Cmd) []tea.Msg {
	if c == nil {
		return nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- c() }()
	select {
	case msg := <-ch:
		if b, ok := msg.(tea.BatchMsg); ok {
			var out []tea.Msg
			for _, c := range b {
				out = append(out, run(c)...)
			}
			return out
		}
		if msg == nil {
			return nil
		}
		return []tea.Msg{msg}
	case <-time.After(400 * time.Millisecond):
		return nil
	}
}

// settle ejecuta los comandos pendientes y devuelve sus mensajes al modelo hasta que no quede trabajo (el resaltado asíncrono).
func (e *ed) settle() *ed {
	deadline := time.Now().Add(5 * time.Second)
	for len(e.cmds) > 0 && time.Now().Before(deadline) {
		cmds := e.cmds
		e.cmds = nil
		for _, c := range cmds {
			for _, msg := range run(c) {
				e.send(msg)
			}
		}
	}
	return e
}

// rows devuelve las filas de la pantalla con sus colores.
func (e *ed) rows() []string { return strings.Split(e.m.View().Content, "\n") }

// sgrBefore devuelve la última secuencia de color antes de text en una fila.
func sgrBefore(row, text string) string {
	i := strings.Index(row, text)
	if i < 0 {
		return ""
	}
	pre := row[:i]
	return pre[strings.LastIndex(pre, "\x1b["):]
}

// colored abre un archivo con contenido en un directorio temporal y devuelve un editor de 100x24 sobre él.
func colored(t *testing.T, name, content string, opts ...editor.Option) *ed {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	opts = append([]editor.Option{editor.WithHighlightDelay(time.Millisecond)}, opts...)
	m, err := editor.Open(p, opts...)
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	return e.send(tea.WindowSizeMsg{Width: 100, Height: 24})
}

const goProg = "package demo\n\n// Hola\nfunc f() int {\n\treturn 12\n}\n"

func TestSmallFilesAreHighlightedOnOpen(t *testing.T) {
	e := colored(t, "a.go", goProg)
	rows := e.rows()
	if !strings.Contains(sgrBefore(rows[0], "package"), darkMauve) {
		t.Errorf("la palabra clave va en mauve nada más abrir: %q", sgrBefore(rows[0], "package"))
	}
	if !strings.Contains(sgrBefore(rows[2], "// Hola"), darkGray) {
		t.Errorf("el comentario va en gris: %q", sgrBefore(rows[2], "// Hola"))
	}
	if sgr := sgrBefore(rows[3][strings.Index(rows[3], "func")+4:], "f"); !strings.Contains(sgr, darkBlue) {
		t.Errorf("la función va en azul: %q", sgr)
	}
	if e.m.HighlightJobs() != 0 {
		t.Error("un texto pequeño se resalta al cargar, sin trabajo en segundo plano")
	}
	if e.m.HighlightRev() != e.m.Buffer().Rev() {
		t.Error("los colores corresponden al texto actual")
	}
}

func TestUnknownTypeIsPlainWithoutErrors(t *testing.T) {
	e := colored(t, "notas.xyz", "package demo func return 12 \"x\"\n")
	row := e.rows()[0]
	for _, w := range []string{"package", "func", "return"} {
		if sgr := sgrBefore(row, w); strings.Contains(sgr, darkMauve) {
			t.Errorf("%s no debe colorearse en un tipo desconocido: %q", w, sgr)
		}
	}
	if e.m.HighlightRev() != -1 {
		t.Error("sin lenguaje no hay colores")
	}
	e.typ("ix").settle() // editar tampoco lanza trabajo
	if e.m.HighlightJobs() != 0 {
		t.Errorf("trabajos = %d", e.m.HighlightJobs())
	}
}

func TestTypingDoesNotWaitForHighlighting(t *testing.T) {
	e := colored(t, "a.go", goProg)
	e.typ("A")
	nm, cmd := e.m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	e.m = nm.(editor.Model)
	if cmd == nil {
		t.Fatal("la edición programa el resaltado")
	}
	if !strings.Contains(e.m.Buffer().Line(0), "x") || !strings.Contains(e.screenPlain(), "demox") {
		t.Errorf("la pulsación se ve al instante:\n%s", e.screenPlain())
	}
	if e.m.HighlightJobs() != 0 {
		t.Error("todavía no se ha tokenizado nada: el resaltado va después")
	}
	if e.m.HighlightRev() >= e.m.Buffer().Rev() {
		t.Error("los colores son de la versión anterior hasta que llegue el resultado")
	}
	// mientras tanto se siguen mostrando los colores anteriores
	if !strings.Contains(sgrBefore(e.rows()[3], "func"), darkMauve) {
		t.Error("las demás líneas conservan su color")
	}
	e.cmds = append(e.cmds, cmd)
	e.settle()
	if e.m.HighlightJobs() != 1 || e.m.HighlightRev() != e.m.Buffer().Rev() {
		t.Errorf("tras terminar: trabajos=%d rev=%d/%d", e.m.HighlightJobs(), e.m.HighlightRev(), e.m.Buffer().Rev())
	}
}

func (e *ed) screenPlain() string { return e.screen() }

func TestDebounceCollapsesABurstOfEdits(t *testing.T) {
	e := colored(t, "a.go", goProg).typ("A").typ("abcdefghij").settle()
	if e.m.HighlightJobs() != 1 {
		t.Errorf("una ráfaga de pulsaciones lanza una sola tokenización: %d", e.m.HighlightJobs())
	}
	e.send(keyEsc).typ("x").settle()
	if e.m.HighlightJobs() != 2 {
		t.Errorf("otra edición, otra tokenización: %d", e.m.HighlightJobs())
	}
}

func TestStaleResultsAreDiscarded(t *testing.T) {
	e := colored(t, "a.go", goProg).typ("A").typ("x")
	// la primera edición llega a tokenizar...
	var job tea.Cmd
	for _, c := range e.cmds {
		for _, msg := range run(c) {
			nm, next := e.m.Update(msg)
			e.m = nm.(editor.Model)
			if next != nil && job == nil {
				job = next
			}
		}
	}
	e.cmds = nil
	if job == nil {
		t.Fatal("debía lanzarse la tokenización")
	}
	oldRev := e.m.Buffer().Rev()
	e.typ("y") // ...pero el texto cambia antes de que termine
	for _, msg := range run(job) {
		e.send(msg)
	}
	if e.m.HighlightRev() >= e.m.Buffer().Rev() || e.m.HighlightRev() == oldRev {
		t.Errorf("el resultado viejo (rev %d) no se instala: colores en rev %d, texto en rev %d", oldRev, e.m.HighlightRev(), e.m.Buffer().Rev())
	}
	e.settle()
	if e.m.HighlightRev() != e.m.Buffer().Rev() {
		t.Errorf("al final se tokeniza la versión actual: %d/%d", e.m.HighlightRev(), e.m.Buffer().Rev())
	}
}

func TestMultilineConstructAfterEdit(t *testing.T) {
	src := "package a\n\nfunc f() {\n\tx := 1\n\ty := 2\n}\n\n// fin\n"
	e := colored(t, "a.go", src)
	if !strings.Contains(sgrBefore(e.rows()[2], "func"), darkMauve) {
		t.Fatal("preparación: sin comentario, «func» es una palabra clave")
	}
	// se abre un comentario de bloque encima de la función y se cierra al final del archivo
	e.typ("gg").typ("2j").typ("O").typ("/*").send(keyEsc)
	e.typ("G").typ("o").typ("*/").send(keyEsc).settle()
	rows := e.rows()
	for i := 2; i <= 9; i++ { // desde «/*» hasta «*/», todo es comentario
		if strings.Contains(rows[i], darkMauve) || strings.Contains(rows[i], "38;2;245;165;116") {
			t.Errorf("fila %d dentro del comentario no debe tener colores de código: %q", i, rows[i])
		}
	}
	if sgr := sgrBefore(rows[3], "func"); !strings.Contains(sgr, darkGray) {
		t.Errorf("«func» pasa a verse como comentario: %q", sgr)
	}
	if !strings.Contains(sgrBefore(rows[0], "package"), darkMauve) {
		t.Error("lo anterior al comentario no cambia")
	}
	// quitar el cierre deshace el efecto: el comentario ya no se cierra y vuelve el código
	e.typ("u").settle()
	if sgr := sgrBefore(e.rows()[3], "func"); !strings.Contains(sgr, darkMauve) {
		t.Errorf("sin el cierre «func» vuelve a ser una palabra clave: %q", sgr)
	}
}

func TestSpansShiftWithoutWaitingForRetokenization(t *testing.T) {
	e := colored(t, "a.go", goProg)
	before := e.rows()
	if !strings.Contains(sgrBefore(before[3], "func"), darkMauve) {
		t.Fatal("preparación")
	}
	// abrir una línea encima de «package»: todo baja una fila
	e.typ("O").send(keyEsc)
	rows := e.rows()
	if !strings.Contains(sgrBefore(rows[4], "func"), darkMauve) {
		t.Errorf("«func» bajó a la fila 4 conservando su color antes de retokenizar: %q", rows[4])
	}
	if !strings.Contains(sgrBefore(rows[3], "// Hola"), darkGray) {
		t.Errorf("el comentario también bajó: %q", rows[3])
	}
	if e.m.HighlightRev() == e.m.Buffer().Rev() {
		t.Fatal("el resultado asíncrono aún no ha llegado")
	}
	// borrar la línea nueva: sube todo otra vez
	e.typ("dd")
	rows = e.rows()
	if !strings.Contains(sgrBefore(rows[3], "func"), darkMauve) || !strings.Contains(sgrBefore(rows[0], "package"), darkMauve) {
		t.Errorf("tras borrar la línea todo vuelve a su sitio:\n%s", strings.Join(rows[:5], "\n"))
	}
	// y deshacer también
	e.typ("u")
	if !strings.Contains(sgrBefore(e.rows()[4], "func"), darkMauve) {
		t.Error("deshacer desplaza los colores como cualquier edición")
	}
	e.settle()
	if e.m.HighlightRev() != e.m.Buffer().Rev() {
		t.Error("al final todo se retokeniza")
	}
}

func TestTypingInsideAStringKeepsItsColor(t *testing.T) {
	e := colored(t, "a.go", "x := \"hola mundo\" + 1\n")
	e.typ("07l").typ("i").typ("zzz")
	row := e.rows()[0]
	if !strings.Contains(sgrBefore(row, "zzz"), darkGreen) {
		t.Errorf("el texto escrito dentro de la cadena sigue siendo cadena mientras llega el resultado: %q", sgrBefore(row, "zzz"))
	}
	if !strings.Contains(sgrBefore(row, "mundo"), darkGreen) {
		t.Errorf("el resto de la cadena conserva el color: %q", sgrBefore(row, "mundo"))
	}
}

func TestSafetyLimitsInEditor(t *testing.T) {
	// archivo mayor de 2 MiB: sin color y sin errores
	big := strings.Repeat("func main() {}\n", 2<<20/15+100)
	e := colored(t, "big.go", big)
	if strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) {
		t.Error("un archivo de más de 2 MiB no se resalta")
	}
	e.typ("ix").settle()
	if strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) {
		t.Error("tras editar sigue sin resaltarse")
	}

	// línea gigantesca: sin color; el resto sí
	long := "var a = \"" + strings.Repeat("x", 100000) + "\"\nfunc main() {}\n"
	e = colored(t, "long.go", long)
	rows := e.rows()
	if strings.Contains(rows[0], darkGreen) {
		t.Error("la línea de 100000 caracteres se muestra sin color")
	}
	if !strings.Contains(sgrBefore(rows[1], "func"), darkMauve) {
		t.Errorf("el resto del archivo sigue coloreado: %q", sgrBefore(rows[1], "func"))
	}
}

func TestThemeToggleRecolorsWithoutRetokenizing(t *testing.T) {
	e := colored(t, "a.go", goProg).typ("A").typ("x").send(keyEsc).settle()
	jobs := e.m.HighlightJobs()
	dark := e.rows()
	if !strings.Contains(sgrBefore(dark[0], "package"), darkMauve) || !strings.Contains(dark[1], "48;2;23;27;39") {
		t.Fatalf("tema oscuro: %q", dark[0])
	}
	e.typ(" t") // Space t
	if !strings.Contains(e.m.Message(), "claro") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	light := e.rows()
	if !strings.Contains(sgrBefore(light[0], "package"), lightMauve) {
		t.Errorf("la palabra clave pasa al mauve del tema claro: %q", sgrBefore(light[0], "package"))
	}
	if !strings.Contains(light[1], "48;2;244;247;252") { // fondo claro #f4f7fc
		t.Errorf("fondo claro: %q", light[1])
	}
	if strings.Contains(strings.Join(light, ""), darkMauve) {
		t.Error("no queda ningún color del tema oscuro")
	}
	if e.settle(); e.m.HighlightJobs() != jobs {
		t.Errorf("cambiar de tema no retokeniza: %d → %d trabajos", jobs, e.m.HighlightJobs())
	}
	e.typ(" t")
	if !strings.Contains(sgrBefore(e.rows()[0], "package"), darkMauve) || !strings.Contains(e.m.Message(), "oscuro") {
		t.Error("vuelve al tema oscuro")
	}
}

func TestSaveAsWithAnotherExtensionChangesLanguage(t *testing.T) {
	m, _ := editor.Open("", editor.WithHighlightDelay(time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}).typ("ifunc main() {}").send(keyEsc)
	if strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) {
		t.Fatal("sin nombre no hay lenguaje")
	}
	e.ex("w " + filepath.Join(t.TempDir(), "nuevo.go")).settle()
	if !strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) {
		t.Errorf("al guardar como .go se colorea: %q", sgrBefore(e.rows()[0], "func"))
	}
}

func TestRenameInExplorerChangesLanguage(t *testing.T) {
	root := project(t, "a.txt=func main() {}\n")
	e := openProj(t, root, editor.WithHighlightDelay(time.Millisecond))
	e.ex("e " + filepath.Join(root, "a.txt"))
	if strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) {
		t.Fatal(".txt es texto plano")
	}
	e.space("e").typ("j").typ("r").send(keyBackspace, keyBackspace, keyBackspace).typ("go").send(keyEnter).settle()
	e.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(sgrBefore(e.rows()[0], "func"), darkMauve) && !strings.Contains(sgrBefore(strings.Join(e.rows()[:1], ""), "func"), darkMauve) {
		t.Errorf("tras renombrar a .go se colorea:\n%s", e.rows()[0])
	}
}

func TestReloadRehighlights(t *testing.T) {
	e := colored(t, "a.go", "x := 1\n")
	p := e.m.Buffer().Path()
	os.WriteFile(p, []byte("package nuevo\n// y más texto\n"), 0o644)
	later := time.Now().Add(3 * time.Second)
	os.Chtimes(p, later, later)
	e.ex("e")
	e.wantText("package nuevo\n// y más texto\n")
	if !strings.Contains(sgrBefore(e.rows()[0], "package"), darkMauve) {
		t.Errorf("tras recargar se recolorea: %q", e.rows()[0])
	}
}

func TestEachBufferKeepsItsOwnHighlight(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("package a\n"), 0o644)
	os.WriteFile(b, []byte("package b\n"), 0o644)
	m, _ := editor.Open(a, editor.WithHighlightDelay(time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}).ex("e " + b)
	if strings.Contains(sgrBefore(e.rows()[0], "package"), darkMauve) {
		t.Error("b.txt es texto plano")
	}
	e.ex("bp")
	if !strings.Contains(sgrBefore(e.rows()[0], "package"), darkMauve) {
		t.Error("a.go conserva sus colores al volver")
	}
}

func TestMarkdownAndJSONAreHighlighted(t *testing.T) {
	e := colored(t, "datos.json", "{\"clave\": 12, \"ok\": true}\n")
	row := e.rows()[0]
	if sgr := sgrBefore(row[strings.Index(row, "clave"):], "12"); !strings.Contains(sgr, "38;2;245;165;116") {
		t.Errorf("número JSON en melocotón: %q", sgr)
	}
	e = colored(t, "LEEME.MD", "# Título\n\ntexto\n")
	if !strings.Contains(sgrBefore(e.rows()[0], "Título"), darkBlue) && !strings.Contains(sgrBefore(e.rows()[0], "# Título"), darkBlue) {
		t.Errorf("encabezado Markdown (extensión en mayúsculas): %q", e.rows()[0])
	}
}

func TestHighlightWithVisualSelectionAndCursorLine(t *testing.T) {
	e := colored(t, "a.go", goProg)
	row := e.rows()[0] // la línea del cursor
	sgr := sgrBefore(row, "package")
	if !strings.Contains(sgr, darkMauve) || !strings.Contains(sgr, "48;2;35;44;66") {
		t.Errorf("sobre la línea actual: %q", sgr)
	}
	e.typ("vlll")
	sgr = sgrBefore(e.rows()[0], "p")
	if !strings.Contains(sgr, darkMauve) || !strings.Contains(sgr, "48;2;50;61;92") {
		t.Errorf("sobre la selección: %q", sgr)
	}
}

func TestBuffersWithoutTrailingNewlineAndCRLF(t *testing.T) {
	e := colored(t, "a.go", "package a\r\nfunc f() {}")
	rows := e.rows()
	if !strings.Contains(sgrBefore(rows[1], "func"), darkMauve) {
		t.Errorf("CRLF y sin salto final: %q", rows[1])
	}
	_ = buffer.Pos{}
}

// Con el bucle real de Bubble Tea: escribir en un archivo grande no espera a que termine el resaltado.
func TestTeatestTypingDoesNotWaitForHighlight(t *testing.T) {
	src := strings.Repeat("func (n *Node) Lookup(seq []rune) (string, bool) { return \"x\", 12 > 3 } // c\n", 3500) // ~270 KB
	p := filepath.Join(t.TempDir(), "grande.go")
	os.WriteFile(p, []byte(src), 0o644)
	m, _ := editor.Open(p, editor.WithHighlightDelay(5*time.Millisecond))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 24))
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool { return bytes.Contains(b, []byte("NORMAL")) }, teatest.WithDuration(5*time.Second))
	tm.Type("ixyz") // la tokenización de ~270 KB tarda mucho más que esto
	start := time.Now()
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool { return bytes.Contains(b, []byte("xyzfunc")) }, teatest.WithDuration(2*time.Second))
	if d := time.Since(start); d > time.Second {
		t.Errorf("la pulsación tardó %v en verse", d)
	}
	tm.Send(keyEsc)
	_ = tm.Quit()
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(10*time.Second)).(editor.Model)
	if fm.Buffer().Line(0)[:3] != "xyz" {
		t.Error("el texto escrito está en el buffer")
	}
}
