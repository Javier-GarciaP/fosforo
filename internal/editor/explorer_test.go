package editor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// project crea archivos de prueba (un nombre terminado en «/» es una carpeta) y devuelve la raíz.
func project(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		name, content, _ := strings.Cut(f, "=")
		p := filepath.Join(root, name)
		if strings.HasSuffix(name, "/") {
			os.MkdirAll(p, 0o755)
			continue
		}
		os.MkdirAll(filepath.Dir(p), 0o755)
		if content == "" {
			content = "contenido de " + name + "\n"
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// openProj crea un editor de 100x24 con el explorador apuntando a root (sin iconos, para leer la estructura).
func openProj(t *testing.T, root string, opts ...editor.Option) *ed {
	t.Helper()
	opts = append([]editor.Option{editor.WithRoot(root), editor.WithIcons(ui.IconsNone)}, opts...)
	m, err := editor.Open("", opts...)
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24})
	return e
}

// panel devuelve las filas de texto del panel del explorador (24 columnas) sin espacios finales.
func (e *ed) panel() []string {
	var out []string
	for _, row := range strings.Split(e.screen(), "\n") {
		r := []rune(row)
		if len(r) < 25 {
			out = append(out, "")
			continue
		}
		out = append(out, strings.TrimRight(string(r[:24]), " "))
	}
	return out
}

func (e *ed) panelHas(want string) bool {
	for _, r := range e.panel() {
		if strings.TrimSpace(r) == want {
			return true
		}
	}
	return false
}

func (e *ed) selected() string { return filepath.Base(e.m.ExplorerSelection()) }

func (e *ed) space(keys string) *ed { return e.typ(" " + keys) }

func TestToggleExplorerWithSpaceE(t *testing.T) {
	root := project(t, "main.go", "docs/README.md")
	e := openProj(t, root)
	if e.m.ExplorerVisible() {
		t.Fatal("empieza oculto")
	}
	e.space("e")
	if !e.m.ExplorerVisible() || e.m.Focus() != editor.FocusExplorer {
		t.Fatal("Space e lo muestra y le da el foco")
	}
	rows := strings.Split(e.screen(), "\n")
	for i := 0; i < 22; i++ {
		if r := []rune(rows[i]); len(r) < 25 || r[24] != '│' {
			t.Fatalf("fila %d: el panel mide 24 columnas y un separador: %q", i, rows[i])
		}
	}
	if !strings.Contains(e.screen(), "EXPLORER") {
		t.Error("la statusline indica que el foco está en el explorador")
	}
	e.space("e")
	if e.m.ExplorerVisible() || e.m.Focus() != editor.FocusEditor {
		t.Error("Space e lo oculta y devuelve el foco al editor")
	}
	if strings.Contains(e.screen(), "main.go") {
		t.Error("oculto, no se ve")
	}
}

func TestFocusToggleWithTabAndCtrlW(t *testing.T) {
	e := openProj(t, project(t, "a.txt")).space("e")
	for _, k := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: 'w', Mod: tea.ModCtrl}} {
		e.send(k)
		if e.m.Focus() != editor.FocusEditor || !e.m.ExplorerVisible() {
			t.Fatalf("%v: el foco pasa al editor sin cerrar el explorador", k)
		}
		if !strings.Contains(e.screen(), "NORMAL") {
			t.Error("con el foco en el editor la statusline vuelve a NORMAL")
		}
		e.send(k)
		if e.m.Focus() != editor.FocusExplorer {
			t.Fatalf("%v: vuelve al explorador", k)
		}
	}
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}).space("e") // oculto
	e.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(e.m.Message(), "oculto") {
		t.Errorf("con el explorador oculto Tab avisa: %q", e.m.Message())
	}
}

