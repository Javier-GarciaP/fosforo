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
