package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/markdown"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// DefaultMarkdownDelay es la espera tras la última edición antes de volver a renderizar la vista previa dividida (debounce).
const DefaultMarkdownDelay = 150 * time.Millisecond

// syncMarkdownBytes: los documentos de hasta este tamaño se renderizan sin salir del bucle de eventos (es instantáneo).
const syncMarkdownBytes = 64 << 10

// modePreview es el contexto de teclas de la vista previa a pantalla completa.
const modePreview = "preview"

// mdView es cómo se ve un documento Markdown.
type mdView int

const (
	mdSource  mdView = iota // solo el fuente (edición)
	mdPreview               // vista previa a pantalla completa, solo lectura
	mdSplit                 // fuente a la izquierda y vista previa a la derecha
)

// mdState es el estado de Markdown de un buffer que es un documento .md o .markdown.
type mdState struct {
	view   mdView
	scroll int // primera línea visible de la vista previa a pantalla completa
	tooBig bool

	doc       *markdown.Doc // último renderizado
	docRev    int           // revisión del texto de doc
	docWidth  int           // ancho para el que se renderizó
	requested [2]int        // (revisión, ancho) para los que ya se programó un renderizado
	size      int
	running   bool
}

func isMarkdownPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return true
	}
	return false
}

// initMarkdown prepara el estado de Markdown de un buffer. Los documentos .md se abren en vista previa; si superan los
// 2 MiB se abren como fuente y se avisa.
func (bs *bufState) initMarkdown(preview bool) {
	if !isMarkdownPath(bs.buf.Path()) {
		bs.md = nil
		return
	}
	md := &mdState{requested: [2]int{-1, -1}, docRev: -1}
	if bs.md != nil { // se conserva cómo se estaba viendo (p. ej. tras recargar o renombrar)
		md.view, md.scroll = bs.md.view, bs.md.scroll
	} else if preview {
		md.view = mdPreview
	}
	if bs.buf.Rev() >= 0 && len(lfText(bs.buf)) > markdown.MaxBytes {
		md.tooBig, md.view = true, mdSource
	}
	bs.md = md
}

// WithMarkdownDelay cambia la espera antes de volver a renderizar la vista previa dividida tras una edición.
func WithMarkdownDelay(d time.Duration) Option { return func(m *Model) { m.mdDelay = d } }

// MarkdownJobs devuelve cuántos renderizados se han lanzado en segundo plano (para las pruebas).
func (m Model) MarkdownJobs() int { return int(m.mdJobs.Load()) }

// ViewKind describe cómo se ve el buffer activo: «source», «preview» o «split» ("" si no es Markdown).
func (m Model) ViewKind() string {
	switch md := m.bufs[m.active].md; {
	case md == nil:
		return ""
	case md.view == mdPreview:
		return "preview"
	case md.view == mdSplit:
		return "split"
	}
	return "source"
}

// PreviewScroll devuelve la primera línea visible de la vista previa del buffer activo.
func (m Model) PreviewScroll() int { return m.previewScroll() }

// PreviewDoc devuelve el documento renderizado que se está mostrando (nil si todavía no hay).
func (m Model) PreviewDoc() *markdown.Doc {
	if md := m.bufs[m.active].md; md != nil {
		return md.doc
	}
	return nil
}

func (m *Model) mdState() *mdState { return m.bufs[m.active].md }

// mdFull indica que el buffer activo se ve como vista previa a pantalla completa.
func (m *Model) mdFull() bool {
	md := m.mdState()
	return md != nil && md.view == mdPreview
}

func (m *Model) mdSplit() bool {
	md := m.mdState()
	return md != nil && md.view == mdSplit
}

// mdWidth es el ancho al que se ajusta el texto en la vista previa actual.
func (m *Model) mdWidth() int {
	return ui.MarkdownContentWidth(ui.PreviewWidth(m.width, m.showExplorer && m.explorer != nil, true, m.mdSplit()))
}

func (m *Model) previewRows() int { return ui.TextRows(m.height) }

func (m *Model) previewMax() int {
	md := m.mdState()
	if md == nil || md.doc == nil {
		return 0
	}
	return max(0, len(md.doc.Lines)-m.previewRows())
}

// previewScroll es la primera línea visible de la vista previa: la propia en pantalla completa, o la que corresponde
// al scroll del fuente en el modo dividido.
func (m *Model) previewScroll() int {
	md := m.mdState()
	if md == nil || md.doc == nil {
		return 0
	}
	if md.view == mdSplit {
		return min(md.doc.OutLine(m.top), m.previewMax())
	}
	return max(0, min(md.scroll, m.previewMax()))
}