func TestTreeDisplayAndIndentation(t *testing.T) {
	root := project(t, "internal/keymap/trie.go", "internal/buffer/b.go", "docs/README.md", "Makefile", "go.mod", ".git/config", "node_modules/x/y.js")
	e := openProj(t, root).space("e")
	rows := e.panel()
	// raíz y primer nivel: carpetas primero (atenuadas, con «/»), luego archivos por orden alfabético
	want := []string{" " + filepath.Base(root), "   docs/", "   internal/", "   go.mod", "   Makefile"} // 2 columnas por nivel, con 1 de margen
	for i, w := range want {
		if !strings.HasPrefix(rows[i], w) {
			t.Errorf("fila %d = %q, quería %q\n%s", i, rows[i], w, strings.Join(rows[:8], "\n"))
		}
	}
	if e.panelHas(".git/") || e.panelHas("node_modules/") {
		t.Error(".git y node_modules no se muestran")
	}
	e.typ("jj").typ("l") // expandir internal
	e.typ("jjl")         // bajar a keymap y expandirlo
	e.typ("j")
	rows = e.panel()
	// trie.go está 3 niveles por debajo de la raíz: 6 columnas más a la derecha que ella
	col := func(name string) int {
		for _, r := range rows {
			if strings.TrimSpace(r) == name {
				return len(r) - len(strings.TrimLeft(r, " "))
			}
		}
		t.Fatalf("no se ve %q:\n%s", name, strings.Join(rows[:9], "\n"))
		return 0
	}
	rootCol := col(filepath.Base(root))
	if got := col("internal/") - rootCol; got != 2 {
		t.Errorf("sangría de internal/ = %d, quería 2", got)
	}
	if got := col("keymap/") - rootCol; got != 4 {
		t.Errorf("sangría de keymap/ = %d, quería 4", got)
	}
	if got := col("trie.go") - rootCol; got != 6 {
		t.Errorf("sangría de trie.go = %d columnas bajo la raíz, quería 6", got)
	}
}

func TestTreeHasNoArrowsOrGuideLines(t *testing.T) {
	e := openProj(t, project(t, "a/b/c.txt", "d.txt"), editor.WithIcons(ui.IconsNerd)).space("e")
	e.typ("l").typ("jl").typ("jl")
	for _, row := range e.panel() {
		if strings.ContainsAny(row, "▸▾▶▼▷▽►▼├└─│┬┴") {
			t.Errorf("el árbol no usa flechas ni líneas guía: %q", row)
		}
	}
}

func TestIconsNerdAndNone(t *testing.T) {
	root := project(t, "carpeta/x.go", "main.go", "Makefile", "notas.xyz")
	panelRunes := func(e *ed) string { return strings.Join(e.panel(), "\n") }
	hasPrivate := func(s string) bool {
		for _, r := range s {
			if r >= 0xE000 && r <= 0xF8FF || r >= 0xF0000 {
				return true
			}
		}
		return false
	}

	e := openProj(t, root, editor.WithIcons(ui.IconsNerd)).space("e")
	nerd := panelRunes(e)
	for _, glyph := range []string{"\U000f0256", "\ue627", "\U000f0224"} { // carpeta cerrada, Go, archivo por defecto
		if !strings.Contains(nerd, glyph) {
			t.Errorf("falta el icono %U", []rune(glyph)[0])
		}
	}
	e.typ("jl") // expande carpeta: pasa a icono de carpeta abierta
	if !strings.Contains(panelRunes(e), "\U000f0dcf") {
		t.Error("la carpeta expandida usa el icono de carpeta abierta")
	}

	e = openProj(t, root, editor.WithIcons(ui.IconsNone)).space("e")
	if txt := panelRunes(e); hasPrivate(txt) {
		t.Errorf("con icons=none no se imprime ningún icono:\n%s", txt)
	}
	if !e.panelHas("carpeta/") || !e.panelHas("main.go") {
		t.Errorf("sin iconos las carpetas terminan en «/»:\n%s", strings.Join(e.panel(), "\n"))
	}
}

