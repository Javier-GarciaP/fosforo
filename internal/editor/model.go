// Package editor contiene el estado y la lógica del editor.
package editor

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
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

// ModeInsert es la etiqueta del único modo por ahora: se escribe directamente (modal-editing añade el resto).
const ModeInsert = "INSERT"

// Model es el modelo de Bubble Tea del editor.
type Model struct {
	buf           *buffer.Buffer
	cur           buffer.Cursor
	top, left     int // primera línea y primera columna visual visibles
	width, height int
	label         string // nombre mostrado en la statusline
	msg           string
	msgErr        bool
	dialog        *dialog
	theme         ui.Theme
}

// dialog es el estado del diálogo «Cambios sin guardar».
type dialog struct{ selected int }

var dialogButtons = []string{"Guardar (y)", "Descartar (n)", "Cancelar (Esc)"}

// New crea un modelo sobre un buffer ya cargado; label es el nombre que se muestra.
func New(b *buffer.Buffer, label string) Model {
	m := Model{buf: b, cur: buffer.NewCursor(buffer.Pos{}), label: label, theme: ui.Dark()}
	if ro, why := b.ReadOnly(); ro {
		m.msg, m.msgErr = "Solo lectura: "+why, true
	}
	return m
}

// Open abre path en un modelo nuevo. Sin ruta, empieza con un buffer vacío sin nombre.
// Si el archivo no existe se muestra vacío y se crea al guardar.
func Open(path string) (Model, error) {
	if path == "" {
		return New(buffer.New(""), ""), nil
	}
	b, err := buffer.Open(path)
	if err != nil {
		return Model{}, err
	}
	return New(b, path), nil
}

// Buffer devuelve el buffer en edición.
func (m Model) Buffer() *buffer.Buffer { return m.buf }

// Cursor devuelve la posición del cursor.
func (m Model) Cursor() buffer.Pos { return m.cur.Pos }

// Top devuelve la primera línea visible y Left la primera columna visual visible.
func (m Model) Top() int  { return m.top }
func (m Model) Left() int { return m.left }

// Message devuelve el mensaje de la línea de comandos.
func (m Model) Message() string { return m.msg }

// DialogOpen indica si el diálogo de cambios sin guardar está abierto.
func (m Model) DialogOpen() bool { return m.dialog != nil }

// Init no lanza comandos iniciales.
func (m Model) Init() tea.Cmd { return nil }

// Update procesa mensajes: redimensionado, teclas y pegado.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		return m.key(msg)
	case tea.PasteMsg:
		m.msg, m.msgErr = "", false
		m.insert(msg.Content)
	}
	m.scroll()
	return m, nil
}

func (m Model) key(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.msg, m.msgErr = "", false
	if m.dialog != nil {
		return m.dialogKey(k)
	}
	switch k.String() {
	case "ctrl+s":
		m.save()
	case "ctrl+q", "ctrl+c":
		if m.buf.Modified() {
			m.dialog = &dialog{}
		} else {
			return m, tea.Quit
		}
	case "left":
		m.cur.Left(m.buf)
	case "right":
		m.cur.Right(m.buf)
	case "up":
		m.cur.Up(m.buf)
	case "down":
		m.cur.Down(m.buf)
	case "home":
		m.cur.Home(m.buf)
	case "end":
		m.cur.End(m.buf)
	case "pgup":
		m.cur.PageUp(m.buf, m.pageSize())
	case "pgdown":
		m.cur.PageDown(m.buf, m.pageSize())
	case "enter":
		m.insert("\n" + m.indentBeforeCursor())
	case "tab":
		m.insert("\t")
	case "backspace":
		m.backspace()
	case "delete":
		m.deleteForward()
	default:
		if k.Text != "" && k.Mod&(tea.ModCtrl|tea.ModAlt) == 0 {
			m.insert(k.Text)
		}
	}
	m.scroll()
	return m, nil
}

func (m Model) dialogKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "left", "shift+tab":
		m.dialog.selected = (m.dialog.selected + len(dialogButtons) - 1) % len(dialogButtons)
	case "right", "tab":
		m.dialog.selected = (m.dialog.selected + 1) % len(dialogButtons)
	case "y":
		return m.saveAndQuit()
	case "n":
		return m, tea.Quit
	case "esc", "ctrl+c":
		m.dialog = nil
	case "enter":
		switch m.dialog.selected {
		case 0:
			return m.saveAndQuit()
		case 1:
			return m, tea.Quit
		default:
			m.dialog = nil
		}
	}
	return m, nil
}

func (m Model) saveAndQuit() (tea.Model, tea.Cmd) {
	if !m.save() {
		m.dialog = nil // el error se muestra en la línea de comandos y no se sale
		return m, nil
	}
	return m, tea.Quit
}

// save guarda el buffer y deja el resultado en la línea de comandos. Devuelve si se guardó.
func (m *Model) save() bool {
	err := m.buf.Save()
	switch {
	case err == nil:
		m.msg = fmt.Sprintf("%q %dL escritas", m.label, m.buf.LineCount())
		return true
	case errors.Is(err, buffer.ErrNoPath):
		m.msg, m.msgErr = "E32: sin nombre de archivo", true
	case errors.Is(err, buffer.ErrReadOnly):
		m.msg, m.msgErr = m.readOnlyMessage(), true
	default:
		m.msg, m.msgErr = "E212: no se puede guardar: "+err.Error(), true
	}
	return false
}

func (m *Model) readOnlyMessage() string {
	_, why := m.buf.ReadOnly()
	return "Solo lectura (" + why + "): no se puede modificar"
}

// edit ejecuta una modificación del buffer y atiende el caso de solo lectura. Devuelve si se aplicó.
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

func (m *Model) insert(text string) {
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

	cols := ui.TextCols(m.width, lines)
	vc := m.cur.VisualCol(m.buf)
	if vc < m.left {
		m.left = vc - cols/2 // al volver atrás se centra el cursor, o se vuelve al inicio si la línea cabe
	}
	if vc >= m.left+cols {
		m.left = vc - cols + 1
	}
	m.left = max(0, m.left)
}

// TooSmall indica si la terminal es menor que el mínimo. Mientras no se conoce el tamaño (0x0) no se avisa.
func (m Model) TooSmall() bool {
	if m.width == 0 && m.height == 0 {
		return false
	}
	return m.width < MinWidth || m.height < MinHeight
}

func (m Model) view() ui.EditorView {
	v := ui.EditorView{
		Buf: m.buf, Cursor: m.cur.Pos, Top: m.top, Left: m.left,
		Mode: ModeInsert, PathLabel: m.label, Message: m.msg, IsError: m.msgErr,
	}
	if m.dialog != nil {
		name := m.label
		if name == "" {
			name = "[Sin nombre]"
		}
		v.Dialog = &ui.Dialog{
			Title:    "Cambios sin guardar",
			Lines:    []string{fmt.Sprintf("%q tiene cambios sin guardar.", name), "¿Qué quieres hacer?"},
			Buttons:  dialogButtons,
			Selected: m.dialog.selected,
			Hint:     "←→ elegir   ⏎ confirmar",
		}
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
		v = tea.NewView(ui.RenderEditor(ev, m.width, m.height, m.theme))
		if x, y, ok := ui.CursorCell(ev, m.width, m.height); ok {
			c := tea.NewCursor(x, y)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	}
	v.AltScreen = true
	name := m.label
	if name == "" {
		name = "[Sin nombre]"
	}
	v.WindowTitle = name + " - " + version.Name
	return v
}
