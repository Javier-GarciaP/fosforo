package fs_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Javier-GarciaP/fosforo/internal/fs"
)

// mk crea archivos (y carpetas con «/» final) dentro de root.
func mk(t *testing.T, root string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(root, n)
		if strings.HasSuffix(n, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func names(rows []fs.Row) []string {
	var out []string
	for _, r := range rows {
		out = append(out, strings.Repeat("  ", r.Depth)+r.Node.Name)
	}
	return out
}

func TestOrderDirectoriesFirstThenCaseInsensitive(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "zeta.txt", "Alfa.txt", "beta/", "Docs/", "alfa2.txt", ".hidden", "Makefile")
	tr, err := fs.New(root)
	if err != nil {
		t.Fatal(err)
	}
	got := names(tr.Visible())
	want := []string{filepath.Base(root), "  beta", "  Docs", "  .hidden", "  Alfa.txt", "  alfa2.txt", "  Makefile", "  zeta.txt"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("orden:\n got %q\nwant %q", got, want)
	}
}

func TestIgnoredNamesAreHidden(t *testing.T) {
	root := t.TempDir()
	mk(t, root, ".git/config", "node_modules/x/y.js", "src/main.go", "node_modules.txt")
	tr, _ := fs.New(root)
	got := strings.Join(names(tr.Visible()), "|")
	for _, n := range tr.Visible() {
		if n.Node.Name == ".git" || n.Node.Name == "node_modules" {
			t.Errorf("%s no debe mostrarse", n.Node.Name)
		}
	}
	if !strings.Contains(got, "node_modules.txt") {
		t.Error("solo se ocultan los nombres exactos")
	}
	// lista de ignorados propia
	tr2, _ := fs.New(root, "src")
	for _, n := range tr2.Visible() {
		if n.Node.Name == "src" {
			t.Error("src debía ignorarse")
		}
	}
}

func TestLazyLoading(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "a/b/c.txt", "d/e.txt")
	tr, _ := fs.New(root)
	a := tr.Root.Children[0]
	if a.Name != "a" || a.Loaded || a.Expanded {
		t.Fatalf("a no debe leerse hasta expandirse: %+v", a)
	}
	if rows := tr.Visible(); len(rows) != 3 { // raíz, a, d
		t.Errorf("filas = %d", len(rows))
	}
	if err := tr.Expand(a); err != nil {
		t.Fatal(err)
	}
	if !a.Loaded || !a.Expanded || len(a.Children) != 1 || a.Children[0].Loaded {
		t.Errorf("tras expandir a: %+v", a)
	}
	a.Collapse()
	if a.Expanded || !a.Loaded {
		t.Error("colapsar conserva los hijos cargados")
	}
	tr.Root.Collapse()
	if !tr.Root.Expanded {
		t.Error("la raíz no se colapsa")
	}
}

func TestUnreadableDirectoryDoesNotAbort(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root no hay directorios ilegibles")
	}
	root := t.TempDir()
	mk(t, root, "privado/secreto.txt", "ok/a.txt")
	if err := os.Chmod(filepath.Join(root, "privado"), 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(filepath.Join(root, "privado"), 0o755)
	tr, err := fs.New(root)
	if err != nil {
		t.Fatal(err)
	}
	var priv *fs.Node
	for _, c := range tr.Root.Children {
		if c.Name == "privado" {
			priv = c
		}
	}
	if err := tr.Expand(priv); err == nil || priv.Err == nil || priv.Expanded {
		t.Fatalf("debe dar error y quedar cerrado: err=%v node=%+v", err, priv)
	}
	if err := tr.Expand(tr.Root.Children[1]); err != nil { // «ok»
		t.Fatalf("el resto del árbol sigue funcionando: %v", err)
	}
}

func TestSymlinksAreNotFollowed(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "real/a.txt")
	if err := os.Symlink(root, filepath.Join(root, "real", "ciclo")); err != nil {
		t.Skip("enlaces simbólicos no disponibles")
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "enlace")); err != nil {
		t.Fatal(err)
	}
	tr, _ := fs.New(root)
	for _, c := range tr.Root.Children {
		if c.Name == "enlace" && c.IsDir {
			t.Error("un enlace a directorio no es un directorio del árbol")
		}
	}
	real := tr.Reveal(filepath.Join(root, "real"))
	for _, c := range real.Children {
		if c.Name == "ciclo" && (c.IsDir || len(c.Children) > 0) {
			t.Error("no debe seguirse el enlace que apunta a la raíz")
		}
	}
	if len(tr.Visible()) > 10 {
		t.Errorf("hay un ciclo: %d filas", len(tr.Visible()))
	}
}