func TestExplorerNavigationKeys(t *testing.T) {
	root := project(t, "a.txt", "b.txt", "c.txt", "d.txt")
	e := openProj(t, root).space("e")
	if e.selected() != filepath.Base(root) {
		t.Fatalf("empieza en la raíz: %s", e.selected())
	}
	e.typ("j")
	if e.selected() != "a.txt" {
		t.Errorf("j: %s", e.selected())
	}
	e.typ("2j")
	if e.selected() != "c.txt" {
		t.Errorf("2j: %s", e.selected())
	}
	e.typ("k")
	if e.selected() != "b.txt" {
		t.Errorf("k: %s", e.selected())
	}
	e.typ("G")
	if e.selected() != "d.txt" {
		t.Errorf("G: %s", e.selected())
	}
	e.typ("j")
	if e.selected() != "d.txt" {
		t.Errorf("j en la última fila: %s", e.selected())
	}
	e.typ("gg")
	if e.selected() != filepath.Base(root) {
		t.Errorf("gg: %s", e.selected())
	}
	e.typ("k")
	if e.selected() != filepath.Base(root) {
		t.Errorf("k en la primera fila: %s", e.selected())
	}
	e.send(keyDown, keyDown, keyUp)
	if e.selected() != "a.txt" {
		t.Errorf("flechas: %s", e.selected())
	}
}

func TestExpandCollapseAndParent(t *testing.T) {
	root := project(t, "d/e/f.txt", "d/g.txt", "z.txt")
	e := openProj(t, root).space("e")
	e.typ("j") // d
	if e.selected() != "d" {
		t.Fatal(e.selected())
	}
	e.typ("l")
	if !e.panelHas("e/") || !e.panelHas("g.txt") {
		t.Errorf("l expande:\n%s", strings.Join(e.panel(), "\n"))
	}
	e.typ("l") // ya expandido: baja al primer hijo
	if e.selected() != "e" {
		t.Errorf("l sobre una carpeta abierta baja al primer hijo: %s", e.selected())
	}
	e.typ("h") // e está cerrada: sube al padre
	if e.selected() != "d" {
		t.Errorf("h sobre una carpeta cerrada sube al padre: %s", e.selected())
	}
	e.typ("h") // d abierta: se colapsa
	if e.panelHas("g.txt") || e.selected() != "d" {
		t.Errorf("h colapsa:\n%s", strings.Join(e.panel(), "\n"))
	}
	e.typ("l").typ("jj") // baja a g.txt
	if e.selected() != "g.txt" {
		t.Fatal(e.selected())
	}
	e.typ("h")
	if e.selected() != "d" {
		t.Errorf("h sobre un archivo sube a su carpeta: %s", e.selected())
	}
	e.typ("gg").typ("h")
	if e.selected() != filepath.Base(root) || !e.panelHas("z.txt") {
		t.Error("la raíz no se colapsa")
	}
	e.send(tea.KeyPressMsg{Code: tea.KeyRight}).send(tea.KeyPressMsg{Code: tea.KeyLeft})
}

func TestOpenFileFromExplorer(t *testing.T) {
	root := project(t, "a.txt=uno\ndos\n", "b.txt=tres\n")
	e := openProj(t, root).space("e").typ("j")
	e.send(keyEnter)
	if e.m.Focus() != editor.FocusEditor {
		t.Error("tras abrir, el foco pasa al editor")
	}
	e.wantText("uno\ndos\n")
	if e.m.BufferCount() != 1 {
		t.Errorf("el buffer vacío inicial se sustituye: %d buffers", e.m.BufferCount())
	}
	if !strings.Contains(e.screen(), "a.txt") || !strings.Contains(e.screen(), "uno") {
		t.Errorf("pantalla:\n%s", e.screen())
	}
	// el editor recibe las teclas: mover y borrar
	e.typ("jx")
	e.wantText("uno\nos\n")
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}).typ("jl") // otro archivo
	e.send(keyEnter)
	e.wantText("tres\n")
	if e.m.BufferCount() != 2 {
		t.Fatalf("buffers = %d", e.m.BufferCount())
	}
}

