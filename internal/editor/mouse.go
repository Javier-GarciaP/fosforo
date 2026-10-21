package editor

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// DoubleClickInterval es el tiempo máximo entre dos pulsaciones en la misma celda para que cuenten como doble clic.
const DoubleClickInterval = 400 * time.Millisecond

// wheelStep son las líneas que se desplaza la rueda por cada muesca.
const wheelStep = 3

// lastClick recuerda la última pulsación para detectar el doble clic.
type lastClick struct {
	x, y int
	at   time.Time
	ok   bool
}

// WithClock cambia el reloj con que se detecta el doble clic (para las pruebas).
func WithClock(now func() time.Time) Option { return func(m *Model) { m.now = now } }

// WithMouse activa o desactiva la captura del ratón al empezar (por defecto, activada).
func WithMouse(on bool) Option { return func(m *Model) { m.mouseOn = on } }

// MouseEnabled indica si el editor captura el ratón.
func (m Model) MouseEnabled() bool { return m.mouseOn }

// layout devuelve la geometría actual: la misma que usa la vista para dibujar.
func (m *Model) layout() ui.Layout {
	var md *ui.MarkdownView
	if s := m.mdState(); s != nil && s.view != mdSource {
		md = &ui.MarkdownView{Split: s.view == mdSplit}
	}
	return ui.ComputeLayout(m.width, m.height, m.showExplorer && m.explorer != nil, md, m.buf.LineCount())
}

// clickCount registra una pulsación y devuelve true si es el segundo clic de un doble clic.
func (m *Model) isDoubleClick(x, y int) bool {
	now := m.now()
	double := m.click.ok && m.click.x == x && m.click.y == y && now.Sub(m.click.at) <= DoubleClickInterval
	if double {
		m.click.ok = false // un tercer clic empieza otra cuenta
		return true
	}
	m.click = lastClick{x: x, y: y, at: now, ok: true}
	return false
}

// mouse atiende un evento del ratón. Cada acción se traduce a un comando del registro: no hay lógica paralela.
func (m *Model) mouse(msg tea.MouseMsg) tea.Cmd {
	if !m.mouseOn || m.width == 0 {
		return nil
	}
	mm := msg.Mouse()
	switch msg.(type) {
	case tea.MouseWheelMsg:
		return m.mouseWheel(mm)
	case tea.MouseClickMsg:
		if mm.Button != tea.MouseLeft {
			return nil
		}
		return m.mouseClick(mm)
	case tea.MouseMotionMsg:
		if mm.Button == tea.MouseLeft && m.dragging {
			return m.mouseDrag(mm)
		}
	case tea.MouseReleaseMsg:
		m.dragging = false
	}
	return nil
}

func (m *Model) mouseClick(mm tea.Mouse) tea.Cmd {
	m.dragging = false
	switch {
	case m.dialog != nil:
		m.dialogClick(mm)
		return nil
	case m.picker != nil:
		return m.pickerClick(mm)
	case m.prompt != nil:
		m.suggestionClick(mm)
		return nil
	}
	m.resetPending() // un clic cancela una secuencia a medias y el panel del líder
	switch m.layout().ZoneAt(mm.X, mm.Y) {
	case ui.ZoneExplorer:
		return m.dispatchByName("explorer.click", command.Args{Count: m.explorerRowAt(mm.Y) + 1, HasCount: true})
	case ui.ZoneEditor:
		return m.editorClick(mm)
	case ui.ZonePreview:
		if m.mdFull() {
			m.focus = FocusEditor
		}
	}
	return nil
}

// editorPos convierte una celda de pantalla en una posición del texto del buffer activo.
func (m *Model) editorPos(mm tea.Mouse) buffer.Pos {
	return ui.PosAt(m.buf, m.layout(), m.top, m.left, mm.X, mm.Y)
}

func (m *Model) editorClick(mm tea.Mouse) tea.Cmd {
	m.focus = FocusEditor
	m.resetPending()
	pos := m.editorPos(mm)
	if m.isDoubleClick(mm.X, mm.Y) {
		a, b := m.wordAt(pos)
		return m.dispatchByName("visual.select_range", command.Args{Line: a.Line, Col: a.Col, Line2: b.Line, Col2: b.Col})
	}
	if m.mode.isVisual() {
		m.setMode(ModeNormal)
	}
	cmd := m.dispatchByName("cursor.move_to", command.Args{Line: pos.Line, Col: pos.Col})
	m.dragging, m.dragAnchor, m.dragCell = true, pos, [2]int{mm.X, mm.Y}
	return cmd
}

