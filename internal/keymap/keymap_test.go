package keymap_test

import (
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/keymap"
)

func newKM(t *testing.T) *keymap.Keymap {
	t.Helper()
	k := keymap.New()
	for seq, cmd := range map[string]string{
		"g g":         "cursor.first_line",
		"G":           "cursor.last_line",
		"d":           "operator.delete",
		"space f f":   "picker.files",
		"space f b":   "picker.buffers",
		"space space": "palette.open",
		"ctrl+d":      "cursor.half_down",
		"esc":         "mode.normal",
	} {
		if err := k.Bind("normal", seq, cmd); err != nil {
			t.Fatal(err)
		}
	}
	k.MustBind("insert", "esc", "mode.normal")
	return k
}

func TestResolve(t *testing.T) {
	k := newKM(t)
	for _, tc := range []struct {
		name    string
		mode    string
		pending []string
		key     string
		want    keymap.Status
		cmd     string
	}{
		{"tecla simple", "normal", nil, "G", keymap.Complete, "cursor.last_line"},
		{"prefijo", "normal", nil, "g", keymap.Pending, ""},
		{"secuencia completa", "normal", []string{"g"}, "g", keymap.Complete, "cursor.first_line"},
		{"prefijo no válido", "normal", []string{"g"}, "x", keymap.Unmapped, ""},
		{"líder en dos niveles", "normal", []string{"space"}, "f", keymap.Pending, ""},
		{"líder completo f", "normal", []string{"space", "f"}, "f", keymap.Complete, "picker.files"},
		{"líder completo b", "normal", []string{"space", "f"}, "b", keymap.Complete, "picker.buffers"},
		{"líder doble", "normal", []string{"space"}, "space", keymap.Complete, "palette.open"},
		{"tecla con modificador", "normal", nil, "ctrl+d", keymap.Complete, "cursor.half_down"},
		{"sin asignar", "normal", nil, "z", keymap.Unmapped, ""},
		{"otro modo", "insert", nil, "esc", keymap.Complete, "mode.normal"},
		{"modo sin teclas", "insert", nil, "G", keymap.Unmapped, ""},
		{"modo inexistente", "visual", nil, "G", keymap.Unmapped, ""},
		{"pendiente inexistente", "normal", []string{"q", "q"}, "g", keymap.Unmapped, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := k.Resolve(tc.mode, tc.pending, tc.key)
			if got.Status != tc.want || got.Command != tc.cmd {
				t.Errorf("Resolve(%s,%v,%q) = %v %q; quería %v %q", tc.mode, tc.pending, tc.key, got.Status, got.Command, tc.want, tc.cmd)
			}
		})
	}
}

func TestBindConflicts(t *testing.T) {
	k := keymap.New()
	k.MustBind("n", "g g", "a.b")
	for name, seq := range map[string]string{
		"duplicada":           "g g",
		"prefijo de otra":     "g",
		"extiende un comando": "g g g",
		"vacía":               "   ",
	} {
		if err := k.Bind("n", seq, "x.y"); err == nil {
			t.Errorf("%s: %q debía rechazarse", name, seq)
		}
	}
	if err := k.Bind("otro", "g", "x.y"); err != nil {
		t.Errorf("el mismo prefijo en otro modo es válido: %v", err)
	}
	// la tabla original no se corrompió con los intentos fallidos
	if r := k.Resolve("n", []string{"g"}, "g"); r.Status != keymap.Complete || r.Command != "a.b" {
		t.Errorf("tabla alterada: %+v", r)
	}
}

func TestHasPrefix(t *testing.T) {
	k := newKM(t)
	if !k.HasPrefix("normal", nil) || !k.HasPrefix("normal", []string{"g"}) || !k.HasPrefix("normal", []string{"space", "f"}) {
		t.Error("prefijos válidos")
	}
	if k.HasPrefix("normal", []string{"z"}) || k.HasPrefix("nada", []string{"g"}) {
		t.Error("prefijos inválidos")
	}
}

func TestKeysSplitsOnWhitespace(t *testing.T) {
	got := keymap.Keys("  space   f  f ")
	if len(got) != 3 || got[0] != "space" {
		t.Errorf("Keys = %q", got)
	}
}

func TestMustBindPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("debía entrar en pánico")
		}
	}()
	k := keymap.New()
	k.MustBind("n", "a", "x.y")
	k.MustBind("n", "a", "x.z")
}

func TestWalkListsEverythingInStableOrder(t *testing.T) {
	k := newKM(t)
	var got []string
	k.Walk(func(mode, seq, cmd string) { got = append(got, mode+"|"+seq+"|"+cmd) })
	if len(got) != 9 { // 8 secuencias en normal y 1 en insert
		t.Fatalf("se esperaban 9 secuencias, hay %d: %v", len(got), got)
	}
	if got[0] != "insert|esc|mode.normal" || got[1] != "normal|G|cursor.last_line" {
		t.Errorf("orden inesperado: %v", got[:3])
	}
	var again []string
	k.Walk(func(mode, seq, cmd string) { again = append(again, mode+"|"+seq+"|"+cmd) })
	for i := range got {
		if got[i] != again[i] {
			t.Fatal("el orden debe ser estable")
		}
	}
}
