package fuzzy_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/fuzzy"
)

func TestSubsequence(t *testing.T) {
	files := []string{"internal/keymap/trie.go", "docs/README.md"}
	got := fuzzy.Filter("ktr", files, 0)
	if len(got) != 1 || files[got[0].Index] != "internal/keymap/trie.go" {
		t.Fatalf("ktr debe encontrar solo trie.go: %+v", got)
	}
	// k, t y r resaltadas: «k» de keymap, «t» y «r» de trie
	if want := []int{9, 16, 17}; !reflect.DeepEqual(got[0].Pos, want) {
		t.Errorf("posiciones = %v, quería %v", got[0].Pos, want)
	}
}

func TestNoMatchWhenOrderOrLettersMissing(t *testing.T) {
	for _, tc := range []struct{ pat, text string }{
		{"abc", "acb"}, {"xyz", "abc"}, {"aa", "a"}, {"ab", ""}, {"trie", "tri"},
	} {
		if _, ok := fuzzy.Score(tc.pat, tc.text); ok {
			t.Errorf("%q no debe coincidir con %q", tc.pat, tc.text)
		}
	}
	if m, ok := fuzzy.Score("", "lo que sea"); !ok || m.Score != 0 || len(m.Pos) != 0 {
		t.Error("el patrón vacío coincide con todo")
	}
}

func TestSmartCase(t *testing.T) {
	if _, ok := fuzzy.Score("trie", "TRIE.GO"); !ok {
		t.Error("en minúsculas no distingue")
	}
	if _, ok := fuzzy.Score("Trie", "trie.go"); ok {
		t.Error("con mayúsculas en el patrón sí distingue")
	}
	if _, ok := fuzzy.Score("Trie", "Trie.go"); !ok {
		t.Error("coincidencia exacta con mayúsculas")
	}
}

func rank(pattern string, items ...string) []string {
	var out []string
	for _, r := range fuzzy.Filter(pattern, items, 0) {
		out = append(out, items[r.Index])
	}
	return out
}

func TestRanking(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pattern string
		items   []string
		first   string
	}{
		{"el nombre de archivo gana", "trie", []string{"changes/keymap-trie/design.md", "internal/keymap/trie.go"}, "internal/keymap/trie.go"},
		{"consecutivo mejor que disperso", "abc", []string{"a_b_c.txt", "abc.txt"}, "abc.txt"},
		{"inicio de palabra", "fb", []string{"xfxbx", "foo_bar"}, "foo_bar"},
		{"mayúscula intermedia", "gb", []string{"ggbb", "GoBack"}, "GoBack"},
		{"más corto a igualdad", "go", []string{"internal/ui/golden_helper.go", "go.mod"}, "go.mod"},
		{"tras una barra", "kt", []string{"akxt.go", "x/k/t.go"}, "x/k/t.go"},
		{"coincidencia exacta primero", "main", []string{"domain.go", "main.go", "remaining.txt"}, "main.go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := rank(tc.pattern, tc.items...)
			if len(got) == 0 || got[0] != tc.first {
				t.Errorf("orden = %v, el primero debía ser %q", got, tc.first)
			}
		})
	}
}

func TestOrderIsDeterministicAndStable(t *testing.T) {
	items := []string{"b.go", "a.go", "c.go", "ab.go"}
	first := rank("go", items...)
	for i := 0; i < 20; i++ {
		if got := rank("go", items...); !reflect.DeepEqual(got, first) {
			t.Fatalf("el orden cambia entre ejecuciones: %v vs %v", got, first)
		}
	}
	if first[0] != "b.go" { // misma puntuación y longitud: gana el orden original
		t.Errorf("a igualdad gana el orden original: %v", first)
	}
}

func TestPositionsAreValidAndMatchPattern(t *testing.T) {
	text := "internal/keymap/trie_test.go"
	for _, pat := range []string{"ktr", "itg", "internal", "tt", "go", "i/k/t"} {
		m, ok := fuzzy.Score(pat, text)
		if !ok {
			continue
		}
		rt := []rune(text)
		if len(m.Pos) != len([]rune(pat)) {
			t.Fatalf("%q: %d posiciones", pat, len(m.Pos))
		}
		prev := -1
		for k, p := range m.Pos {
			if p <= prev || p >= len(rt) || strings.ToLower(string(rt[p])) != strings.ToLower(string([]rune(pat)[k])) {
				t.Errorf("%q: posición %d inválida (%v)", pat, p, m.Pos)
			}
			prev = p
		}
	}
}

func TestUnicode(t *testing.T) {
	m, ok := fuzzy.Score("ñd", "señal_ñandú.txt")
	if !ok || len(m.Pos) != 2 {
		t.Fatalf("%+v %v", m, ok)
	}
	if rt := []rune("señal_ñandú.txt"); rt[m.Pos[0]] != 'ñ' || rt[m.Pos[1]] != 'd' {
		t.Errorf("posiciones por runa: %v", m.Pos)
	}
}

func TestFilterLimit(t *testing.T) {
	var items []string
	for i := 0; i < 500; i++ {
		items = append(items, "archivo_"+strings.Repeat("x", i%7)+".go")
	}
	if got := fuzzy.Filter("go", items, 200); len(got) != 200 {
		t.Errorf("límite: %d", len(got))
	}
	if got := fuzzy.Filter("go", items, 0); len(got) != 500 {
		t.Errorf("sin límite: %d", len(got))
	}
}

func BenchmarkFilter50k(b *testing.B) {
	items := make([]string, 50000)
	for i := range items {
		items[i] = "internal/pkg" + strings.Repeat("a", i%13) + "/sub/archivo_" + strings.Repeat("b", i%9) + ".go"
	}
	for i := 0; i < b.N; i++ {
		fuzzy.Filter("pkgsub", items, 200)
	}
}
