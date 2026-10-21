package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func project(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	w := func(name, body string) {
		p := filepath.Join(d, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("a.go", "package a\nfunc Hello() {}\n// hello world\n")
	w("sub/b.txt", "nada\n日本 hello\n")
	w("node_modules/x.js", "hello\n")
	w("bin.dat", "hello\x00\x01")
	return d
}

func collect(t *testing.T, o Options) ([]Match, Result) {
	t.Helper()
	var all []Match
	res, err := Run(context.Background(), o, func(b []Match) { all = append(all, b...) })
	if err != nil {
		t.Fatal(err)
	}
	return all, res
}

func backends(t *testing.T, f func(t *testing.T, force bool)) {
	t.Run("builtin", func(t *testing.T) { f(t, true) })
	if RipgrepPath() == "" {
		t.Log("rg no instalado: se omite el backend rg")
		return
	}
	t.Run("rg", func(t *testing.T) { f(t, false) })
}

func TestSearchSmartCaseAndIgnores(t *testing.T) {
	backends(t, func(t *testing.T, force bool) {
		d := project(t)
		got, _ := collect(t, Options{Root: d, Query: "hello", ForceBuiltin: force})
		// insensible: Hello() y hello world en a.go, y sub/b.txt; nunca node_modules ni binarios
		paths := map[string]int{}
		for _, m := range got {
			paths[m.Path]++
		}
		if paths["a.go"] != 2 || paths["sub/b.txt"] != 1 || len(paths) != 2 {
			t.Fatalf("resultados: %v", paths)
		}
		got, _ = collect(t, Options{Root: d, Query: "Hello", ForceBuiltin: force})
		if len(got) != 1 || got[0].Line != 2 || got[0].Col != 5 {
			t.Fatalf("smart-case: %+v", got)
		}
	})
}

func TestSearchColumnGraphemes(t *testing.T) {
	backends(t, func(t *testing.T, force bool) {
		d := project(t)
		got, _ := collect(t, Options{Root: d, Query: "hello", ForceBuiltin: force})
		for _, m := range got {
			if m.Path == "sub/b.txt" && (m.Col != 3 || m.Line != 2) {
				t.Fatalf("columna en grafemas: %+v", m)
			}
		}
	})
}

func TestSearchLimit(t *testing.T) {
	backends(t, func(t *testing.T, force bool) {
		d := project(t)
		got, res := collect(t, Options{Root: d, Query: "hello", Limit: 2, ForceBuiltin: force})
		if len(got) != 2 || !res.Truncated {
			t.Fatalf("límite: %d %+v", len(got), res)
		}
	})
}

func TestSearchEmptyAndNone(t *testing.T) {
	d := project(t)
	if got, _ := collect(t, Options{Root: d, Query: ""}); len(got) != 0 {
		t.Fatal("consulta vacía")
	}
	backends(t, func(t *testing.T, force bool) {
		if got, _ := collect(t, Options{Root: d, Query: "zzzz", ForceBuiltin: force}); len(got) != 0 {
			t.Fatal("sin coincidencias")
		}
	})
}

func TestSearchCancel(t *testing.T) {
	d := project(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Run(ctx, Options{Root: d, Query: "hello", ForceBuiltin: true}, func([]Match) {}); err == nil {
		t.Fatal("debería devolver el error de cancelación")
	}
}
