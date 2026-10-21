package editor_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

func grepProject(t *testing.T) *ed {
	root := project(t,
		"a.go=package a\n\nfunc Alpha() {}\nfunc beta() {}\n// alpha de nuevo\n",
		"docs/README.md=# Docs\n\nalpha en markdown\n",
		"node_modules/x.js=alpha\n",
	)
	return openProj(t, root, editor.WithGrepDelay(time.Millisecond))
}

func TestProjectSearchListsAndJumps(t *testing.T) {
	e := grepProject(t).space("sg")
	if !e.m.PickerOpen() || e.m.PickerTitle() != "Buscar en el proyecto" {
		t.Fatalf("Space s g abre la búsqueda: %q", e.m.PickerTitle())
	}
	if !strings.Contains(e.screen(), "Escribe para buscar") {
		t.Errorf("sin consulta se explica qué hacer:\n%s", e.screen())
	}
	e.typ("alpha").settle()
	res := e.results()
	if len(res) != 3 { // a.go ×2 (Alpha, alpha) + README; node_modules se ignora
		t.Fatalf("resultados: %v", res)
	}
	if !strings.HasPrefix(res[0], "a.go:3") {
		t.Errorf("orden por archivo y línea: %v", res)
	}
	e.send(keyDown, keyDown).send(keyEnter).settle() // README.md
	if e.m.PickerOpen() {
		t.Fatal("Enter cierra la ventana")
	}
	if !strings.HasSuffix(e.m.Buffer().Path(), "README.md") {
		t.Fatalf("abre el archivo: %q", e.m.Buffer().Path())
	}
	if got := e.m.Cursor(); got.Line != 2 || got.Col != 0 {
		t.Errorf("cursor en la línea del resultado: %+v", got)
	}
	if e.m.Mode() != editor.ModeNormal || strings.Contains(e.screen(), "PREVIEW") {
		t.Errorf("un .md se abre en el fuente, no en la vista previa:\n%s", e.screen())
	}
}

func TestProjectSearchNewQueryReplacesResults(t *testing.T) {
	e := grepProject(t).space("sg").typ("alpha").settle()
	if len(e.results()) != 3 {
		t.Fatalf("%v", e.results())
	}
	e.typ("zzz").settle()
	if len(e.results()) != 0 || !strings.Contains(e.screen(), "Sin resultados") {
		t.Errorf("sin coincidencias: %v", e.results())
	}
	e.send(keyBackspace, keyBackspace, keyBackspace).settle()
	if len(e.results()) != 3 {
		t.Errorf("al borrar vuelven los resultados: %v", e.results())
	}
	e.send(keyEsc)
	if e.m.PickerOpen() {
		t.Error("Esc cierra")
	}
}

func TestProjectSearchWordUnderCursor(t *testing.T) {
	root := project(t, "a.go=func Alpha() {}\nvar x = Alpha\n", "b.go=Alpha()\n")
	e := openProj(t, root, editor.WithGrepDelay(time.Millisecond))
	e.ex("e " + root + "/a.go")
	e.typ("w").space("sw").settle()
	if e.m.PickerTitle() != "Buscar palabra" {
		t.Fatalf("título: %q", e.m.PickerTitle())
	}
	if res := e.results(); len(res) != 3 {
		t.Fatalf("resultados para «Alpha»: %v", res)
	}
	if !strings.Contains(e.screen(), "Alpha") {
		t.Error("la consulta se muestra")
	}
}
