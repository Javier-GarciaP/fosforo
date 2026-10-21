package editor

import (
	"sort"
	"strings"

	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/keymap"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// leaderGroups da nombre a los grupos de la tecla líder (por su prefijo). Un grupo sin nombre aquí se muestra con su tecla.
// Los comandos no necesitan estar aquí: las entradas se generan recorriendo el árbol de teclas.
var leaderGroups = map[string]string{
	"space f": "archivos",
	"space b": "buffers",
	"space m": "markdown",
	"space s": "buscar",
}

// keyLabel es cómo se muestra una tecla en el panel.
func keyLabel(key string) string {
	if key == "space" {
		return "Space"
	}
	return key
}

// leaderEntries genera las entradas del panel para un prefijo a partir del árbol de teclas y del registro: cada una es un
// comando registrado o un grupo, de modo que el panel nunca se desincroniza de los atajos. Los grupos van primero.
func leaderEntries(km *keymap.Keymap, reg *command.Registry[*Model], mode string, prefix []string, groups map[string]string) []ui.LeaderEntry {
	var groupsOut, cmdsOut []ui.LeaderEntry
	for _, ch := range km.Children(mode, prefix) {
		e := ui.LeaderEntry{Key: keyLabel(ch.Key), Group: ch.Group}
		if ch.Group {
			full := strings.Join(append(append([]string(nil), prefix...), ch.Key), " ")
			e.Label = groups[full]
			if e.Label == "" {
				e.Label = ch.Key
			}
			groupsOut = append(groupsOut, e)
			continue
		}
		if c, ok := reg.Get(ch.Command); ok {
			e.Label = c.Label()
		} else {
			e.Label = ch.Command // no debería ocurrir: una prueba comprueba que todo está registrado
		}
		cmdsOut = append(cmdsOut, e)
	}
	byKey := func(es []ui.LeaderEntry) {
		sort.SliceStable(es, func(i, j int) bool {
			a, b := strings.ToLower(es[i].Key), strings.ToLower(es[j].Key)
			if (a == "space") != (b == "space") { // «Space» (la paleta) al final
				return b == "space"
			}
			if a != b {
				return a < b
			}
			return es[i].Key < es[j].Key
		})
	}
	byKey(groupsOut)
	byKey(cmdsOut)
	return append(groupsOut, cmdsOut...)
}

// leaderView devuelve el panel de ayuda de la secuencia pendiente y el texto de la línea de comandos.
func (m *Model) leaderView() (*ui.LeaderView, string) {
	mode := m.keymapContext()
	title := "<Space> líder"
	prefix := strings.Join(m.pending[1:], " ")
	if prefix != "" {
		full := strings.Join(m.pending, " ")
		name := leaderGroups[full]
		title = "<Space> " + prefix
		if name != "" {
			title += "  " + name
		}
	}
	entries := leaderEntries(defaultKeymap, registry, mode, m.pending, leaderGroups)
	if len(entries) == 0 {
		return nil, ""
	}
	hint := "<Space> "
	if prefix != "" {
		hint += prefix + " "
	}
	return &ui.LeaderView{Title: title, Entries: entries}, hint + "esperando tecla..."
}

// keymapContext es el árbol de teclas que está activo ahora (el que resuelve las pulsaciones).
func (m *Model) keymapContext() string {
	switch {
	case m.focusInExplorer():
		return modeExplorer
	}
	return m.mode.keymapName()
}
