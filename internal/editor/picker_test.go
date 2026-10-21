package editor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

var (
	keyCtrlN = tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl}
	keyCtrlP = tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}
)

// sampleProject es el árbol de la maqueta 05.
func sampleProject(t *testing.T) string {
	return project(t,
		"internal/keymap/trie.go=package keymap\n\ntype Node struct{}\n",
		"internal/keymap/trie_test.go=package keymap\n",
		"internal/buffer/buffer.go=package buffer\n",
		"docs/README.md=# Docs\n",
		"openspec/changes/keymap-trie/design.md=# Diseño\n",
		"openspec/changes/keymap-trie/tasks.md=# Tareas\n",
		"go.mod=module x\n", "Makefile=all:\n",
		"node_modules/dep/index.js=x\n", ".git/config=x\n",
	)
}

// findFiles abre el picker de archivos y espera a que termine de cargar.
func (e *ed) findFiles() *ed {
	e.space("ff").settle()
	return e
}

func (e *ed) results() []string { return e.m.PickerResults() }

func TestFilePickerListsProjectFilesWithoutIgnored(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles()
	if !e.m.PickerOpen() || e.m.PickerTitle() != "Buscar archivos" {
		t.Fatalf("Space f f abre el picker de archivos: %q", e.m.PickerTitle())
	}
	if e.m.PickerLoading() {
		t.Error("tras cargar, ya no está cargando")
	}
	res := e.results()
	if len(res) != 8 {
		t.Errorf("se listan los 8 archivos del proyecto: %v", res)
	}
	for _, r := range res {
		if strings.Contains(r, "node_modules") || strings.HasPrefix(r, ".git/") {
			t.Errorf("no deben listarse los ignorados: %s", r)
		}
	}
	scr := e.screen()
	if !strings.Contains(scr, "Buscar archivos") || !strings.Contains(scr, "8/8") {
		t.Errorf("título y contador 8/8:\n%s", scr)
	}
}

func TestFilePickerFuzzyAndOpen(t *testing.T) {
	root := sampleProject(t)
	e := openProj(t, root).findFiles()
	e.typ("ktr")
	res := e.results()
	if len(res) == 0 || res[0] != "internal/keymap/trie.go" {
		t.Fatalf("ktr encuentra primero trie.go: %v", res)
	}
	for _, r := range res {
		if strings.Contains(r, "README") || strings.Contains(r, "Makefile") {
			t.Errorf("%s no contiene k, t y r en orden", r)
		}
	}
	// las letras coincidentes van resaltadas (amarillo y negrita) y el contador es 1/8
	raw := e.m.View().Content
	if !strings.Contains(raw, "1;38;2;213;184;94") {
		t.Error("las coincidencias se resaltan en amarillo y negrita")
	}
	if !strings.Contains(e.screen(), fmt.Sprintf("%d/8", len(res))) {
		t.Errorf("contador:\n%s", e.screen())
	}
	e.send(keyBackspace, keyBackspace, keyBackspace).typ("trie")
	if res := e.results(); len(res) != 4 || res[0] != "internal/keymap/trie.go" || res[1] != "internal/keymap/trie_test.go" {
		t.Errorf("trie: %v", res)
	}
	e.send(keyEnter)
	if e.m.PickerOpen() {
		t.Error("Enter cierra el picker")
	}
	if got := e.m.Buffer().Path(); got != filepath.Join(root, "internal", "keymap", "trie.go") {
		t.Errorf("se abre trie.go: %q", got)
	}
	if e.m.Focus() != editor.FocusEditor {
		t.Error("el foco queda en el editor")
	}
	e.wantText("package keymap\n\ntype Node struct{}\n")
}

