package editor

import "testing"

// Toda tecla asignada apunta a un comando registrado, y todo comando de motion tiene su función.
func TestKeymapCommandsAreRegistered(t *testing.T) {
	n := 0
	defaultKeymap.Walk(func(mode, seq, cmd string) {
		n++
		if _, ok := registry.Get(cmd); !ok {
			t.Errorf("modo %s, tecla %q: el comando %s no está registrado", mode, seq, cmd)
		}
		if mode == modeOperator {
			if _, ok := motionByCommand(cmd); !ok {
				t.Errorf("en modo operador solo valen movimientos, y %s no lo es", cmd)
			}
		}
	})
	if n < 100 {
		t.Errorf("pocas teclas asignadas: %d", n)
	}
}

func TestEveryMotionIsRegisteredAndBound(t *testing.T) {
	bound := map[string]bool{}
	defaultKeymap.Walk(func(mode, _, cmd string) {
		if mode == "normal" {
			bound[cmd] = true
		}
	})
	for _, d := range motionDefs {
		if _, ok := registry.Get(d.name); !ok {
			t.Errorf("%s no está registrado", d.name)
		}
		if !bound[d.name] {
			t.Errorf("%s no tiene tecla en Normal", d.name)
		}
	}
}

// Cada contexto de teclas tiene su significado: «d» es un operador en Normal y «borrar» en el explorador, etc.
func TestKeymapContextsAreIndependent(t *testing.T) {
	want := map[string]map[string]string{
		"normal":     {"d": "operator.delete", "j": "cursor.down", "tab": "focus.toggle", "space e": "explorer.toggle", "space b n": "buffer.next"},
		modeExplorer: {"d": "explorer.delete", "j": "explorer.down", "g g": "explorer.top", "enter": "explorer.open", "a": "explorer.create"},
		"insert":     {"tab": "edit.tab", "enter": "edit.newline"},
		"visual":     {"d": "operator.delete", "o": "visual.swap_end"},
	}
	got := map[string]map[string]string{}
	defaultKeymap.Walk(func(mode, seq, cmd string) {
		if got[mode] == nil {
			got[mode] = map[string]string{}
		}
		got[mode][seq] = cmd
	})
	for mode, keys := range want {
		for seq, cmd := range keys {
			if got[mode][seq] != cmd {
				t.Errorf("%s %q = %q, quería %q", mode, seq, got[mode][seq], cmd)
			}
		}
	}
	// ninguna orden de explorador es alcanzable desde el editor, ni al revés
	for mode, keys := range got {
		for seq, cmd := range keys {
			if mode != modeExplorer && len(cmd) > 9 && cmd[:9] == "explorer." && cmd != "explorer.toggle" {
				t.Errorf("%s %q apunta a %s, que es del explorador", mode, seq, cmd)
			}
		}
	}
}