func TestExplorerShowsCurrentFileOnOpen(t *testing.T) {
	root := project(t, "internal/keymap/trie.go", "internal/buffer/b.go", "docs/README.md")
	e := openProj(t, root)
	e.ex("e " + filepath.Join(root, "internal", "keymap", "trie.go"))
	e.space("e")
	if e.selected() != "trie.go" {
		t.Fatalf("se selecciona el archivo actual: %s", e.selected())
	}
	if !e.panelHas("keymap/") || !e.panelHas("buffer/") || e.panelHas("README.md") {
		t.Errorf("se expanden solo las carpetas padre:\n%s", strings.Join(e.panel(), "\n"))
	}
	// la fila seleccionada lleva el fondo de selección
	var row string
	for _, r := range strings.Split(e.m.View().Content, "\n") {
		if strings.Contains(r, "trie.go") {
			row = r
			break
		}
	}
	if !strings.Contains(row, "48;2;35;44;66") {
		t.Errorf("la fila seleccionada debe resaltarse: %q", row)
	}
}

func TestEditorKeysDoNotReachExplorerAndViceVersa(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e")
	e.typ("j").typ("d") // en el explorador «d» pide confirmar el borrado: no es un operador
	if !e.m.DialogOpen() {
		t.Error("d en el explorador abre la confirmación de borrado")
	}
	e.send(keyEsc)
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}).typ("ix").send(keyEsc)
	e.wantText("x\n")
}

func TestCreateFileAndFolder(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e")
	e.typ("a")
	e.mode(editor.ModeCommand)
	if !strings.Contains(e.screen(), "Nuevo: ") {
		t.Errorf("el prompt de creación:\n%s", e.screen())
	}
	e.typ("util/strings.go").send(keyEnter)
	p := filepath.Join(root, "util", "strings.go")
	if info, err := os.Stat(p); err != nil || info.IsDir() {
		t.Fatalf("el archivo se crea con sus carpetas intermedias: %v", err)
	}
	if e.selected() != "strings.go" {
		t.Errorf("aparece seleccionado: %s", e.selected())
	}
	if !e.panelHas("util/") || !e.panelHas("strings.go") {
		t.Errorf("árbol:\n%s", strings.Join(e.panel(), "\n"))
	}
	e.mode(editor.ModeNormal)

	// «a» sobre un archivo crea al lado; con «/» final crea una carpeta
	e.typ("a").typ("otra/").send(keyEnter)
	if info, err := os.Stat(filepath.Join(root, "util", "otra")); err != nil || !info.IsDir() {
		t.Errorf("carpeta creada junto al archivo seleccionado: %v", err)
	}
	// «a» sobre una carpeta crea dentro
	e.typ("a").typ("dentro.txt").send(keyEnter)
	if _, err := os.Stat(filepath.Join(root, "util", "otra", "dentro.txt")); err != nil {
		t.Errorf("dentro de la carpeta seleccionada: %v", err)
	}
}