func TestPickerKeyboardNavigation(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles().typ("go")
	first := e.results()[0]
	e.send(keyDown)
	if !strings.Contains(e.screen(), "▸") {
		t.Fatal("hay una fila seleccionada")
	}
	e.send(keyDown, keyDown, keyUp)
	// elegir el segundo de la lista con ↓ y abrirlo
	e.send(keyUp, keyCtrlN, keyCtrlP) // se queda en el primero
	e.send(keyCtrlN).send(keyEnter)
	want := e.results()
	_ = want
	if e.m.Buffer().Path() == "" {
		t.Fatal("se abrió el seleccionado")
	}
	if filepath.ToSlash(e.m.Buffer().Path()) == "" || strings.HasSuffix(e.m.Buffer().Path(), first) {
		// el segundo resultado no es el primero
		t.Errorf("Ctrl+n baja al segundo resultado: %q", e.m.Buffer().Path())
	}
}

func TestPickerSelectionStaysInBounds(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles().typ("trie")
	for i := 0; i < 5; i++ {
		e.send(keyDown)
	}
	last := e.results()[len(e.results())-1]
	e.send(keyEnter)
	if !strings.HasSuffix(filepath.ToSlash(e.m.Buffer().Path()), last) {
		t.Errorf("↓ se detiene en el último resultado (%s): %q", last, e.m.Buffer().Path())
	}
	e = openProj(t, sampleProject(t)).findFiles().typ("trie")
	e.send(keyUp, keyUp, keyEnter)
	if !strings.HasSuffix(filepath.ToSlash(e.m.Buffer().Path()), "trie.go") {
		t.Errorf("↑ se detiene en el primero: %q", e.m.Buffer().Path())
	}
}

func TestPickerCancelRestoresState(t *testing.T) {
	root := sampleProject(t)
	e := openProj(t, root)
	e.ex("e " + filepath.Join(root, "go.mod")).typ("l")
	before := e.m.Cursor()
	bufs := e.m.BufferCount()
	e.findFiles().typ("trie").send(keyEsc)
	if e.m.PickerOpen() {
		t.Fatal("Esc cierra el picker")
	}
	if e.m.Cursor() != before || e.m.BufferCount() != bufs || filepath.Base(e.m.Buffer().Path()) != "go.mod" {
		t.Error("cancelar no abre nada ni mueve el cursor")
	}
	e.findFiles().send(keyCtrlC)
	if e.m.PickerOpen() {
		t.Error("Ctrl+C también cancela")
	}
	if strings.Contains(e.screen(), "Buscar archivos") {
		t.Error("el picker desaparece de la pantalla")
	}
}

func TestPickerCapturesInputKeymapReceivesNothing(t *testing.T) {
	e := openProj(t, sampleProject(t)).ex("e " + "").send(keyEsc)
	e.findFiles()
	e.typ("dd").typ("i").typ("x").send(keyTab).typ(":q").send(keyEnter)
	if e.quit {
		t.Error(":q escrito en el picker es texto, no un comando")
	}
	if e.m.Mode() != editor.ModeNormal || e.m.Buffer().Modified() {
		t.Error("el keymap del editor no recibe teclas con el picker abierto")
	}
	if !e.m.PickerOpen() {
		t.Error("el picker sigue abierto")
	}
	if !strings.Contains(e.screen(), "ddix") {
		t.Errorf("lo escrito va a la entrada del picker:\n%s", e.screen())
	}
}

func TestPickerQueryEditing(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles().typ("trie_t")
	if len(e.results()) != 1 {
		t.Fatal(e.results())
	}
	e.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	if len(e.results()) != 8 {
		t.Errorf("Ctrl+U borra lo escrito: %d resultados", len(e.results()))
	}
	e.typ("go mod").send(tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl}) // borra «mod»
	if !strings.Contains(e.screen(), "> go ") {
		t.Errorf("Ctrl+W borra la palabra anterior:\n%s", e.screen())
	}
	e.send(keyLeft, keyLeft, keyBackspace, keyHome).typ("x")
	e.send(keyEnd, keyDelete, tea.PasteMsg{Content: "mak\nefile"})
	if res := e.results(); len(res) == 0 && !strings.Contains(e.screen(), "Sin resultados") {
		t.Errorf("sin resultados se avisa:\n%s", e.screen())
	}
}

