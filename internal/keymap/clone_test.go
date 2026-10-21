package keymap_test

import (
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/keymap"
)

func TestCloneAndUnbind(t *testing.T) {
	k := keymap.New()
	k.MustBind("normal", "space f f", "file.find")
	k.MustBind("normal", "space f r", "file.recent")
	c := k.Clone()
	if !c.Unbind("normal", "space f f") || c.Unbind("normal", "space f f") {
		t.Fatal("unbind")
	}
	if k.Resolve("normal", []string{"space", "f"}, "f").Command != "file.find" {
		t.Fatal("el original no debe cambiar")
	}
	if c.Resolve("normal", []string{"space", "f"}, "f").Status != keymap.Unmapped {
		t.Fatal("la copia debe haber perdido la tecla")
	}
	c.Unbind("normal", "space f r")
	if c.HasPrefix("normal", []string{"space"}) {
		t.Fatal("los grupos vacíos se podan")
	}
	if c.Unbind("normal", "space") || c.Unbind("nope", "x") {
		t.Fatal("no existen")
	}
}