func TestCreateErrorsAreShownInline(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e").typ("a").typ("a.txt").send(keyEnter)
	if !strings.Contains(e.m.Message(), "ya existe") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	if !strings.Contains(e.screen(), "ya existe") {
		t.Error("el error se ve en la línea de comandos")
	}
	e.typ("a").typ("../fuera.txt").send(keyEnter)
	if !strings.Contains(e.m.Message(), "no válido") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "fuera.txt")); err == nil {
		t.Error("nunca se crea fuera de la carpeta")
	}
	e.typ("a").typ("nada").send(keyEsc) // Esc cancela
	if _, err := os.Stat(filepath.Join(root, "nada")); err == nil {
		t.Error("cancelar no crea nada")
	}
	e.mode(editor.ModeNormal)
	e.typ("a").send(keyEnter) // nombre vacío: no hace nada
	if strings.Contains(e.m.Message(), "creado") || e.m.Mode() != editor.ModeNormal {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestRenameUpdatesOpenBuffer(t *testing.T) {
	root := project(t, "a.txt=uno\n", "b.txt=dos\n")
	e := openProj(t, root).space("e").typ("j").send(keyEnter) // abre a.txt
	e.typ("ix").send(keyEsc)                                  // cambios sin guardar
	e.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if e.selected() != "a.txt" {
		t.Fatal(e.selected())
	}
	e.typ("r")
	if !strings.Contains(e.screen(), "Renombrar: a.txt") {
		t.Errorf("el prompt viene con el nombre actual:\n%s", e.screen())
	}
	e.send(keyBackspace, keyBackspace, keyBackspace, keyBackspace).typ("c.md").send(keyEnter)
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Error("el original ya no existe")
	}
	newName := "ac.md" // «a.txt» sin «.txt» más «c.md»
	if _, err := os.Stat(filepath.Join(root, newName)); err != nil {
		entries, _ := os.ReadDir(root)
		t.Fatalf("no se creó el archivo renombrado: %v", entries)
	}
	if got := e.m.Buffer().Path(); filepath.Base(got) != newName {
		t.Errorf("el buffer usa la nueva ruta: %q", got)
	}
	if !e.m.Buffer().Modified() || e.text() != "xuno\n" {
		t.Error("el buffer conserva sus cambios sin guardar")
	}
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}).ex("w")
	if got, _ := os.ReadFile(filepath.Join(root, newName)); string(got) != "xuno\n" {
		t.Errorf("guardar escribe en el nuevo nombre: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Error("guardar no recrea el nombre viejo")
	}
}

func TestRenameFolderUpdatesBuffersInside(t *testing.T) {
	root := project(t, "src/main.go=package main\n")
	e := openProj(t, root)
	e.ex("e " + filepath.Join(root, "src", "main.go"))
	e.space("e")
	e.typ("gg").typ("j") // src
	if e.selected() != "src" {
		t.Fatal(e.selected())
	}
	e.typ("r")
	e.send(keyBackspace, keyBackspace, keyBackspace).typ("lib").send(keyEnter)
	if got := e.m.Buffer().Path(); got != filepath.Join(root, "lib", "main.go") {
		t.Errorf("ruta del buffer = %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, "lib", "main.go")); err != nil {
		t.Error(err)
	}
}

func TestRenameConflictAndRoot(t *testing.T) {
	root := project(t, "a.txt", "b.txt")
	e := openProj(t, root).space("e")
	e.typ("r") // sobre la raíz
	if !strings.Contains(e.m.Message(), "raíz") || e.m.Mode() != editor.ModeNormal {
		t.Errorf("no se renombra la raíz: %q", e.m.Message())
	}
	e.typ("j").typ("r").send(keyBackspace, keyBackspace, keyBackspace, keyBackspace, keyBackspace).typ("b.txt").send(keyEnter)
	if !strings.Contains(e.m.Message(), "ya existe") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err != nil {
		t.Error("el original no cambia")
	}
}

func TestDeleteWithConfirmation(t *testing.T) {
	root := project(t, "a.txt", "dir/x.txt")
	e := openProj(t, root).space("e").typ("jj") // a.txt (dir va primero: dir, a.txt)
	if e.selected() != "a.txt" {
		t.Fatal(e.selected())
	}
	e.typ("d")
	if !e.m.DialogOpen() || !strings.Contains(e.screen(), "Borrar") || !strings.Contains(e.screen(), "a.txt") {
		t.Fatalf("pide confirmación:\n%s", e.screen())
	}
	e.send(keyEsc)
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err != nil || e.m.DialogOpen() {
		t.Error("cancelar no borra")
	}
	e.typ("d").typ("n") // otra forma de no confirmar: «n» no es «Borrar»
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err != nil {
		t.Error("n tampoco borra")
	}
	e.send(keyEsc)
	e.typ("d").typ("y")
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Error("confirmar borra")
	}
	if e.panelHas("a.txt") {
		t.Error("desaparece del árbol")
	}
	// una carpeta se borra con todo su contenido
	e.typ("gg").typ("j").typ("d")
	if !strings.Contains(e.screen(), "carpeta") {
		t.Errorf("el diálogo avisa de que borra la carpeta:\n%s", e.screen())
	}
	e.typ("y")
	if _, err := os.Stat(filepath.Join(root, "dir")); err == nil {
		t.Error("la carpeta se borra con su contenido")
	}
	e.typ("gg").typ("d")
	if e.m.DialogOpen() || !strings.Contains(e.m.Message(), "raíz") {
		t.Errorf("no se borra la raíz: %q", e.m.Message())
	}
}