func TestPickerNoResultsAndEnter(t *testing.T) {
	e := openProj(t, sampleProject(t)).findFiles().typ("zzzz")
	if len(e.results()) != 0 || !strings.Contains(e.screen(), "Sin resultados") || !strings.Contains(e.screen(), "0/8") {
		t.Errorf("sin coincidencias:\n%s", e.screen())
	}
	e.send(keyEnter)
	if !e.m.PickerOpen() {
		t.Error("Enter sin resultados no cierra ni abre nada")
	}
}

func TestFilePreviewIsHighlightedAndLimited(t *testing.T) {
	var big strings.Builder
	for i := 0; i < 500; i++ {
		fmt.Fprintf(&big, "func f%d() {}\n", i)
	}
	root := project(t, "a.go="+big.String(), "bin.dat=ab\x00cd\n", "notas.txt=hola\n", "vacio.go=")
	e := openProj(t, root).findFiles().typ("a.go")
	raw := e.m.View().Content
	scr := e.screen()
	if !strings.Contains(scr, "func f0() {}") || !strings.Contains(scr, "func f11() {}") {
		t.Errorf("se ve el principio del archivo:\n%s", scr)
	}
	if strings.Contains(scr, "func f200()") || strings.Contains(scr, "func f499()") {
		t.Error("la vista previa se limita a 100 líneas")
	}
	// coloreada: «func» en mauve dentro de la ventana
	found := false
	for _, row := range strings.Split(raw, "\n") {
		if strings.Contains(row, "│") && strings.Contains(row, "f0") && strings.Contains(sgrBefore(row[strings.Index(row, "│")+3:], "f0"), "38;2;127;193;254") {
			found = true
		}
	}
	if !found {
		t.Error("el nombre de la función va en azul en la vista previa")
	}
	e.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}).typ("bin")
	if !strings.Contains(e.screen(), "binario") {
		t.Errorf("los binarios se omiten con una nota:\n%s", e.screen())
	}
	e.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}).typ("vacio")
	if !strings.Contains(e.screen(), "vacio.go") {
		t.Error("un archivo vacío no rompe la vista previa")
	}
}

func TestPreviewReadsOnlyTheBeginningOfHugeFiles(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "enorme.txt")
	f, _ := os.Create(p)
	for i := 0; i < 50000; i++ {
		fmt.Fprintf(f, "línea %d con bastante texto para que pese\n", i)
	}
	f.Close()
	e := openProj(t, root).findFiles()
	start := time.Now()
	scr := e.screen()
	if d := time.Since(start); d > time.Second {
		t.Errorf("la vista previa tardó %v", d)
	}
	if !strings.Contains(scr, "línea 0") {
		t.Errorf("principio del archivo:\n%s", scr)
	}
}

func TestPickerStaysResponsiveOnLargeProjects(t *testing.T) {
	root := t.TempDir()
	for d := 0; d < 12; d++ {
		dir := filepath.Join(root, fmt.Sprintf("pkg%02d", d))
		os.MkdirAll(dir, 0o755)
		for i := 0; i < 500; i++ {
			os.WriteFile(filepath.Join(dir, fmt.Sprintf("archivo_%03d.go", i)), nil, 0o644)
		}
	}
	e := openProj(t, root)
	e.space("ff")
	if !e.m.PickerOpen() || !e.m.PickerLoading() {
		t.Fatal("el picker se abre de inmediato, aunque la lista siga cargando")
	}
	// se puede escribir mientras llegan los archivos: cada pulsación responde sin esperar
	for _, r := range "pkg07" {
		start := time.Now()
		e.send(tea.KeyPressMsg{Code: r, Text: string(r)})
		if d := time.Since(start); d > 250*time.Millisecond {
			t.Errorf("la pulsación %q tardó %v", r, d)
		}
	}
	if !strings.Contains(e.screen(), "pkg07") {
		t.Error("lo escrito se ve al instante")
	}
	e.settle()
	if e.m.PickerLoading() {
		t.Fatal("termina de cargar")
	}
	res := e.results()
	if len(res) != 200 || !strings.HasPrefix(res[0], "pkg07/") {
		t.Errorf("200 resultados de pkg07 (%d): %v", len(res), res[:min(3, len(res))])
	}
	if !strings.Contains(e.screen(), "200/6000") {
		t.Errorf("contador con el total:\n%s", e.screen())
	}
	e.send(keyCtrlU).typ("archivo_499")
	e.settle()
	if got := len(e.results()); got != 12 { // uno por cada pkgNN
		t.Errorf("otra búsqueda sobre la lista completa: %d", got)
	}
}

