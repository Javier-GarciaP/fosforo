// Package editor contiene el estado y la lógica del editor: modos, teclas, comandos y vista.
package editor

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/keymap"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
	"github.com/Javier-GarciaP/fosforo/internal/version"
)

// Tamaño mínimo de terminal en el que se dibuja la interfaz.
const (
	MinWidth  = 40
	MinHeight = 10
)

// scrollMargin son las líneas que se mantienen visibles entre el cursor y el borde al desplazarse.
const scrollMargin = 3

// DefaultTimeout es lo que se espera la siguiente tecla de una secuencia incompleta (p. ej. «g» antes de «gg»).
const DefaultTimeout = 1000 * time.Millisecond

// TimeoutMsg lo envía el temporizador de una secuencia incompleta. Seq identifica la secuencia: si ya pasó otra
// tecla, el mensaje es viejo y se ignora.
type TimeoutMsg struct{ Seq int }

// Model es el modelo de Bubble Tea del editor.
type Model struct {
	bufs          []*bufState // buffers abiertos; el activo también vive en buf/cur/top/left/label
	active        int
	buf           *buffer.Buffer
	cur           buffer.Cursor
	top, left     int // primera línea y primera columna visual visibles
	width, height int
	label         string // nombre mostrado en la statusline
	msg           string
	msgErr        bool
	dialog        *dialog
	quit          bool
	batchSave     bool // se guardan varios buffers seguidos: un conflicto con el disco aborta en lugar de preguntar

	root         string // carpeta raíz del explorador
	explorer     *explorerState
	showExplorer bool
	focus        Focus
	icons        ui.IconMode

	mode    Mode
	anchor  buffer.Pos // origen de la selección en Visual
	prompt  *prompt
	pending []string // teclas de una secuencia incompleta
	count   int      // conteo escrito (0 = ninguno)
	op      string   // operador pendiente: "d", "c" o "y"
	opCount int      // conteo escrito antes del operador
	seq     int      // se incrementa con cada tecla; invalida temporizadores viejos
	timeout time.Duration
	hlDelay time.Duration
	hlJobs  *atomic.Int64
	light   bool // tema claro

	reg           register
	clipboard     Clipboard
	exHistory     []string
	searchHistory []string
	lastSearch    string
	lastForward   bool
}

// Option configura el modelo.
type Option func(*Model)

// WithTimeout cambia la espera de las secuencias incompletas.
func WithTimeout(d time.Duration) Option { return func(m *Model) { m.timeout = d } }

// WithClipboard cambia cómo se copia al portapapeles (por defecto OSC 52 y la utilidad del sistema).
func WithClipboard(c Clipboard) Option { return func(m *Model) { m.clipboard = c } }

// New crea un modelo sobre un buffer ya cargado; label es el nombre que se muestra. Empieza en modo Normal.
func New(b *buffer.Buffer, label string, opts ...Option) Model {
	m := Model{
		buf: b, cur: buffer.NewCursor(buffer.Pos{}), label: label,
		timeout: DefaultTimeout, clipboard: defaultClipboard, lastForward: true,
		hlDelay: DefaultHighlightDelay, hlJobs: new(atomic.Int64),
	}
	m.bufs = []*bufState{newBufState(b, label)}
	for _, o := range opts {
		o(&m)
	}
	if ro, why := b.ReadOnly(); ro {
		m.msg, m.msgErr = "Solo lectura: "+why, true
	}
	return m
}

// Open abre path en un modelo nuevo. Sin ruta, empieza con un buffer vacío sin nombre.
// Si el archivo no existe se muestra vacío y se crea al guardar.
func Open(path string, opts ...Option) (Model, error) {
	if path == "" {
		return New(buffer.New(""), "", opts...), nil
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		// Una carpeta abre el explorador sobre ella, con el foco, y un buffer vacío.
		m := New(buffer.New(""), "", append(opts, WithRoot(path))...)
		m.toggleExplorer()
		return m, nil
	}
	b, err := buffer.Open(path)
	if err != nil {
		return Model{}, err
	}
	return New(b, path, opts...), nil
}

// Accesores para la interfaz y las pruebas.
func (m Model) Buffer() *buffer.Buffer { return m.buf }
func (m Model) Cursor() buffer.Pos     { return m.cur.Pos }
func (m Model) Top() int               { return m.top }
func (m Model) Left() int              { return m.left }
func (m Model) Mode() Mode             { return m.mode }
func (m Model) Message() string        { return m.msg }
func (m Model) DialogOpen() bool       { return m.dialog != nil }
func (m Model) Quitting() bool         { return m.quit }
func (m Model) Seq() int               { return m.seq }

