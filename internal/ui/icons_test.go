package ui

import "testing"

func TestResolveIcon(t *testing.T) {
	for _, tc := range []struct {
		name         string
		isDir, open  bool
		wantKey      string
		wantColorKey string
	}{
		{"trie.go", false, false, "go", "teal"},
		{"TRIE.GO", false, false, "go", "teal"}, // extensión sin distinguir mayúsculas
		{"go.mod", false, false, "go", "teal"},  // nombre exacto
		{"Makefile", false, false, "make", "peach"},
		{"makefile", false, false, "make", "peach"},
		{"LICENSE", false, false, "license", "yellow"},
		{".gitignore", false, false, "git", "peach"},
		{"Dockerfile", false, false, "docker", "blue"},
		{"package.json", false, false, "npm", "red"}, // el nombre exacto gana a la extensión .json
		{"datos.json", false, false, "json", "yellow"},
		{"README.md", false, false, "markdown", "teal"},
		{"notas.xyz", false, false, "default", "dim"},
		{"sinextension", false, false, "default", "dim"},
		{"archivo.tar.gz", false, false, "archive", "dim"}, // la última extensión
		{"buffer", true, false, "folder", "dim"},
		{"keymap", true, true, "folder_open", "blue"},
		{"main.go", true, false, "folder", "dim"}, // una carpeta con nombre de archivo sigue siendo carpeta
	} {
		got := ResolveIcon(tc.name, tc.isDir, tc.open)
		want := iconDefs[tc.wantKey]
		if got != want || got.Color != tc.wantColorKey {
			t.Errorf("%s (dir=%v open=%v) = %+v, quería %s %+v", tc.name, tc.isDir, tc.open, got, tc.wantKey, want)
		}
	}
}

func TestEveryIconHasGlyphAndColor(t *testing.T) {
	if len(iconDefs) < 38 {
		t.Fatalf("faltan iconos: %d", len(iconDefs))
	}
	th := Dark()
	for key, ic := range iconDefs {
		if ic.Glyph == "" || ic.Color == "" {
			t.Errorf("%s incompleto: %+v", key, ic)
		}
		if th.color(ic.Color) == th.FG && ic.Color != "fg" {
			t.Errorf("%s: el color %q no existe en el tema", key, ic.Color)
		}
		for _, r := range ic.Glyph {
			if r < 0xE000 { // todos están en áreas de uso privado de las Nerd Fonts
				t.Errorf("%s: el glifo %U no es de uso privado", key, r)
			}
		}
	}
	for name, key := range iconByName {
		if _, ok := iconDefs[key]; !ok {
			t.Errorf("el nombre %q apunta a un icono inexistente (%s)", name, key)
		}
	}
	for ext, key := range iconByExt {
		if _, ok := iconDefs[key]; !ok {
			t.Errorf("la extensión %q apunta a un icono inexistente (%s)", ext, key)
		}
	}
}