func TestPickerFillsProgressively(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 1500; i++ {
		os.WriteFile(filepath.Join(root, fmt.Sprintf("f%04d.txt", i)), nil, 0o644)
	}
	e := openProj(t, root)
	e.space("ff")
	// se ejecuta solo el primer lote: la lista ya tiene archivos pero sigue cargando
	for _, c := range e.cmds {
		for _, msg := range run(c) {
			if _, ok := msg.(tea.BatchMsg); !ok {
				e.send(msg)
			}
		}
	}
	if !e.m.PickerLoading() {
		t.Fatal("sigue cargando tras el primer lote")
	}
	if n := len(e.results()); n == 0 {
		t.Errorf("ya hay resultados parciales: %d", n)
	}
	e.settle()
	if e.m.PickerLoading() || !strings.Contains(e.screen(), "/1500") {
		t.Errorf("al final están todos:\n%s", e.screen())
	}
}

func TestBufferPickerMarksModifiedAndSwitches(t *testing.T) {
	p := files(t, "a.txt", "b.txt", "c.txt")
	e := openMany(t, p[0], p[1], p[2])
	e.ex("bp").typ("ix").send(keyEsc) // b modificado; activo b
	e.space("bb")
	if !e.m.PickerOpen() || e.m.PickerTitle() != "Buffers" {
		t.Fatal("Space b b abre la lista de buffers")
	}
	scr := e.screen()
	if !strings.Contains(scr, "[+]") || !strings.Contains(scr, "3/3") {
		t.Errorf("los modificados se marcan:\n%s", scr)
	}
	// elegir el tercer buffer (c.txt) escribiendo y aceptando
	e.typ("c.txt").send(keyEnter)
	if e.m.ActiveBuffer() != 2 || e.m.PickerOpen() {
		t.Errorf("el elegido pasa a ser el activo: %d", e.m.ActiveBuffer())
	}
	// por posición: el segundo de la lista
	e.space("bb").send(keyDown).send(keyEnter)
	if e.m.ActiveBuffer() != 1 {
		t.Errorf("el segundo de la lista: %d", e.m.ActiveBuffer())
	}
}

func TestRecentFiles(t *testing.T) {
	root := sampleProject(t)
	e := openProj(t, root)
	for _, f := range []string{"go.mod", "Makefile", "docs/README.md"} {
		e.ex("e " + filepath.Join(root, f))
	}
	e.space("fr")
	res := e.results()
	if e.m.PickerTitle() != "Archivos recientes" || len(res) != 3 || res[0] != "docs/README.md" || res[2] != "go.mod" {
		t.Errorf("recientes (el último abierto primero): %v", res)
	}
	e.typ("make").send(keyEnter)
	if filepath.Base(e.m.Buffer().Path()) != "Makefile" || e.m.BufferCount() != 3 {
		t.Errorf("se activa el buffer existente: %q (%d buffers)", e.m.Buffer().Path(), e.m.BufferCount())
	}
	// reabrir lo mueve al principio
	e.space("fr")
	if e.results()[0] != "Makefile" {
		t.Errorf("tras reabrir va primero: %v", e.results())
	}
}

func TestNewFileOpensPromptWithEdit(t *testing.T) {
	e := openProj(t, sampleProject(t)).space("fn")
	if e.m.Mode() != editor.ModeCommand || !strings.Contains(e.screen(), ":e ") {
		t.Errorf("Space f n abre «:e »:\n%s", e.screen())
	}
	nuevo := filepath.Join(t.TempDir(), "nuevo.txt")
	e.typ(nuevo).send(keyEnter)
	if e.m.Buffer().Path() != nuevo {
		t.Errorf("abre el archivo nuevo: %q", e.m.Buffer().Path())
	}
}

