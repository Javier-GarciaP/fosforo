package editor

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/fs"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// Focus es el panel que recibe las teclas.
type Focus int

const (
	FocusEditor Focus = iota
	FocusExplorer
)

// explorerState es el árbol de archivos y la fila seleccionada.
type explorerState struct {
	tree   *fs.Tree
	sel    int // índice en tree.Visible()
	scroll int
}

// WithRoot fija la carpeta raíz del explorador (por defecto, la carpeta de trabajo).
func WithRoot(dir string) Option { return func(m *Model) { m.root = absPath(dir) } }

// WithIcons elige si el explorador muestra iconos de Nerd Font (por defecto) o ninguno.
func WithIcons(mode ui.IconMode) Option { return func(m *Model) { m.icons = mode } }

// ExplorerVisible indica si el panel está a la vista (oculto si la terminal es demasiado estrecha).
func (m Model) ExplorerVisible() bool { return m.explorerShown() }

// Focus devuelve el panel con el foco.
func (m Model) Focus() Focus {
	if m.focusInExplorer() {
		return FocusExplorer
	}
	return FocusEditor
}

// explorerShown: el explorador está activado y cabe en pantalla (en terminales de menos de 80 columnas se oculta solo).
func (m *Model) explorerShown() bool {
	return m.showExplorer && m.explorer != nil && (m.width == 0 || ui.PanelCols(m.width, true) > 0)
}

func (m *Model) focusInExplorer() bool { return m.focus == FocusExplorer && m.explorerShown() }

// editorCols es el ancho de la zona del editor (sin el panel del explorador).
func (m *Model) editorCols() int {
	return m.width - ui.PanelCols(m.width, m.showExplorer && m.explorer != nil)
}

func (m *Model) ensureTree() bool {
	if m.explorer != nil {
		return true
	}
	root := m.root
	if root == "" {
		root = absPath(".")
	}
	t, err := fs.New(root)
	if err != nil {
		m.msg, m.msgErr = "E484: no se puede abrir el explorador: "+err.Error(), true
		return false
	}
	m.root = t.Root.Path
	m.explorer = &explorerState{tree: t}
	return true
}

// toggleExplorer muestra u oculta el panel. Al mostrarlo se revela el archivo actual y recibe el foco.
func (m *Model) toggleExplorer() {
	if m.showExplorer {
		m.showExplorer = false
		m.focus = FocusEditor
		return
	}
	if !m.ensureTree() {
		return
	}
	m.showExplorer, m.focus = true, FocusExplorer
	m.revealCurrent()
}

// revealCurrent expande las carpetas del archivo actual y lo selecciona.
func (m *Model) revealCurrent() {
	if p := m.buf.Path(); p != "" && m.explorer != nil {
		if n := m.explorer.tree.Reveal(absPath(p)); n != nil {
			m.selectNode(n.Path)
		}
	}
}

// selectNode selecciona en el panel el nodo con esa ruta, si está visible.
func (m *Model) selectNode(path string) {
	for i, r := range m.explorer.tree.Visible() {
		if r.Node.Path == path {
			m.explorer.sel = i
			return
		}
	}
}

func (m *Model) toggleFocus() {
	if !m.explorerShown() {
		m.msg = "El explorador está oculto (Space e lo muestra)"
		return
	}
	if m.focus == FocusExplorer {
		m.focus = FocusEditor
	} else {
		m.focus = FocusExplorer
	}
}

// selected devuelve la fila seleccionada y su nodo.
func (m *Model) selected() (fs.Row, bool) {
	if m.explorer == nil {
		return fs.Row{}, false
	}
	rows := m.explorer.tree.Visible()
	m.explorer.sel = max(0, min(m.explorer.sel, len(rows)-1))
	if len(rows) == 0 {
		return fs.Row{}, false
	}
	return rows[m.explorer.sel], true
}

func (m *Model) explorerMove(delta int) {
	rows := m.explorer.tree.Visible()
	m.explorer.sel = max(0, min(m.explorer.sel+delta, len(rows)-1))
}

// explorerOpen abre el archivo seleccionado (el foco pasa al editor) o expande el directorio.
func (m *Model) explorerOpen() {
	row, ok := m.selected()
	if !ok {
		return
	}
	n := row.Node
	if !n.IsDir {
		if m.openPath(n.Path, m.displayName(n.Path)) {
			m.focus = FocusEditor
		}
		return
	}
	if n.Expanded {
		if len(n.Children) > 0 { // ya abierto: baja al primer hijo
			m.explorerMove(1)
		}
		return
	}
	if err := m.explorer.tree.Expand(n); err != nil {
		m.msg, m.msgErr = fmt.Sprintf("No se puede leer %s: %v", n.Name, simplifyErr(err)), true
	}
}