// Pending devuelve las teclas pendientes tal como se muestran en la línea de comandos (p. ej. «3d», «g»).
func (m Model) Pending() string {
	var sb strings.Builder
	if m.opCount > 0 {
		sb.WriteString(strconv.Itoa(m.opCount))
	}
	sb.WriteString(m.op)
	if m.count > 0 {
		sb.WriteString(strconv.Itoa(m.count))
	}
	sb.WriteString(strings.Join(m.pending, ""))
	return sb.String()
}

// Init no lanza comandos iniciales.
func (m Model) Init() tea.Cmd { return nil }

// Update procesa mensajes: redimensionado, teclas, pegado y temporizador.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		cmd = m.handleKey(msg)
	case tea.PasteMsg:
		switch {
		case m.dialog != nil:
		case m.mode == ModeCommand:
			m.prompt.insert(msg.Content)
		case m.mode == ModeInsert:
			m.msg, m.msgErr = "", false
			m.withUndo(func() { m.insertText(msg.Content) })
		}
	case TimeoutMsg:
		if msg.Seq == m.seq && len(m.pending) > 0 {
			m.resetPending() // la secuencia incompleta se descarta sin cambiar el texto
		}
	case tea.FocusMsg: // la terminal recupera el foco: se revisan el disco y el explorador
		m.checkExternal()
		m.refreshExplorer()
	case hlDebounceMsg:
		cmd = m.highlightDebounce(msg)
	case hlDoneMsg:
		m.highlightDone(msg)
	}
	m.clampCursor()
	m.scroll()
	m.scrollExplorer()
	// Cada mensaje puede haber cambiado el texto: se programa (con debounce) retokenizar lo que haga falta.
	// Si acaba de terminar una tokenización y el texto ya cambió, se vuelve a programar.
	if next := m.scheduleHighlight(); next != nil {
		cmd = tea.Batch(cmd, next)
	}
	if m.quit {
		cmd = tea.Sequence(cmd, tea.Quit)
	}
	return m, cmd
}

// handleKey envía una tecla al diálogo, a la línea de comandos o al modo actual.
func (m *Model) handleKey(k tea.KeyPressMsg) tea.Cmd {
	if m.dialog != nil {
		return m.dialogKey(k)
	}
	m.seq++
	m.msg, m.msgErr = "", false
	if m.mode == ModeCommand {
		m.promptKey(k)
		return nil
	}
	return m.modalKey(k)
}

func keyName(k tea.KeyPressMsg) string {
	if s := k.String(); s != " " {
		return s
	}
	return "space"
}

func isCountKey(key string, count int) bool {
	return len(key) == 1 && key[0] >= '0' && key[0] <= '9' && (key != "0" || count > 0)
}

// modalKey resuelve una tecla según el modo: conteos, operadores pendientes y secuencias del keymap.
func (m *Model) modalKey(k tea.KeyPressMsg) tea.Cmd {
	key := keyName(k)
	if key == "esc" || key == "ctrl+c" { // Esc actúa de inmediato, sin esperar ni resolver secuencias
		return m.dispatchByName("mode.normal", command.Args{})
	}
	mode := m.mode.keymapName()
	switch {
	case m.op != "":
		mode = modeOperator
	case m.focusInExplorer():
		mode = modeExplorer
	}
	if m.mode != ModeInsert && len(m.pending) == 0 && isCountKey(key, m.count) {
		m.count = m.count*10 + int(key[0]-'0')
		return nil
	}
	if m.op != "" && len(m.pending) == 0 && key == m.op { // dd, cc, yy
		op, total := m.op, max(1, m.opCount)*max(1, m.count)
		m.resetPending()
		var cmd tea.Cmd
		m.withUndo(func() { cmd = m.operatorLines(op, total) })
		return cmd
	}
	res := defaultKeymap.Resolve(mode, m.pending, key)
	switch res.Status {
	case keymap.Pending:
		m.pending = append(m.pending[:len(m.pending):len(m.pending)], key)
		return m.armTimeout()
	case keymap.Complete:
		args := command.Args{Count: m.count, HasCount: m.count > 0}
		m.pending, m.count = nil, 0
		return m.dispatchResolved(res.Command, args)
	}
	if m.mode == ModeInsert && k.Text != "" && k.Mod&(tea.ModCtrl|tea.ModAlt) == 0 {
		m.insertText(k.Text) // en Insert, lo que no es un atajo se escribe
		return nil
	}
	m.resetPending() // tecla sin función: se ignora y no escribe nada
	return nil
}

