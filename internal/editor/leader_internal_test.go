package editor

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/keymap"
)

func testRegistry(t *testing.T, cmds ...command.Command[*Model]) *command.Registry[*Model] {
	t.Helper()
	r := command.NewRegistry[*Model]()
	for _, c := range cmds {
		r.MustRegister(c)
	}
	return r
}

func noop(*Model, command.Args) tea.Cmd { return nil }

// Un atajo nuevo aparece solo en el panel, dentro de su grupo, sin tocar nada más.
func TestLeaderEntriesAreBuiltFromKeymapAndRegistry(t *testing.T) {
	reg := testRegistry(t,
		command.Command[*Model]{Name: "x.yes", Desc: "Hacer x y", Short: "x-y corto", Keys: []string{"Space x y"}, Run: noop},
		command.Command[*Model]{Name: "x.zed", Desc: "Hacer x z", Keys: []string{"Space x z"}, Run: noop},
		command.Command[*Model]{Name: "w.save", Desc: "Guardar", Short: "guardar", Keys: []string{"Space w"}, Run: noop},
		command.Command[*Model]{Name: "p.pal", Desc: "Paleta", Keys: []string{"Space Space"}, Run: noop},
	)
	km := keymap.New()
	km.MustBind("normal", "space x y", "x.yes")
	km.MustBind("normal", "space x z", "x.zed")
	km.MustBind("normal", "space w", "w.save")
	km.MustBind("normal", "space space", "p.pal")
	groups := map[string]string{"space x": "equis"}

	top := leaderEntries(km, reg, "normal", []string{"space"}, groups)
	var got []string
	for _, e := range top {
		got = append(got, e.Key+"="+e.Label)
	}
	want := []string{"x=equis", "w=guardar", "Space=Paleta"} // grupos primero; Space (la paleta) al final
	if len(got) != len(want) {
		t.Fatalf("entradas = %v, quería %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entrada %d = %q, quería %q", i, got[i], want[i])
		}
	}
	if !top[0].Group || top[1].Group {
		t.Error("solo x es un grupo")
	}

	inX := leaderEntries(km, reg, "normal", []string{"space", "x"}, groups)
	if len(inX) != 2 || inX[0].Label != "x-y corto" || inX[1].Label != "Hacer x z" {
		t.Errorf("dentro del grupo (etiqueta corta y, si falta, la descripción): %+v", inX)
	}

	// un comando nuevo con un atajo nuevo aparece sin cambiar nada más
	reg.MustRegister(command.Command[*Model]{Name: "x.new", Desc: "Nuevo", Keys: []string{"Space x n"}, Run: noop})
	km.MustBind("normal", "space x n", "x.new")
	inX = leaderEntries(km, reg, "normal", []string{"space", "x"}, groups)
	if len(inX) != 3 || inX[0].Key != "n" {
		t.Errorf("el atajo nuevo aparece en el grupo: %+v", inX)
	}

	// un grupo sin nombre se muestra con su tecla; un prefijo que no existe no da entradas
	km.MustBind("normal", "space q a", "x.new")
	for _, e := range leaderEntries(km, reg, "normal", []string{"space"}, groups) {
		if e.Key == "q" && (e.Label != "q" || !e.Group) {
			t.Errorf("grupo sin nombre: %+v", e)
		}
	}
	if got := leaderEntries(km, reg, "normal", []string{"space", "nada"}, groups); len(got) != 0 {
		t.Errorf("prefijo inexistente: %+v", got)
	}
	if got := leaderEntries(km, reg, "explorer", []string{"space"}, groups); len(got) != 0 {
		t.Errorf("otro contexto: %+v", got)
	}
}

// Todo lo que cuelga de Space en el editor real es un comando registrado o un grupo, y todo comando con atajo «Space ...» está en el árbol.
func TestRealLeaderTreeMatchesRegistry(t *testing.T) {
	for _, mode := range []string{"normal", modeExplorer} {
		var walk func(prefix []string)
		walk = func(prefix []string) {
			for _, ch := range defaultKeymap.Children(mode, prefix) {
				if ch.Group {
					walk(append(append([]string(nil), prefix...), ch.Key))
					continue
				}
				if _, ok := registry.Get(ch.Command); !ok {
					t.Errorf("%s %v %s: el comando %s no está registrado", mode, prefix, ch.Key, ch.Command)
				}
			}
		}
		walk([]string{"space"})
	}
	bound := map[string]bool{}
	defaultKeymap.Walk(func(mode, seq, cmd string) {
		if mode == "normal" && len(seq) > 5 && seq[:5] == "space" {
			bound[cmd] = true
		}
	})
	for _, c := range registry.List() {
		for _, k := range c.Keys {
			if len(k) > 6 && k[:6] == "Space " && !bound[c.Name] {
				t.Errorf("%s declara el atajo %q pero no está en el árbol de teclas", c.Name, k)
			}
		}
	}
}

func TestKeyLabel(t *testing.T) {
	if keyLabel("space") != "Space" || keyLabel("f") != "f" || keyLabel("Q") != "Q" {
		t.Error("keyLabel")
	}
}