func simplifyErr(err error) error {
	var fe *fs.Error
	if errors.As(err, &fe) {
		return fe.Err
	}
	if i := strings.LastIndex(err.Error(), ": "); i >= 0 {
		return errors.New(err.Error()[i+2:])
	}
	return err
}

// explorerCollapse cierra el directorio seleccionado o sube a su carpeta padre.
func (m *Model) explorerCollapse() {
	row, ok := m.selected()
	if !ok {
		return
	}
	n := row.Node
	if n.IsDir && n.Expanded && n.Parent != nil {
		n.Collapse()
		return
	}
	if n.Parent != nil {
		m.selectNode(n.Parent.Path)
	}
}

// refreshExplorer vuelve a leer el árbol conservando lo expandido y la selección.
func (m *Model) refreshExplorer() {
	if m.explorer == nil {
		return
	}
	var keep string
	if row, ok := m.selected(); ok {
		keep = row.Node.Path
	}
	m.explorer.tree.Refresh()
	if keep != "" {
		m.selectNode(keep)
	}
}

// targetDir es la carpeta donde se crea algo: la seleccionada si es una carpeta, o la que contiene al archivo seleccionado.
func (m *Model) targetDir() string {
	row, ok := m.selected()
	if !ok {
		return m.explorer.tree.Root.Path
	}
	if row.Node.IsDir {
		return row.Node.Path
	}
	return filepath.Dir(row.Node.Path)
}

// explorerCreate abre el prompt para crear un archivo (o una carpeta si termina en «/») dentro de la carpeta seleccionada.
func (m *Model) explorerCreate() {
	m.openPromptKind(promptCreate, "Nuevo: ", "", m.targetDir())
}

func (m *Model) explorerRename() {
	row, ok := m.selected()
	if !ok || row.Node.Parent == nil {
		m.msg, m.msgErr = "No se puede renombrar la raíz del proyecto", true
		return
	}
	m.openPromptKind(promptRename, "Renombrar: ", row.Node.Name, row.Node.Path)
}

func (m *Model) explorerDelete() {
	row, ok := m.selected()
	if !ok || row.Node.Parent == nil {
		m.msg, m.msgErr = "No se puede borrar la raíz del proyecto", true
		return
	}
	n := row.Node
	lines := []string{fmt.Sprintf("¿Borrar %q?", n.Name), "Esta acción no se puede deshacer."}
	if n.IsDir {
		lines = []string{fmt.Sprintf("¿Borrar la carpeta %q?", n.Name), "Se borra todo su contenido y no se puede deshacer."}
	}
	m.dialog = &dialog{
		title: "Borrar", lines: lines, buttons: []dialogButton{btnDelete, btnCancel},
		choose: func(m *Model, i int) {
			if i != 0 {
				return
			}
			if err := fs.Remove(n.Path); err != nil {
				m.msg, m.msgErr = err.Error(), true
				return
			}
			m.refreshExplorer()
			m.msg, m.msgErr = fmt.Sprintf("%q borrado", n.Name), false
		},
	}
}

// finishCreate crea lo escrito en el prompt y lo deja seleccionado en el árbol.
func (m *Model) finishCreate(dir, name string) {
	path, err := fs.Create(dir, name)
	if err != nil {
		m.msg, m.msgErr = err.Error(), true
		return
	}
	m.explorer.tree.Refresh()
	if n := m.explorer.tree.Reveal(path); n != nil {
		m.selectNode(n.Path)
	}
	m.msg, m.msgErr = fmt.Sprintf("%q creado", m.displayName(path)), false
}

// finishRename renombra y actualiza la ruta de los buffers abiertos afectados sin perder sus cambios.
func (m *Model) finishRename(oldPath, newName string) {
	newPath, err := fs.Rename(oldPath, newName)
	if err != nil {
		m.msg, m.msgErr = err.Error(), true
		return
	}
	m.stash()
	for _, bs := range m.bufs {
		p := absPath(bs.buf.Path())
		var np string
		switch {
		case p == oldPath:
			np = newPath
		case strings.HasPrefix(p, oldPath+string(filepath.Separator)):
			np = newPath + p[len(oldPath):]
		default:
			continue
		}
		bs.buf.SetPath(np)
		bs.label = m.displayName(np)
	}
	m.load(m.active)
	m.explorer.tree.Refresh()
	if n := m.explorer.tree.Reveal(newPath); n != nil {
		m.selectNode(n.Path)
	}
	m.msg, m.msgErr = fmt.Sprintf("%q renombrado a %q", filepath.Base(oldPath), m.displayName(newPath)), false
}

