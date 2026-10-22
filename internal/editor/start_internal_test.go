package editor

import (
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// Cada atajo de la pantalla de inicio sigue asignado, en modo Normal, al comando que dice la etiqueta.
// Si alguien cambia una tecla o un comando, este test avisa antes de que la pantalla enseñe algo falso.
func TestStartEntriesAreBound(t *testing.T) {
	want := map[string]string{
		"Buscar archivo":     "file.find",
		"Nuevo archivo":      "file.new",
		"Archivos recientes": "file.recent",
		"Buscar texto":       "search.project",
		"Explorador":         "explorer.toggle",
		"Configuración":      "config.open",
		"Lista de atajos":    "help.keys",
		"Salir":              "app.quit",
	}
	bound := map[string]string{}
	defaultKeymap.Walk(func(mode, seq, cmd string) {
		if mode == "normal" {
			bound[seq] = cmd
		}
	})
	if len(ui.StartEntries) != len(want) {
		t.Fatalf("la lista tiene %d acciones y el test espera %d", len(ui.StartEntries), len(want))
	}
	for _, e := range ui.StartEntries {
		cmd, ok := want[e.Label]
		if !ok {
			t.Errorf("acción %q sin comando esperado en el test", e.Label)
			continue
		}
		seq := "space " + strings.TrimPrefix(e.Keys, "Space ")
		if got := bound[seq]; got != cmd {
			t.Errorf("%q: la tecla %q va a %q, no a %q", e.Label, e.Keys, got, cmd)
		}
	}
}
