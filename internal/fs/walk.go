package fs

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Walk recorre root y envía por out las rutas de los archivos (relativas a root, con «/») en lotes de hasta batch
// elementos, respetando la lista de ignorados. No sigue enlaces simbólicos a directorios y omite los ilegibles.
// Cierra out al terminar o cuando ctx se cancela, de modo que quien consume puede ir mostrando resultados mientras llegan.
func Walk(ctx context.Context, root string, ignore []string, batch int, out chan<- []string) {
	defer close(out)
	if batch <= 0 {
		batch = 500
	}
	if len(ignore) == 0 {
		ignore = DefaultIgnore
	}
	skip := map[string]bool{}
	for _, n := range ignore {
		skip[n] = true
	}
	var buf []string
	flush := func() bool {
		if len(buf) == 0 {
			return true
		}
		select {
		case out <- buf:
			buf = nil
			return true
		case <-ctx.Done():
			return false
		}
	}
	var walk func(dir, rel string) bool
	walk = func(dir, rel string) bool {
		if ctx.Err() != nil {
			return false
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return true // un directorio ilegible no detiene el recorrido
		}
		sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name()) })
		var dirs []os.DirEntry
		for _, e := range entries {
			if skip[e.Name()] {
				continue
			}
			if e.Type().IsDir() {
				dirs = append(dirs, e)
				continue
			}
			buf = append(buf, filepath.ToSlash(filepath.Join(rel, e.Name())))
			if len(buf) >= batch && !flush() {
				return false
			}
		}
		for _, d := range dirs {
			if !walk(filepath.Join(dir, d.Name()), filepath.Join(rel, d.Name())) {
				return false
			}
		}
		return true
	}
	if walk(root, "") {
		flush()
	}
}