// mouseDrag arrastra con el botón pulsado: entra en Visual desde el punto de pulsación hasta el puntero.
func (m *Model) mouseDrag(mm tea.Mouse) tea.Cmd {
	if cell := [2]int{mm.X, mm.Y}; cell == m.dragCell { // varias muestras en la misma celda: no se repite el trabajo
		return nil
	} else {
		m.dragCell = cell
	}
	pos := m.editorPos(mm)
	return m.dispatchByName("visual.select_range", command.Args{
		Line: m.dragAnchor.Line, Col: m.dragAnchor.Col, Line2: pos.Line, Col2: pos.Col,
	})
}

// wordAt devuelve el principio y el final (incluido) de la palabra en pos, con la misma definición que w y b.
// Sobre blancos devuelve solo esa posición.
func (m *Model) wordAt(pos buffer.Pos) (start, end buffer.Pos) {
	cl := buffer.Clusters(m.buf.Line(pos.Line))
	if len(cl) == 0 {
		return pos, pos
	}
	col := min(pos.Col, len(cl)-1)
	c := charClass(cl[col])
	a, b := col, col
	if c != classBlank {
		for a > 0 && charClass(cl[a-1]) == c {
			a--
		}
		for b+1 < len(cl) && charClass(cl[b+1]) == c {
			b++
		}
	}
	return buffer.Pos{Line: pos.Line, Col: a}, buffer.Pos{Line: pos.Line, Col: b}
}

func (m *Model) mouseWheel(mm tea.Mouse) tea.Cmd {
	delta := 0
	switch mm.Button {
	case tea.MouseWheelUp:
		delta = -wheelStep
	case tea.MouseWheelDown:
		delta = wheelStep
	default:
		return nil
	}
	if m.dialog != nil {
		return nil
	}
	if m.picker != nil {
		if g := ui.PickerGeom(m.pickerView(), m.width, m.height); g.Box.Contains(mm.X, mm.Y) {
			return m.dispatchByName("view.scroll", command.Args{Text: "picker", Count: delta, HasCount: true})
		}
		return nil
	}
	pane := ""
	switch m.layout().ZoneAt(mm.X, mm.Y) {
	case ui.ZoneExplorer:
		pane = "explorer"
	case ui.ZoneEditor:
		pane = "editor"
	case ui.ZonePreview:
		pane = "preview"
	default:
		return nil
	}
	return m.dispatchByName("view.scroll", command.Args{Text: pane, Count: delta, HasCount: true})
}

// scrollPane desplaza un panel delta líneas sin cambiar el foco.
func (m *Model) scrollPane(pane string, delta int) {
	switch pane {
	case "editor":
		m.scrollEditor(delta)
	case "preview":
		if m.mdFull() {
			m.previewMove(delta)
		} else {
			m.scrollEditor(delta) // en el modo dividido la vista previa sigue al fuente
		}
	case "explorer":
		if m.explorer != nil && m.showExplorer {
			rows := len(m.explorer.tree.Visible())
			m.explorer.scroll = max(0, min(m.explorer.scroll+delta, max(0, rows-ui.TextRows(m.height))))
		}
	case "picker":
		if p := m.picker; p != nil {
			p.sel = max(0, min(p.sel+delta, max(0, len(p.results)-1)))
		}
	}
}

// scrollEditor desplaza el texto del editor y mantiene el cursor dentro de lo visible (respetando el margen de 3 líneas).
func (m *Model) scrollEditor(delta int) {
	rows, lines := ui.TextRows(m.height), m.buf.LineCount()
	maxTop := max(0, lines-rows)
	m.top = max(0, min(m.top+delta, maxTop))
	margin := min(scrollMargin, (rows-1)/2)
	lo, hi := m.top, m.top+rows-1
	if m.top > 0 {
		lo = m.top + margin
	}
	if m.top < maxTop {
		hi = m.top + rows - 1 - margin
	}
	lo = min(lo, hi)
	if target := max(lo, min(m.cur.Line, hi)); target != m.cur.Line {
		m.cur.Vertical(m.buf, target-m.cur.Line)
	}
}

// explorerRowAt devuelve el índice de la fila del explorador bajo la fila de pantalla y.
func (m *Model) explorerRowAt(y int) int {
	if m.explorer == nil {
		return -1
	}
	return m.explorer.scroll + y
}