// scheduleMarkdown renderiza (o programa renderizar) la vista previa del buffer activo si su texto o su ancho cambiaron.
// A pantalla completa y al cambiar el ancho se renderiza enseguida; con ediciones en el modo dividido se espera el debounce.
func (m *Model) scheduleMarkdown() tea.Cmd {
	bs := m.bufs[m.active]
	md := bs.md
	if md == nil || md.view == mdSource || m.width == 0 {
		return nil
	}
	rev, w := bs.buf.Rev(), m.mdWidth()
	if md.doc != nil && md.docRev == rev && md.docWidth == w {
		return nil
	}
	if md.requested == [2]int{rev, w} {
		return nil
	}
	md.requested = [2]int{rev, w}
	immediate := md.doc == nil || md.docWidth != w || md.view == mdPreview
	if immediate {
		text := lfText(bs.buf)
		if len(text) <= syncMarkdownBytes {
			md.doc, md.docRev, md.docWidth = renderDoc(text, w), rev, w
			return nil
		}
		return m.startMarkdownJob(bs, rev, w)
	}
	id := bs.id
	return tea.Tick(m.mdDelay, func(time.Time) tea.Msg { return mdDebounceMsg{id: id, rev: rev, width: w} })
}

func renderDoc(text string, w int) *markdown.Doc {
	d := markdown.Render(text, w)
	return &d
}

type mdDebounceMsg struct {
	id         int64
	rev, width int
}

type mdDoneMsg struct {
	id         int64
	rev, width int
	doc        *markdown.Doc
}

func (m *Model) startMarkdownJob(bs *bufState, rev, w int) tea.Cmd {
	if bs.md.running {
		bs.md.requested = [2]int{-1, -1} // se reintentará al terminar el renderizado en curso
		return nil
	}
	bs.md.running = true
	m.mdJobs.Add(1)
	text, id := lfText(bs.buf), bs.id
	return func() tea.Msg { return mdDoneMsg{id: id, rev: rev, width: w, doc: renderDoc(text, w)} }
}

func (m *Model) markdownDebounce(msg mdDebounceMsg) tea.Cmd {
	bs := m.bufByID(msg.id)
	if bs == nil || bs.md == nil || bs.buf.Rev() != msg.rev || m.mdWidth() != msg.width || bs.md.view == mdSource {
		return nil // hay ediciones más recientes (otro debounce se ocupará) o ya no se ve la vista previa
	}
	return m.startMarkdownJob(bs, msg.rev, msg.width)
}

func (m *Model) markdownDone(msg mdDoneMsg) {
	bs := m.bufByID(msg.id)
	if bs == nil || bs.md == nil {
		return
	}
	bs.md.running = false
	if bs.buf.Rev() == msg.rev {
		bs.md.doc, bs.md.docRev, bs.md.docWidth = msg.doc, msg.rev, msg.width
	}
}

// ---- comandos ----

const msgNotMarkdown = "Este archivo no es Markdown"

func (m *Model) markdownAvailable() bool {
	md := m.mdState()
	switch {
	case md == nil:
		m.msg, m.msgErr = msgNotMarkdown, true
	case md.tooBig:
		m.msg, m.msgErr = "Vista previa desactivada: el documento supera los 2 MiB", true
	default:
		return true
	}
	return false
}

// markdownToggle alterna entre la vista previa y el fuente.
func (m *Model) markdownToggle() {
	if !m.markdownAvailable() {
		return
	}
	md := m.mdState()
	if md.view == mdPreview {
		m.previewToSource()
		return
	}
	m.sourceToPreview()
}

// previewToSource pasa a editar el fuente con el cursor en el encabezado más cercano a lo que se estaba viendo.
func (m *Model) previewToSource() {
	md := m.mdState()
	scroll := m.previewScroll()
	line := 0
	if md.doc != nil {
		if src, ok := md.doc.HeadingAt(scroll); ok {
			line = src
		} else {
			line = md.doc.SrcLine(scroll)
		}
	}
	md.view = mdSource
	m.focus = FocusEditor
	m.setMode(ModeNormal)
	line = max(0, min(line, m.buf.LineCount()-1))
	m.cur.Set(m.buf, buffer.Pos{Line: line, Col: firstNonBlank(m.buf.Line(line))})
	m.top = max(0, line-scrollMargin)
}

// sourceToPreview pasa a la vista previa a pantalla completa mostrando el bloque donde estaba el cursor.
func (m *Model) sourceToPreview() {
	md := m.mdState()
	m.resetPending()
	m.setMode(ModeNormal)
	m.ensureDocNow()
	scroll := 0
	if md.doc != nil {
		scroll = md.doc.OutLine(m.cur.Line)
	}
	md.view, md.scroll = mdPreview, scroll
	m.focus = FocusEditor
	m.clampPreview()
}