func TestRefreshWithR(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e")
	os.WriteFile(filepath.Join(root, "desde_fuera.txt"), []byte("x"), 0o644)
	if e.panelHas("desde_fuera.txt") {
		t.Fatal("el árbol no se actualiza solo")
	}
	e.typ("R")
	if !e.panelHas("desde_fuera.txt") {
		t.Errorf("R refresca:\n%s", strings.Join(e.panel(), "\n"))
	}
}

func TestRefreshAfterSaveAndOnFocus(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e").typ("jl")
	e.send(tea.KeyPressMsg{Code: tea.KeyTab}).typ("ix").send(keyEsc).ex("w " + filepath.Join(root, "nuevo.txt"))
	if !e.panelHas("nuevo.txt") {
		t.Errorf("guardar un archivo nuevo lo muestra en el árbol:\n%s", strings.Join(e.panel(), "\n"))
	}
	os.WriteFile(filepath.Join(root, "otro.txt"), []byte("x"), 0o644)
	e.send(tea.FocusMsg{})
	if !e.panelHas("otro.txt") {
		t.Error("al recuperar el foco de la terminal se refresca")
	}
}

func TestExplorerHidesOnNarrowTerminal(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e")
	if !e.m.ExplorerVisible() {
		t.Fatal("visible a 100 columnas")
	}
	e.send(tea.WindowSizeMsg{Width: 79, Height: 24})
	if e.m.ExplorerVisible() || strings.Contains(e.screen(), "a.txt") {
		t.Error("con menos de 80 columnas se oculta solo")
	}
	e.typ("ix") // las teclas van al editor aunque el foco estuviera en el explorador
	e.wantText("x\n")
	e.send(keyEsc).send(tea.WindowSizeMsg{Width: 100, Height: 24})
	if !e.m.ExplorerVisible() {
		t.Error("al ensanchar vuelve a mostrarse")
	}
}

func TestEditorUsesRemainingWidth(t *testing.T) {
	root := project(t, "a.txt")
	e := openProj(t, root).space("e").send(tea.KeyPressMsg{Code: tea.KeyTab}).typ("i").typ(strings.Repeat("x", 90))
	// 100 - 25 de panel - 4 de gutter = 71 columnas de texto: el cursor obliga a desplazar
	if e.m.Left() == 0 {
		t.Error("el área del editor es más estrecha con el explorador abierto")
	}
	if v := e.m.View(); v.Cursor == nil || v.Cursor.X < 25+4 {
		t.Errorf("el cursor se dibuja dentro del editor: %+v", v.Cursor)
	}
	e.send(keyEsc).send(tea.KeyPressMsg{Code: tea.KeyTab})
	if v := e.m.View(); v.Cursor != nil {
		t.Error("con el foco en el explorador no se muestra el cursor del editor")
	}
}

func TestExplorerScrollsWithSelection(t *testing.T) {
	var files []string
	for i := 0; i < 60; i++ {
		files = append(files, fmt.Sprintf("f%02d.txt", i))
	}
	e := openProj(t, project(t, files...)).space("e").typ("G")
	if e.selected() != "f59.txt" || !e.panelHas("f59.txt") {
		t.Errorf("la selección queda visible al bajar:\n%s", strings.Join(e.panel(), "\n"))
	}
	if e.panelHas("f00.txt") {
		t.Error("las primeras filas salen de la vista")
	}
	e.typ("gg")
	if !e.panelHas("f00.txt") {
		t.Error("al volver arriba reaparecen")
	}
}