// explorerClick: un clic selecciona la fila; en una carpeta la abre o cierra y en un archivo lo abre y pasa el foco al editor.
func (m *Model) explorerClick(row int) {
	if m.explorer == nil || !m.showExplorer {
		return
	}
	rows := m.explorer.tree.Visible()
	m.focus = FocusExplorer
	if row < 0 || row >= len(rows) {
		return
	}
	m.explorer.sel = row
	n := rows[row].Node
	switch {
	case !n.IsDir:
		if m.openPath(n.Path, m.displayName(n.Path)) {
			m.focus = FocusEditor
		}
	case n.Parent == nil: // la raíz no se cierra
	case n.Expanded:
		n.Collapse()
	default:
		if err := m.explorer.tree.Expand(n); err != nil {
			m.msg, m.msgErr = "No se puede leer "+n.Name, true
		}
	}
}

// ---- ventanas superpuestas ----

func (m *Model) dialogClick(mm tea.Mouse) {
	g := ui.DialogGeom(m.dialog.view(), m.width, m.height)
	if i := g.ButtonAt(mm.X, mm.Y); i >= 0 {
		m.chooseDialog(i)
	}
}

func (m *Model) pickerClick(mm tea.Mouse) tea.Cmd {
	p := m.picker
	g := ui.PickerGeom(m.pickerView(), m.width, m.height)
	if !g.Box.Contains(mm.X, mm.Y) { // un clic fuera cierra la ventana sin ejecutar nada
		m.closePicker()
		return nil
	}
	row := g.RowAt(mm.X, mm.Y)
	idx := p.scroll + row
	if row < 0 || idx >= len(p.results) {
		return nil
	}
	if m.isDoubleClick(mm.X, mm.Y) {
		p.sel = idx
		m.pickerAccept()
		return nil
	}
	p.sel = idx
	return nil
}

// suggestionClick: en las sugerencias de «:» un clic elige la sugerencia y un doble clic la ejecuta.
func (m *Model) suggestionClick(mm tea.Mouse) {
	p := m.prompt
	pv := p.suggestionView(m.height, m.width)
	if pv == nil {
		return
	}
	g := ui.PickerGeom(pv, m.width, m.height)
	row := g.RowAt(mm.X, mm.Y)
	idx := p.scroll + row
	if row < 0 || idx >= len(p.sugg) {
		return
	}
	p.suggSel = idx
	if m.isDoubleClick(mm.X, mm.Y) {
		name := palette()[p.sugg[idx].Index].data
		m.exHistory = appendHistory(m.exHistory, p.String())
		m.closePrompt()
		m.runPaletteCommand(name)
	}
}

func registerMouseCommands(r *command.Registry[*Model]) {
	reg := func(name, desc string, keys []string, run func(m *Model, a command.Args)) {
		r.MustRegister(command.Command[*Model]{Name: name, Desc: desc, Short: shortLabels[name], Keys: keys,
			Run: func(m *Model, a command.Args) tea.Cmd { run(m, a); return nil }})
	}
	reg("cursor.move_to", "Mover el cursor a una posición (clic del ratón)", []string{"clic"}, func(m *Model, a command.Args) {
		m.cur.Set(m.buf, buffer.Pos{Line: a.Line, Col: a.Col})
	})
	reg("visual.select_range", "Seleccionar un rango en modo Visual (arrastre o doble clic)", []string{"arrastrar", "doble clic"}, func(m *Model, a command.Args) {
		m.anchor = m.buf.Clamp(buffer.Pos{Line: a.Line, Col: a.Col})
		m.cur.Set(m.buf, buffer.Pos{Line: a.Line2, Col: a.Col2})
		if m.mode != ModeVisual {
			m.setMode(ModeVisual)
		}
	})
	reg("view.scroll", "Desplazar un panel (rueda del ratón)", []string{"rueda"}, func(m *Model, a command.Args) {
		m.scrollPane(a.Text, a.Count)
	})
	reg("explorer.click", "Explorador: clic en una fila (abre archivo o expande carpeta)", []string{"clic"}, func(m *Model, a command.Args) {
		m.explorerClick(a.Count - 1)
	})
	reg("mouse.toggle", "Activar o desactivar el ratón (para usar la selección nativa del terminal)", []string{":mouse"}, func(m *Model, _ command.Args) {
		m.mouseOn = !m.mouseOn
		m.dragging = false
		if m.mouseOn {
			m.msg = "Ratón activado"
		} else {
			m.msg = "Ratón desactivado: selección nativa del terminal"
		}
	})
}
