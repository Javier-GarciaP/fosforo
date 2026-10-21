// Package fs es el árbol de archivos del explorador y las operaciones sobre el disco.
// No depende de la interfaz: se prueba con directorios temporales.
package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultIgnore son los nombres que el árbol nunca muestra.
var DefaultIgnore = []string{".git", "node_modules"}

// Node es una entrada del árbol: un directorio o un archivo.
type Node struct {
	Name     string
	Path     string // ruta absoluta
	IsDir    bool
	Expanded bool
	Loaded   bool // los hijos ya se leyeron del disco
	Parent   *Node
	Children []*Node
	Err      error // error al leer el directorio (p. ej. permisos); se muestra en línea sin abortar
}

// Row es un nodo visible junto con su profundidad (la raíz es 0).
type Row struct {
	Node  *Node
	Depth int
}

// Tree es el árbol de un proyecto. Los hijos de un directorio se leen solo al expandirlo.
type Tree struct {
	Root   *Node
	ignore map[string]bool
}

// New crea el árbol de root, con la raíz expandida. Sin lista de ignorados se usa DefaultIgnore.
func New(root string, ignore ...string) (*Tree, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &Error{Op: "abrir", Path: abs, Err: ErrNotDir}
	}
	if len(ignore) == 0 {
		ignore = DefaultIgnore
	}
	t := &Tree{ignore: map[string]bool{}}
	for _, n := range ignore {
		t.ignore[n] = true
	}
	t.Root = &Node{Name: filepath.Base(abs), Path: abs, IsDir: true}
	_ = t.Expand(t.Root) // un error de lectura queda en Root.Err
	return t, nil
}

// load lee los hijos de un directorio: directorios primero y luego archivos, cada grupo por orden alfabético
// (sin distinguir mayúsculas). Los enlaces simbólicos nunca se siguen: se muestran como archivos, así no hay ciclos.
func (t *Tree) load(n *Node) error {
	entries, err := os.ReadDir(n.Path)
	n.Loaded, n.Err = true, err
	if err != nil {
		n.Children = nil
		return err
	}
	kids := make([]*Node, 0, len(entries))
	for _, e := range entries {
		if t.ignore[e.Name()] {
			continue
		}
		kids = append(kids, &Node{
			Name: e.Name(), Path: filepath.Join(n.Path, e.Name()), IsDir: e.Type().IsDir(), Parent: n,
		})
	}
	sort.SliceStable(kids, func(i, j int) bool {
		a, b := kids[i], kids[j]
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		la, lb := strings.ToLower(a.Name), strings.ToLower(b.Name)
		if la != lb {
			return la < lb
		}
		return a.Name < b.Name
	})
	n.Children = kids
	return nil
}

// Expand expande un directorio, leyéndolo del disco la primera vez. Si no se puede leer devuelve el error,
// lo guarda en n.Err y el directorio queda cerrado.
func (t *Tree) Expand(n *Node) error {
	if !n.IsDir {
		return nil
	}
	if !n.Loaded {
		if err := t.load(n); err != nil {
			return err
		}
	}
	n.Expanded = true
	return nil
}

// Collapse cierra un directorio (sus hijos se conservan en memoria).
func (n *Node) Collapse() {
	if n.IsDir && n.Parent != nil { // la raíz siempre está abierta
		n.Expanded = false
	}
}

// Visible devuelve los nodos que se ven, en orden de pantalla: la raíz y los hijos de cada directorio expandido.
func (t *Tree) Visible() []Row {
	var rows []Row
	var walk func(n *Node, depth int)
	walk = func(n *Node, depth int) {
		rows = append(rows, Row{Node: n, Depth: depth})
		if n.IsDir && n.Expanded {
			for _, c := range n.Children {
				walk(c, depth+1)
			}
		}
	}
	walk(t.Root, 0)
	return rows
}

// Find busca un nodo por ruta entre los ya cargados.
func (t *Tree) Find(path string) *Node {
	var find func(n *Node) *Node
	find = func(n *Node) *Node {
		if n.Path == path {
			return n
		}
		if !n.IsDir || !n.Loaded || !strings.HasPrefix(path, n.Path+string(filepath.Separator)) {
			return nil
		}
		for _, c := range n.Children {
			if r := find(c); r != nil {
				return r
			}
		}
		return nil
	}
	return find(t.Root)
}

// Reveal expande las carpetas padre de path y devuelve su nodo, o nil si está fuera del árbol o no existe.
func (t *Tree) Reveal(path string) *Node {
	rel, err := filepath.Rel(t.Root.Path, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		if rel == "." {
			return t.Root
		}
		return nil
	}
	n := t.Root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if err := t.Expand(n); err != nil {
			return nil
		}
		var next *Node
		for _, c := range n.Children {
			if c.Name == part {
				next = c
				break
			}
		}
		if next == nil {
			return nil
		}
		n = next
	}
	return n
}

// Refresh vuelve a leer del disco todos los directorios cargados, conservando cuáles estaban expandidos.
func (t *Tree) Refresh() {
	var refresh func(n *Node)
	refresh = func(n *Node) {
		if !n.Loaded {
			return
		}
		old := map[string]*Node{}
		for _, c := range n.Children {
			old[c.Name] = c
		}
		if err := t.load(n); err != nil {
			n.Expanded = n.Parent == nil // la raíz sigue «abierta» aunque no se pueda leer
			return
		}
		for i, c := range n.Children {
			if prev, ok := old[c.Name]; ok && prev.IsDir == c.IsDir {
				c.Expanded, c.Loaded, c.Children, c.Err = prev.Expanded, prev.Loaded, prev.Children, prev.Err
				for _, g := range c.Children {
					g.Parent = c
				}
				n.Children[i] = c
				refresh(c)
			}
		}
	}
	refresh(t.Root)
}