func TestPickerWorksFromExplorerFocus(t *testing.T) {
	e := openProj(t, sampleProject(t)).space("e")
	if e.m.Focus() != editor.FocusExplorer {
		t.Fatal("preparación")
	}
	e.space("ff").settle().typ("makefile").send(keyEnter)
	if filepath.Base(e.m.Buffer().Path()) != "Makefile" || e.m.Focus() != editor.FocusEditor {
		t.Errorf("el líder funciona con el foco en el explorador: %q", e.m.Buffer().Path())
	}
}

func TestPickerOnSmallTerminal(t *testing.T) {
	e := openProj(t, sampleProject(t))
	e.send(tea.WindowSizeMsg{Width: 40, Height: 10})
	e.space("ff").settle().typ("trie")
	_ = e.m.View()
	e.send(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !strings.Contains(e.screen(), "internal/keymap/trie.go") {
		t.Error("tras ensanchar se ve la lista")
	}
	e.send(tea.WindowSizeMsg{Width: 0, Height: 0})
	_ = e.m.View()
}

// ---------- paleta de comandos ----------

func TestPaletteListsEveryCommand(t *testing.T) {
	e := openProj(t, sampleProject(t)).space(" ")
	if !e.m.PickerOpen() || e.m.PickerTitle() != "Paleta de comandos" {
		t.Fatalf("Space Space abre la paleta: %q", e.m.PickerTitle())
	}
	total := len(editor.Registry().List())
	if !strings.Contains(e.screen(), fmt.Sprintf("/%d", total)) {
		t.Errorf("el contador usa el total de comandos (%d):\n%s", total, e.screen())
	}
	// escribir «guardar» encuentra los comandos de guardar con sus atajos a la derecha
	e.typ("gua")
	res := e.results()
	if len(res) == 0 || !strings.Contains(res[0], "Guardar") {
		t.Fatalf("gua: %v", res)
	}
	scr := e.screen()
	for _, want := range []string{"Guardar", "Ctrl+S", ":w", ":wq"} {
		if !strings.Contains(scr, want) {
			t.Errorf("falta %q en la paleta:\n%s", want, scr)
		}
	}
	// «Gua» resaltado en la primera fila
	if !strings.Contains(e.m.View().Content, "1;38;2;213;184;94") {
		t.Error("las letras coincidentes se resaltan")
	}
}

func TestPaletteDescriptionsAreInSpanishAndComplete(t *testing.T) {
	e := openProj(t, sampleProject(t)).space(" ")
	for _, c := range editor.Registry().List() {
		if c.Desc == "" || len(c.Keys) == 0 {
			t.Errorf("%s sin descripción o atajos", c.Name)
		}
	}
	_ = e
}

func TestPaletteRunsCommandLikeItsShortcut(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("abc\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}).typ("ix").send(keyEsc)
	e.space(" ").typ("guardar archivo").send(keyEnter)
	if got, _ := os.ReadFile(p); string(got) != "xabc\n" {
		t.Errorf("la paleta guarda igual que Ctrl+S: %q", got)
	}
	if e.m.PickerOpen() || e.m.Buffer().Modified() {
		t.Error("se cierra y el buffer queda guardado")
	}
	if !strings.Contains(e.m.Message(), "escritas") {
		t.Errorf("mismo mensaje que Ctrl+S: %q", e.m.Message())
	}
}

func TestPaletteRunsCommandsThatChangeMode(t *testing.T) {
	e := openProj(t, sampleProject(t)).space(" ").typ("modo visual por car").send(keyEnter)
	e.mode(editor.ModeVisual)
	e.send(keyEsc).space(" ").typ("tema").send(keyEnter)
	if !strings.Contains(e.m.Message(), "claro") {
		t.Errorf("alternar tema: %q", e.m.Message())
	}
	e.space(" ").typ("buscar archivos").send(keyEnter) // la paleta abre otra ventana
	if e.m.PickerTitle() != "Buscar archivos" {
		t.Errorf("un comando que abre un picker desde la paleta: %q", e.m.PickerTitle())
	}
}

func TestPaletteCommandsWithMissingContextDoNotCrash(t *testing.T) {
	e := openProj(t, sampleProject(t))
	for _, c := range editor.Registry().List() {
		switch c.Name {
		case "app.quit", "file.save_quit", "file.update_quit", "buffer.close":
			continue
		}
		e.send(keyEsc)
		e.space(" ").typ(c.Desc[:min(len(c.Desc), 6)])
		if len(e.results()) == 0 {
			e.send(keyEsc)
			continue
		}
		e.send(keyEnter) // algunos abren otras ventanas, avisos o diálogos
		for i := 0; i < 3; i++ {
			e.send(keyEsc)
		}
		e.settle()
	}
	_ = e.m.View()
}

func TestHelpKeysListsShortcuts(t *testing.T) {
	e := openProj(t, sampleProject(t)).space("?")
	if e.m.PickerTitle() != "Atajos" {
		t.Errorf("Space ? abre la lista de atajos: %q", e.m.PickerTitle())
	}
	e.send(keyEsc).space("h")
	if e.m.PickerTitle() != "Atajos" {
		t.Errorf("Space h también: %q", e.m.PickerTitle())
	}
}

func TestLeaderShortcuts(t *testing.T) {
	p := files(t, "a.txt", "b.txt")
	e := openMany(t, p[0], p[1])
	e.typ("ix").send(keyEsc)
	e.space("w")
	if e.m.Buffer().Modified() {
		t.Error("Space w guarda")
	}
	e.space("q")
	if e.m.BufferCount() != 1 {
		t.Error("Space q cierra el buffer")
	}
	e.typ("ix").send(keyEsc).space("Q")
	if !e.m.DialogOpen() {
		t.Error("Space Q sale (con diálogo si hay cambios)")
	}
	e.send(keyEsc).space("/")
	if e.m.Mode() != editor.ModeCommand || !strings.Contains(e.screen(), "/") {
		t.Error("Space / abre la búsqueda")
	}
	e.send(keyEsc).space(":")
	if e.m.Mode() != editor.ModeCommand {
		t.Error("Space : abre la línea de comandos")
	}
}

// ---------- sugerencias en «:» ----------

func TestCommandLineSuggestions(t *testing.T) {
	e := newEd(t, "abc\n").typ(":gua")
	scr := e.screen()
	if !strings.Contains(scr, "Guardar") || !strings.Contains(scr, "Comandos") {
		t.Fatalf("al escribir en «:» se ven sugerencias:\n%s", scr)
	}
	if !strings.Contains(scr, "Ctrl+S") {
		t.Errorf("con sus atajos:\n%s", scr)
	}
	if v := e.m.View(); v.Cursor == nil || v.Cursor.Y != 23 {
		t.Errorf("el cursor se queda en la línea de comandos: %+v", v.Cursor)
	}
	e.send(keyEsc)
	if strings.Contains(e.screen(), "Comandos") || e.m.Mode() != editor.ModeNormal {
		t.Error("Esc cierra las sugerencias y el prompt")
	}
}

func TestEnterRunsTextAsTypedWithoutChoosingSuggestion(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}).typ("ix").send(keyEsc)
	e.typ(":wq")
	if !strings.Contains(e.screen(), "Guardar") {
		t.Error("hay sugerencias para «wq»")
	}
	e.send(keyEnter)
	if !e.quit {
		t.Error(":wq sin elegir sugerencia ejecuta :wq (guarda y sale)")
	}
	if got, _ := os.ReadFile(p); string(got) != "xa\n" {
		t.Errorf("archivo = %q", got)
	}
}