// armTimeout programa el descarte de una secuencia incompleta.
func (m *Model) armTimeout() tea.Cmd {
	seq := m.seq
	return tea.Tick(m.timeout, func(time.Time) tea.Msg { return TimeoutMsg{Seq: seq} })
}

func (m *Model) resetPending() {
	m.pending, m.count, m.op, m.opCount = nil, 0, "", 0
}

// dispatchResolved ejecuta el comando al que llevó una secuencia; con un operador pendiente solo valen los movimientos.
func (m *Model) dispatchResolved(name string, a command.Args) tea.Cmd {
	if m.op == "" {
		return m.dispatchByName(name, a)
	}
	op, opCount := m.op, m.opCount
	def, ok := motionByCommand(name)
	var cmd tea.Cmd
	if ok {
		total := max(1, opCount) * max(1, a.Count)
		m.withUndo(func() { cmd = m.operatorWithMotion(op, def, total, opCount > 0 || a.HasCount) })
	}
	m.op, m.opCount = "", 0
	return cmd
}

// dispatchByName ejecuta un comando del registro; es el único camino de las teclas, de «:» y (más adelante) de la paleta.
func (m *Model) dispatchByName(name string, a command.Args) tea.Cmd {
	var cmd tea.Cmd
	var err error
	m.withUndo(func() { cmd, err = registry.Exec(m, name, a) })
	if err != nil {
		m.fail(err)
	}
	return cmd
}

// withUndo agrupa en un solo paso de deshacer todo lo que haga fn. Si fn deja el editor en Insert, el grupo sigue
// abierto hasta salir de Insert: una sesión de escritura entera es un único paso.
func (m *Model) withUndo(fn func()) {
	if m.mode != ModeInsert {
		m.buf.BeginGroup(m.cur.Pos)
	}
	fn()
	if m.mode != ModeInsert {
		m.buf.EndGroup()
	}
}

// setMode cambia de modo; fuera de Insert el cursor no puede quedar después del último carácter.
func (m *Model) setMode(mode Mode) {
	m.mode = mode
	m.clampCursor()
}

func (m *Model) clampCursor() {
	m.cur.Pos = m.buf.Clamp(m.cur.Pos)
	if m.mode == ModeNormal || m.mode.isVisual() {
		if last := max(0, m.buf.LineLen(m.cur.Line)-1); m.cur.Col > last {
			m.cur.Pos.Col = last
		}
	}
}

func (m *Model) reportSaveError(err error) {
	switch {
	case errors.Is(err, buffer.ErrNoPath):
		m.msg, m.msgErr = "E32: sin nombre de archivo", true
	case errors.Is(err, buffer.ErrReadOnly):
		m.msg, m.msgErr = m.readOnlyMessage(), true
	default:
		m.msg, m.msgErr = "E212: no se puede guardar: "+err.Error(), true
	}
}

func (m *Model) readOnlyMessage() string {
	_, why := m.buf.ReadOnly()
	return "Solo lectura (" + why + "): no se puede modificar"
}

// edit atiende el resultado de una modificación del buffer (aviso de solo lectura o error). Devuelve si se aplicó.
func (m *Model) edit(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, buffer.ErrReadOnly) {
		m.msg, m.msgErr = m.readOnlyMessage(), true
	} else {
		m.msg, m.msgErr = "No se pudo editar: "+err.Error(), true
	}
	return false
}

func (m *Model) insertText(text string) {
	_, end, err := m.buf.Insert(m.cur.Pos, text)
	if m.edit(err) {
		m.cur.Set(m.buf, end)
	}
}

func (m *Model) backspace() {
	prev, ok := m.buf.Prev(m.cur.Pos)
	if !ok {
		return
	}
	if _, err := m.buf.Delete(prev, m.cur.Pos); m.edit(err) {
		m.cur.Set(m.buf, prev)
	}
}

func (m *Model) deleteForward() {
	next, ok := m.buf.Next(m.cur.Pos)
	if !ok {
		return
	}
	_, err := m.buf.Delete(m.cur.Pos, next)
	m.edit(err)
}

// indentBeforeCursor devuelve el sangrado (espacios y tabuladores iniciales) de la línea actual hasta el cursor.
func (m Model) indentBeforeCursor() string {
	var sb strings.Builder
	for i, c := range buffer.Clusters(m.buf.Line(m.cur.Line)) {
		if i >= m.cur.Col || (c != " " && c != "\t") {
			break
		}
		sb.WriteString(c)
	}
	return sb.String()
}