func TestRevealExpandsParents(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "internal/keymap/trie.go", "internal/buffer/b.go", "docs/README.md")
	tr, _ := fs.New(root)
	n := tr.Reveal(filepath.Join(root, "internal", "keymap", "trie.go"))
	if n == nil || n.Name != "trie.go" {
		t.Fatalf("Reveal = %+v", n)
	}
	got := strings.Join(names(tr.Visible()), "|")
	want := strings.Join([]string{filepath.Base(root), "  docs", "  internal", "    buffer", "    keymap", "      trie.go"}, "|")
	if got != want {
		t.Errorf("\n got %s\nwant %s", got, want)
	}
	if tr.Reveal(filepath.Join(root, "no", "existe")) != nil || tr.Reveal("/otra/ruta") != nil || tr.Reveal(filepath.Dir(root)) != nil {
		t.Error("rutas inexistentes o fuera del árbol devuelven nil")
	}
	if tr.Reveal(root) != tr.Root {
		t.Error("la raíz se revela a sí misma")
	}
	if tr.Find(filepath.Join(root, "internal", "keymap", "trie.go")) != n {
		t.Error("Find")
	}
	if tr.Find(filepath.Join(root, "docs", "README.md")) != nil {
		t.Error("Find solo mira lo ya cargado")
	}
}

func TestRefreshKeepsExpansionAndPicksUpChanges(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "a/x.txt", "b/y.txt")
	tr, _ := fs.New(root)
	tr.Reveal(filepath.Join(root, "a", "x.txt"))
	mk(t, root, "nuevo.txt", "a/z.txt")
	os.Remove(filepath.Join(root, "b", "y.txt"))
	tr.Refresh()
	got := strings.Join(names(tr.Visible()), "|")
	if !strings.Contains(got, "nuevo.txt") || !strings.Contains(got, "z.txt") || !strings.Contains(got, "x.txt") {
		t.Errorf("tras refrescar: %s", got)
	}
	if !tr.Find(filepath.Join(root, "a")).Expanded {
		t.Error("a sigue expandido")
	}
	// un archivo borrado que estaba cargado desaparece
	os.RemoveAll(filepath.Join(root, "a"))
	tr.Refresh()
	if strings.Contains(strings.Join(names(tr.Visible()), "|"), "x.txt") {
		t.Error("lo borrado debe desaparecer")
	}
}

func TestNewOnFileFails(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "f.txt")
	if _, err := fs.New(filepath.Join(root, "f.txt")); !errors.Is(err, fs.ErrNotDir) {
		t.Errorf("err = %v", err)
	}
	if _, err := fs.New(filepath.Join(root, "nada")); err == nil {
		t.Error("una ruta inexistente falla")
	}
}

func TestCreate(t *testing.T) {
	root := t.TempDir()
	p, err := fs.Create(root, "util/strings.go")
	if err != nil || p != filepath.Join(root, "util", "strings.go") {
		t.Fatalf("%q %v", p, err)
	}
	if info, err := os.Stat(p); err != nil || info.IsDir() || info.Mode().Perm() != 0o644 {
		t.Errorf("archivo creado: %v %v", info, err)
	}
	d, err := fs.Create(root, "a/b/c/")
	if err != nil || !isDir(d) {
		t.Errorf("directorio: %q %v", d, err)
	}
	if _, err := fs.Create(root, "util/strings.go"); !errors.Is(err, fs.ErrExists) {
		t.Errorf("duplicado: %v", err)
	}
	if _, err := fs.Create(root, "a/b/"); !errors.Is(err, fs.ErrExists) {
		t.Errorf("carpeta existente: %v", err)
	}
	for _, bad := range []string{"", "  ", ".", "..", "../fuera.txt", "/abs.txt", "a/../../x", "a\x00b"} {
		if _, err := fs.Create(root, bad); !errors.Is(err, fs.ErrInvalidName) {
			t.Errorf("%q debía ser ErrInvalidName: %v", bad, err)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "fuera.txt")); err == nil {
		t.Error("nunca se escribe fuera de la carpeta")
	}
	// ruta con archivo en medio
	if _, err := fs.Create(root, "util/strings.go/x.txt"); err == nil {
		t.Error("crear dentro de un archivo falla")
	}
}

func isDir(p string) bool { i, err := os.Stat(p); return err == nil && i.IsDir() }