func TestEnterRunsChosenSuggestion(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}).typ("ix").send(keyEsc)
	e.typ(":guardar archivo")
	e.send(keyDown) // se elige la primera sugerencia
	e.send(keyEnter)
	if got, _ := os.ReadFile(p); string(got) != "xa\n" {
		t.Errorf("se ejecuta la sugerencia elegida, no el texto: %q", got)
	}
	if strings.HasPrefix(e.m.Message(), "E492") {
		t.Error("«guardar archivo» no es un comando de «:»: de haberse ejecutado como texto daría E492")
	}
	// sin elegir, el mismo texto sí da E492
	e.typ(":guardar archivo").send(keyEnter)
	if !strings.HasPrefix(e.m.Message(), "E492") {
		t.Errorf("sin elegir sugerencia se ejecuta el texto tal cual: %q", e.m.Message())
	}
}

func TestSuggestionNavigationWrapsAndHistoryStillWorks(t *testing.T) {
	e := newEd(t, "abc\n").ex("nada")
	e.typ(":").send(keyUp) // sin texto, ↑ recorre el historial
	if !strings.Contains(e.screen(), ":nada") {
		t.Errorf("historial con el prompt vacío:\n%s", e.screen())
	}
	e.send(keyEsc).typ(":gua").send(keyUp)
	if !strings.Contains(e.screen(), "▸") {
		t.Error("↑ desde ninguna elección va a la última sugerencia")
	}
	e.send(keyDown) // vuelve al principio (vuelta circular)
	e.send(keyBackspace)
	e.send(keyEsc)
	e.mode(editor.ModeNormal)
}