// markdownSplit muestra u oculta la vista previa al lado del fuente.
func (m *Model) markdownSplit() {
	if !m.markdownAvailable() {
		return
	}
	md := m.mdState()
	if md.view == mdSplit {
		md.view = mdSource
		return
	}
	md.view = mdSplit
	m.ensureDocNow()
}

// ensureDocNow renderiza ya (sin esperar) la vista previa si es pequeña y está desactualizada.
func (m *Model) ensureDocNow() {
	bs := m.bufs[m.active]
	md := bs.md
	if md == nil || m.width == 0 {
		return
	}
	rev, w := bs.buf.Rev(), m.mdWidth()
	if md.doc != nil && md.docRev == rev && md.docWidth == w {
		return
	}
	if text := lfText(bs.buf); len(text) <= syncMarkdownBytes {
		md.doc, md.docRev, md.docWidth = renderDoc(text, w), rev, w
		md.requested = [2]int{rev, w}
	}
}

func (m *Model) clampPreview() {
	if md := m.mdState(); md != nil {
		md.scroll = max(0, min(md.scroll, m.previewMax()))
	}
}

// previewMove desplaza la vista previa n líneas (negativo, hacia arriba).
func (m *Model) previewMove(n int) {
	if md := m.mdState(); md != nil {
		md.scroll = md.scroll + n
		m.clampPreview()
	}
}

func registerMarkdownCommands(r *command.Registry[*Model]) {
	reg := func(name, desc string, keys []string, run func(m *Model, a command.Args)) {
		r.MustRegister(command.Command[*Model]{Name: name, Desc: desc, Short: shortLabels[name], Keys: keys,
			Run: func(m *Model, a command.Args) tea.Cmd { run(m, a); return nil }})
	}
	reg("markdown.toggle", "Alternar vista previa y fuente de Markdown", []string{"e (en la vista previa)", "Space m p"}, func(m *Model, _ command.Args) { m.markdownToggle() })
	reg("markdown.split", "Dividir fuente y vista previa de Markdown", []string{"Space m s"}, func(m *Model, _ command.Args) { m.markdownSplit() })
	onPreview := func(f func(m *Model, a command.Args)) func(m *Model, a command.Args) {
		return func(m *Model, a command.Args) {
			if m.mdState() == nil || m.mdState().view != mdPreview {
				return // lanzado desde la paleta fuera de una vista previa
			}
			f(m, a)
		}
	}
	reg("preview.down", "Vista previa: bajar", []string{"j", "↓"}, onPreview(func(m *Model, a command.Args) { m.previewMove(a.N()) }))
	reg("preview.up", "Vista previa: subir", []string{"k", "↑"}, onPreview(func(m *Model, a command.Args) { m.previewMove(-a.N()) }))
	reg("preview.half_down", "Vista previa: media pantalla abajo", []string{"Ctrl+D"}, onPreview(func(m *Model, a command.Args) { m.previewMove(max(1, m.previewRows()/2)) }))
	reg("preview.half_up", "Vista previa: media pantalla arriba", []string{"Ctrl+U"}, onPreview(func(m *Model, a command.Args) { m.previewMove(-max(1, m.previewRows()/2)) }))
	reg("preview.page_down", "Vista previa: una pantalla abajo", []string{"PageDown"}, onPreview(func(m *Model, a command.Args) { m.previewMove(max(1, m.previewRows()-1)) }))
	reg("preview.page_up", "Vista previa: una pantalla arriba", []string{"PageUp"}, onPreview(func(m *Model, a command.Args) { m.previewMove(-max(1, m.previewRows()-1)) }))
	reg("preview.top", "Vista previa: al principio", []string{"gg"}, onPreview(func(m *Model, a command.Args) {
		m.mdState().scroll = 0
		if a.HasCount { // {n}gg: a la línea n del documento renderizado
			m.mdState().scroll = a.Count - 1
		}
		m.clampPreview()
	}))
	reg("preview.bottom", "Vista previa: al final", []string{"G"}, onPreview(func(m *Model, a command.Args) {
		m.mdState().scroll = m.previewMax()
		if a.HasCount {
			m.mdState().scroll = a.Count - 1
		}
		m.clampPreview()
	}))
}

// previewStatus devuelve el texto de la derecha de la statusline en la vista previa: tamaño y posición.
func (m *Model) previewStatus() string {
	md := m.mdState()
	if md == nil || md.doc == nil {
		return "markdown"
	}
	total := len(md.doc.Lines)
	pct := 0
	if max := m.previewMax(); max > 0 {
		pct = m.previewScroll() * 100 / max
	}
	if m.previewMax() == 0 {
		pct = 100
	}
	return fmt.Sprintf("markdown  utf-8  %s  %d líneas  %d%%", map[bool]string{false: "LF", true: "CRLF"}[m.buf.CRLF()], total, pct)
}