func TestRename(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "a.txt", "b.txt", "dir/x.txt")
	p, err := fs.Rename(filepath.Join(root, "a.txt"), "c.txt")
	if err != nil || p != filepath.Join(root, "c.txt") {
		t.Fatalf("%q %v", p, err)
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Error("el original ya no existe")
	}
	if _, err := fs.Rename(filepath.Join(root, "c.txt"), "b.txt"); !errors.Is(err, fs.ErrExists) {
		t.Errorf("destino existente: %v", err)
	}
	if _, err := fs.Rename(filepath.Join(root, "nada"), "z"); !errors.Is(err, fs.ErrNotFound) {
		t.Errorf("origen inexistente: %v", err)
	}
	if _, err := fs.Rename(filepath.Join(root, "c.txt"), "../fuera"); !errors.Is(err, fs.ErrInvalidName) {
		t.Errorf("fuera de la carpeta: %v", err)
	}
	if same, err := fs.Rename(filepath.Join(root, "b.txt"), "b.txt"); err != nil || same != filepath.Join(root, "b.txt") {
		t.Errorf("mismo nombre no hace nada: %q %v", same, err)
	}
	nd, err := fs.Rename(filepath.Join(root, "dir"), "otro/dir2") // renombrar y mover
	if err != nil || !isDir(nd) {
		t.Errorf("carpeta: %q %v", nd, err)
	}
	if _, err := os.Stat(filepath.Join(nd, "x.txt")); err != nil {
		t.Error("el contenido viaja con la carpeta")
	}
}

func TestRemove(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "a.txt", "dir/sub/x.txt")
	if err := fs.Remove(filepath.Join(root, "a.txt")); err != nil {
		t.Fatal(err)
	}
	if err := fs.Remove(filepath.Join(root, "dir")); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Errorf("quedan %v", entries)
	}
	err := fs.Remove(filepath.Join(root, "a.txt"))
	var fe *fs.Error
	if !errors.Is(err, fs.ErrNotFound) || !errors.As(err, &fe) || fe.Op != "borrar" {
		t.Errorf("borrar lo inexistente: %v", err)
	}
	if !strings.Contains(err.Error(), "no se pudo borrar") {
		t.Errorf("mensaje: %q", err.Error())
	}
}

func TestRemoveDoesNotFollowSymlink(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "real/a.txt")
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "enlace")); err != nil {
		t.Skip("enlaces simbólicos no disponibles")
	}
	if err := fs.Remove(filepath.Join(root, "enlace")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "real", "a.txt")); err != nil {
		t.Error("borrar el enlace no debe borrar su destino")
	}
}

func collect(t *testing.T, root string, ignore []string, batch int) ([]string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ch := make(chan []string, 4)
	go fs.Walk(ctx, root, ignore, batch, ch)
	var all []string
	batches := 0
	for b := range ch {
		batches++
		all = append(all, b...)
	}
	return all, batches
}

func TestWalkListsFilesRelativeAndIgnores(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "main.go", "internal/keymap/trie.go", "internal/buffer/b.go", ".git/config", "node_modules/x/y.js", "docs/README.md", ".hidden", "vacia/")
	all, _ := collect(t, root, nil, 100)
	want := []string{".hidden", "main.go", "docs/README.md", "internal/buffer/b.go", "internal/keymap/trie.go"}
	got := append([]string(nil), all...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("archivos = %v, quería %v", got, want)
	}
	for _, f := range all {
		if strings.Contains(f, ".git/") || strings.Contains(f, "node_modules") {
			t.Errorf("no debe listar %s", f)
		}
	}
	// lista de ignorados propia
	all, _ = collect(t, root, []string{"internal"}, 100)
	for _, f := range all {
		if strings.HasPrefix(f, "internal/") {
			t.Errorf("internal debía ignorarse: %s", f)
		}
	}
}

func TestWalkBatchesAndProgress(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 25; i++ {
		mk(t, root, fmt.Sprintf("d%d/f%02d.txt", i%3, i))
	}
	all, batches := collect(t, root, nil, 10)
	if len(all) != 25 || batches != 3 {
		t.Errorf("%d archivos en %d lotes, quería 25 en 3", len(all), batches)
	}
}

func TestWalkCancelStopsEarly(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 200; i++ {
		mk(t, root, fmt.Sprintf("f%03d.txt", i))
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan []string) // sin búfer: el recorrido se queda esperando a quien consume
	go fs.Walk(ctx, root, nil, 10, ch)
	<-ch
	cancel()
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("cancelar debe cerrar el canal y detener el recorrido")
	}
}

func TestWalkSkipsSymlinkedDirsAndUnreadable(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "real/a.txt", "b.txt")
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "enlace")); err != nil {
		t.Skip("enlaces simbólicos no disponibles")
	}
	os.Symlink(root, filepath.Join(root, "real", "ciclo"))
	all, _ := collect(t, root, nil, 100)
	count := map[string]int{}
	for _, f := range all {
		count[filepath.Base(f)]++
	}
	if count["a.txt"] != 1 {
		t.Errorf("a.txt aparece %d veces: %v", count["a.txt"], all)
	}
	if len(all) > 5 {
		t.Errorf("hay un ciclo: %v", all)
	}
}

func TestWalkMissingRootClosesImmediately(t *testing.T) {
	all, batches := collect(t, filepath.Join(t.TempDir(), "no-existe"), nil, 10)
	if len(all) != 0 || batches != 0 {
		t.Errorf("%v %d", all, batches)
	}
}