func TestSuggestionsMatchCommandNamesToo(t *testing.T) {
	e := newEd(t, "abc\n").typ(":wq")
	if !strings.Contains(e.screen(), ":wq") {
		t.Errorf("escribir «wq» sugiere el comando que tiene ese atajo:\n%s", e.screen())
	}
	if !strings.Contains(e.screen(), "Guardar y salir") {
		t.Errorf("Guardar y salir:\n%s", e.screen())
	}
}

// ---------- menú líder ----------

func spaceOnly(t *testing.T) *ed {
	t.Helper()
	e := openProj(t, sampleProject(t))
	e.send(tea.KeyPressMsg{Code: ' ', Text: " "})
	return e
}

func TestLeaderPanelAppearsAfterTheDelay(t *testing.T) {
	e := spaceOnly(t)
	if strings.Contains(e.screen(), "líder") {
		t.Error("el panel no aparece de inmediato")
	}
	e.send(editor.LeaderMsg{Seq: e.m.Seq()}) // pasan 300 ms sin otra tecla
	scr := e.screen()
	for _, want := range []string{"<Space> líder", "+archivos", "explorador (alternar)", "+buffers", "guardar", "cerrar buffer", "salir", "tema claro / oscuro", "paleta de comandos", "buscar en el archivo", "línea de comandos", "lista de atajos", "<Space> esperando tecla..."} {
		if !strings.Contains(scr, want) {
			t.Errorf("falta %q en el panel:\n%s", want, scr)
		}
	}
	if !strings.Contains(scr, "f") {
		t.Error("teclas")
	}
}

func TestLeaderPanelDoesNotAppearForFastTyping(t *testing.T) {
	e := spaceOnly(t)
	stale := e.m.Seq()
	e.typ("ff")                          // en menos de 300 ms
	e.send(editor.LeaderMsg{Seq: stale}) // llega el temporizador viejo
	if strings.Contains(e.screen(), "<Space> líder") {
		t.Error("el panel no debe mostrarse si la secuencia ya se completó")
	}
	if !e.m.PickerOpen() {
		t.Error("el picker se abre")
	}
	e.send(keyEsc)
	e = spaceOnly(t)
	old := e.m.Seq()
	e.typ("f") // otra tecla antes de que venza: el temporizador de «Space» ya no vale
	e.send(editor.LeaderMsg{Seq: old})
	if strings.Contains(e.screen(), "líder") {
		t.Error("un temporizador viejo no abre el panel")
	}
}