func (m Model) pageSize() int { return max(1, ui.TextRows(m.height)-1) }

// ---- vista ----

// scroll ajusta la vista para que el cursor quede visible, con margen vertical y sin partir líneas.
func (m *Model) scroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	lines := m.buf.LineCount()
	rows := ui.TextRows(m.height)
	margin := min(scrollMargin, (rows-1)/2)
	if m.cur.Line < m.top+margin {
		m.top = m.cur.Line - margin
	}
	if m.cur.Line > m.top+rows-1-margin {
		m.top = m.cur.Line - (rows - 1 - margin)
	}
	m.top = max(0, min(m.top, lines-rows))

	cols := ui.TextCols(m.editorCols(), lines)
	vc := m.cur.VisualCol(m.buf)
	if vc < m.left {
		m.left = vc - cols/2 // al volver atrás se centra el cursor, o se vuelve al inicio si la línea cabe
	}
	if vc >= m.left+cols {
		m.left = vc - cols + 1
	}
	m.left = max(0, m.left)
}

func (m Model) currentTheme() ui.Theme {
	if m.light {
		return ui.Light()
	}
	return ui.Dark()
}

// toggleTheme alterna entre el tema oscuro y el claro. Los tramos de sintaxis no cambian: solo cambian sus colores.
func (m *Model) toggleTheme() {
	m.light = !m.light
	if m.light {
		m.msg = "Tema claro"
	} else {
		m.msg = "Tema oscuro"
	}
}

// TooSmall indica si la terminal es menor que el mínimo. Mientras no se conoce el tamaño (0x0) no se avisa.
func (m Model) TooSmall() bool {
	if m.width == 0 && m.height == 0 {
		return false
	}
	return m.width < MinWidth || m.height < MinHeight
}

func (m Model) view() ui.EditorView {
	mode := m.mode.Label()
	if m.focusInExplorer() && m.mode == ModeNormal {
		mode = "EXPLORER"
	}
	v := ui.EditorView{
		Buf: m.buf, Cursor: m.cur.Pos, Top: m.top, Left: m.left,
		Mode: mode, PathLabel: m.label, Message: m.msg, IsError: m.msgErr, Pending: m.Pending(),
		Explorer: m.explorerView(), BufIndex: m.active + 1, BufCount: len(m.bufs),
	}
	if hl := m.bufs[m.active].hl; hl != nil {
		v.Highlight = hl.lines
	}
	if m.mode.isVisual() {
		r := m.visualRange()
		end := r.end
		if !r.linewise && end.Col > 0 {
			end.Col--
		}
		v.Selection = &ui.Selection{Start: r.start, End: end, Linewise: r.linewise}
		if lines := r.end.Line - r.start.Line + 1; lines > 1 {
			v.Info = fmt.Sprintf("%d líneas", lines)
		} else {
			v.Info = fmt.Sprintf("%d car.", max(1, buffer.Count(m.buf.Range(r.start, r.end))))
		}
	}
	if m.prompt != nil {
		v.Prompt = &ui.Prompt{Prefix: m.prompt.prefix, Text: m.prompt.String(), Cursor: m.prompt.cur}
	}
	if m.dialog != nil {
		v.Dialog = m.dialog.view()
	}
	return v
}

// View dibuja el editor o, si la terminal es muy pequeña, el aviso de tamaño mínimo.
func (m Model) View() tea.View {
	var v tea.View
	switch {
	case m.width == 0 || m.height == 0:
		v = tea.NewView("")
	case m.TooSmall():
		v = tea.NewView(fmt.Sprintf("Terminal demasiado pequeña (%dx%d).\nMínimo: %dx%d.", m.width, m.height, MinWidth, MinHeight))
	default:
		ev := m.view()
		v = tea.NewView(ui.RenderEditor(ev, m.width, m.height, m.currentTheme()))
		if x, y, ok := ui.CursorCell(ev, m.width, m.height); ok {
			c := tea.NewCursor(x, y)
			c.Shape = tea.CursorBlock
			if m.mode == ModeInsert || m.mode == ModeCommand {
				c.Shape = tea.CursorBar
			}
			v.Cursor = c
		}
	}
	v.AltScreen = true
	v.ReportFocus = true
	v.WindowTitle = m.label0() + " - " + version.Name
	return v
}