// explorerView prepara el panel para dibujarlo.
func (m *Model) explorerView() *ui.ExplorerView {
	if !m.showExplorer || m.explorer == nil || ui.PanelCols(m.width, true) == 0 {
		return nil
	}
	rows := m.explorer.tree.Visible()
	m.selected() // ajusta la selección a los límites
	out := make([]ui.ExplorerRow, len(rows))
	for i, r := range rows {
		out[i] = ui.ExplorerRow{
			Name: r.Node.Name, Depth: r.Depth, IsDir: r.Node.IsDir, Expanded: r.Node.Expanded,
			Root: r.Node.Parent == nil, Err: r.Node.Err != nil,
		}
	}
	return &ui.ExplorerView{
		Rows: out, Selected: m.explorer.sel, Scroll: m.explorer.scroll,
		Focused: m.focusInExplorer(), Icons: m.icons,
	}
}

// scrollExplorer mantiene visible la fila seleccionada.
func (m *Model) scrollExplorer() {
	if m.explorer == nil || m.height == 0 {
		return
	}
	m.selected()
	m.explorer.scroll = ui.ExplorerScroll(m.explorer.scroll, m.explorer.sel, len(m.explorer.tree.Visible()), ui.TextRows(m.height))
}

func registerExplorerCommands(r *command.Registry[*Model]) {
	reg := func(name, desc string, keys []string, run func(m *Model, a command.Args)) {
		r.MustRegister(command.Command[*Model]{Name: name, Desc: desc, Keys: keys, Run: func(m *Model, a command.Args) tea.Cmd { run(m, a); return nil }})
	}
	reg("explorer.toggle", "Mostrar u ocultar el explorador", []string{"Space e"}, func(m *Model, _ command.Args) { m.toggleExplorer() })
	reg("focus.toggle", "Cambiar el foco entre explorador y editor", []string{"Tab", "Ctrl+W"}, func(m *Model, _ command.Args) { m.toggleFocus() })
	reg("explorer.down", "Explorador: bajar", []string{"j", "↓"}, func(m *Model, a command.Args) { m.explorerMove(a.N()) })
	reg("explorer.up", "Explorador: subir", []string{"k", "↑"}, func(m *Model, a command.Args) { m.explorerMove(-a.N()) })
	reg("explorer.top", "Explorador: primera fila", []string{"gg"}, func(m *Model, _ command.Args) { m.explorer.sel = 0 })
	reg("explorer.bottom", "Explorador: última fila", []string{"G"}, func(m *Model, _ command.Args) {
		m.explorer.sel = len(m.explorer.tree.Visible()) - 1
	})
	reg("explorer.open", "Explorador: abrir archivo o expandir carpeta", []string{"Enter", "l"}, func(m *Model, _ command.Args) { m.explorerOpen() })
	reg("explorer.collapse", "Explorador: colapsar carpeta o ir a la carpeta padre", []string{"h"}, func(m *Model, _ command.Args) { m.explorerCollapse() })
	reg("explorer.create", "Explorador: crear archivo (o carpeta con «/» final)", []string{"a"}, func(m *Model, _ command.Args) { m.explorerCreate() })
	reg("explorer.rename", "Explorador: renombrar", []string{"r"}, func(m *Model, _ command.Args) { m.explorerRename() })
	reg("explorer.delete", "Explorador: borrar", []string{"d"}, func(m *Model, _ command.Args) { m.explorerDelete() })
	reg("explorer.refresh", "Explorador: refrescar", []string{"R"}, func(m *Model, _ command.Args) {
		m.refreshExplorer()
		m.msg = "Explorador actualizado"
	})
}

// ExplorerSelection devuelve la ruta de la fila seleccionada del explorador ("" si no hay explorador).
func (m Model) ExplorerSelection() string {
	if m.explorer == nil {
		return ""
	}
	rows := m.explorer.tree.Visible()
	if len(rows) == 0 {
		return ""
	}
	return rows[max(0, min(m.explorer.sel, len(rows)-1))].Node.Path
}

// BufferCount devuelve cuántos buffers hay abiertos y ActiveBuffer la posición (desde 0) del activo.
func (m Model) BufferCount() int  { return len(m.bufs) }
func (m Model) ActiveBuffer() int { return m.active }