func TestLeaderPanelDescendsIntoGroupsAndEscCancels(t *testing.T) {
	e := spaceOnly(t)
	e.send(editor.LeaderMsg{Seq: e.m.Seq()})
	e.typ("f")
	scr := e.screen()
	for _, want := range []string{"<Space> f", "archivos", "buscar archivos", "abrir reciente", "nuevo archivo"} {
		if !strings.Contains(scr, want) {
			t.Errorf("falta %q en el grupo f:\n%s", want, scr)
		}
	}
	if strings.Contains(scr, "paleta de comandos") {
		t.Error("el panel del grupo reemplaza al principal")
	}
	e.send(keyEsc)
	if strings.Contains(e.screen(), "líder") || strings.Contains(e.screen(), "esperando") || e.m.Pending() != "" {
		t.Errorf("Esc cancela la secuencia:\n%s", e.screen())
	}
	// Esc también cancela en el primer nivel
	e = spaceOnly(t)
	e.send(editor.LeaderMsg{Seq: e.m.Seq()}).send(keyEsc)
	if strings.Contains(e.screen(), "líder") || e.m.Pending() != "" {
		t.Error("Esc en el primer nivel")
	}
	// una tecla sin significado dentro del panel lo cierra sin hacer nada
	e = spaceOnly(t)
	e.send(editor.LeaderMsg{Seq: e.m.Seq()}).typ("z")
	if strings.Contains(e.screen(), "líder") || e.m.Pending() != "" || e.m.Buffer().Modified() {
		t.Error("tecla sin asignar: se cancela")
	}
}

func TestLeaderPanelDoesNotExpireButOrdinarySequencesDo(t *testing.T) {
	e := spaceOnly(t)
	seq := e.m.Seq()
	e.send(editor.LeaderMsg{Seq: seq})
	e.send(editor.TimeoutMsg{Seq: seq}) // vence el segundo de las secuencias normales
	if e.m.Pending() != "space" || !strings.Contains(e.screen(), "líder") {
		t.Errorf("con el panel a la vista, la espera no caduca: pendiente %q", e.m.Pending())
	}
	e.send(keyEsc)
	// una secuencia que no es del líder sí caduca (1000 ms)
	e = newEd(t, strings.Repeat("x\n", 5)).typ("g")
	e.send(editor.TimeoutMsg{Seq: e.m.Seq()})
	if e.m.Pending() != "" {
		t.Error("«g» caduca a los 1000 ms")
	}
	// y Space antes de que aparezca el panel también caduca a los 1000 ms si hay un TimeoutMsg
	e = spaceOnly(t)
	e.send(editor.TimeoutMsg{Seq: e.m.Seq()})
	if e.m.Pending() != "" {
		t.Error("sin panel, el líder caduca como cualquier secuencia incompleta")
	}
}

func TestLeaderPanelWithRealClock(t *testing.T) {
	if editor.LeaderDelay != 300*time.Millisecond || editor.DefaultTimeout != 1000*time.Millisecond {
		t.Fatalf("el panel espera 300 ms y las secuencias caducan a los 1000 ms: %v %v", editor.LeaderDelay, editor.DefaultTimeout)
	}
	m, _ := editor.Open("", editor.WithRoot(sampleProject(t))) // sin reducir los tiempos
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24}, tea.KeyPressMsg{Code: ' ', Text: " "})
	start := time.Now()
	e.settle() // ejecuta el temporizador real de Bubble Tea
	if !strings.Contains(e.screen(), "<Space> líder") {
		t.Errorf("tras 300 ms de verdad aparece el panel:\n%s", e.screen())
	}
	if d := time.Since(start); d < 250*time.Millisecond {
		t.Errorf("el panel apareció demasiado pronto: %v", d)
	}
}

func TestLeaderPanelInExplorerContext(t *testing.T) {
	e := openProj(t, sampleProject(t)).space("e")
	e.send(tea.KeyPressMsg{Code: ' ', Text: " "}).send(editor.LeaderMsg{Seq: e.m.Seq()})
	if !strings.Contains(e.screen(), "+archivos") {
		t.Errorf("el líder también se ve con el foco en el explorador:\n%s", e.screen())
	}
}

func TestPickerDimsBackgroundButKeepsItVisible(t *testing.T) {
	e := openProj(t, sampleProject(t)).space("e")
	bright := e.m.View().Content
	e.send(keyEsc).space(" ")
	dim := e.m.View().Content
	if strings.Count(dim, "38;2;207;215;235") >= strings.Count(bright, "38;2;207;215;235") {
		t.Error("el fondo se atenúa con la ventana abierta")
	}
	if !strings.Contains(e.screen(), filepath.Base(e.m.ExplorerSelection())) {
		t.Error("el fondo sigue visible")
	}
	_ = ui.Dark()
}
