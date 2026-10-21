package editor

import (
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// DefaultHighlightDelay es la espera tras la última edición antes de volver a tokenizar (debounce).
const DefaultHighlightDelay = 50 * time.Millisecond

// syncHighlightBytes: los textos de hasta este tamaño se resaltan al cargarlos, sin esperar; el resto, en segundo plano.
const syncHighlightBytes = 100 << 10

var nextBufID atomic.Int64

// truncateAll indica a shiftSpans que descarte todo lo posterior al punto de edición.
const truncateAll = 1 << 30

// hlState es el resaltado de un buffer: tramos por línea para la revisión rev del texto.
// Mientras llega el resultado de la tokenización de la revisión actual se siguen usando los tramos anteriores,
// desplazados por cada edición, de modo que escribir nunca espera al resaltado.
type hlState struct {
	lang      string
	lines     [][]syntax.Span
	rev       int  // revisión del texto de la última tokenización instalada (-1: ninguna)
	requested int  // última revisión para la que se programó una tokenización (-1: ninguna)
	running   bool // hay una tokenización en curso: solo se ejecuta una a la vez
}

// hlDebounceMsg llega cuando pasó el tiempo de espera tras la edición que dejó el buffer en la revisión rev.
type hlDebounceMsg struct {
	id  int64
	rev int
}

// hlDoneMsg lleva el resultado de tokenizar la revisión rev del buffer id.
type hlDoneMsg struct {
	id    int64
	rev   int
	lines [][]syntax.Span
}

// WithHighlightDelay cambia la espera antes de retokenizar tras una edición.
func WithHighlightDelay(d time.Duration) Option { return func(m *Model) { m.hlDelay = d } }

// HighlightRev devuelve la revisión del texto a la que corresponden los colores mostrados (-1 si no hay). Si es menor
// que la del buffer (Buffer().Rev()), los colores son de una versión anterior, desplazados por las ediciones.
func (m Model) HighlightRev() int {
	if hl := m.bufs[m.active].hl; hl != nil {
		return hl.rev
	}
	return -1
}

// HighlightJobs devuelve cuántas tokenizaciones se han lanzado en segundo plano (para las pruebas).
func (m Model) HighlightJobs() int { return int(m.hlJobs.Load()) }

// lfText devuelve el texto del buffer con saltos de línea «\n», sea cual sea el terminador del archivo.
func lfText(b *buffer.Buffer) string {
	var sb strings.Builder
	for i := 0; i < b.LineCount(); i++ {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(b.Line(i))
	}
	return sb.String()
}

// initHighlight prepara el resaltado de un buffer: detecta el lenguaje, engancha el desplazamiento de tramos a sus
// ediciones y, si el texto es pequeño, lo resalta de inmediato.
func (bs *bufState) initHighlight() {
	bs.hl = &hlState{rev: -1, requested: -1}
	bs.buf.OnChange(func(e buffer.Edit) { bs.hl.shift(bs.buf, e) })
	bs.detectLanguage()
}

// detectLanguage elige el lenguaje por el nombre del archivo y descarta los colores anteriores.
func (bs *bufState) detectLanguage() {
	h := bs.hl
	h.lang, h.lines, h.rev, h.requested = syntax.Detect(bs.buf.Path()), nil, -1, -1
	if h.lang == "" {
		return
	}
	text := lfText(bs.buf)
	if len(text) > syntax.MaxFileBytes {
		return
	}
	if len(text) <= syncHighlightBytes {
		h.lines, _ = syntax.Tokenize(h.lang, text)
		h.rev, h.requested = bs.buf.Rev(), bs.buf.Rev()
	}
}

// scheduleHighlight programa (con debounce) la tokenización del buffer activo si su texto cambió desde la última.
func (m *Model) scheduleHighlight() tea.Cmd {
	bs := m.bufs[m.active]
	h := bs.hl
	if h == nil || h.lang == "" || h.requested == bs.buf.Rev() {
		return nil
	}
	h.requested = bs.buf.Rev()
	id, rev := bs.id, h.requested
	return tea.Tick(m.hlDelay, func(time.Time) tea.Msg { return hlDebounceMsg{id: id, rev: rev} })
}

func (m *Model) bufByID(id int64) *bufState {
	for _, bs := range m.bufs {
		if bs.id == id {
			return bs
		}
	}
	return nil
}

// highlightDebounce arranca la tokenización en segundo plano si el texto no cambió desde que se programó.
func (m *Model) highlightDebounce(msg hlDebounceMsg) tea.Cmd {
	bs := m.bufByID(msg.id)
	if bs == nil || bs.hl == nil || bs.buf.Rev() != msg.rev { // ya hay ediciones más recientes: otro debounce se ocupará
		return nil
	}
	if bs.hl.running {
		bs.hl.requested = -1 // se reintentará al terminar la tokenización en curso
		return nil
	}
	text, lang, id, rev := lfText(bs.buf), bs.hl.lang, bs.id, msg.rev
	bs.hl.running = true
	m.hlJobs.Add(1)
	return func() tea.Msg {
		lines, _ := syntax.Tokenize(lang, text)
		return hlDoneMsg{id: id, rev: rev, lines: lines}
	}
}

// highlightDone instala el resultado si todavía corresponde al texto; si ya quedó viejo se descarta.
func (m *Model) highlightDone(msg hlDoneMsg) {
	bs := m.bufByID(msg.id)
	if bs == nil || bs.hl == nil {
		return
	}
	bs.hl.running = false
	if bs.buf.Rev() == msg.rev {
		bs.hl.lines, bs.hl.rev = msg.lines, msg.rev
	}
}

// shift ajusta los tramos tras una edición del texto para que los colores sigan en su sitio hasta que llegue la retokenización.
func (h *hlState) shift(b *buffer.Buffer, e buffer.Edit) {
	if h.lines == nil {
		return
	}
	line := e.Start.Line
	removed, added := strings.Count(e.Deleted, "\n"), strings.Count(e.Inserted, "\n")
	if line+removed >= len(h.lines) {
		h.lines = nil
		return
	}
	off := len(b.Range(buffer.Pos{Line: line}, e.Start)) // bytes de la línea que quedan antes de la edición
	if removed == 0 && added == 0 {
		h.lines[line] = shiftSpans(h.lines[line], off, len(e.Deleted), len(e.Inserted)-len(e.Deleted))
		return
	}

	// Edición de varias líneas. El texto que sobrevive es el principio de la primera línea (hasta off) y el final de la
	// última línea afectada (desde donde acaba lo borrado); esos dos trozos conservan sus colores, y las líneas nuevas
	// del medio quedan sin color hasta retokenizar.
	oldFirst, oldLast := h.lines[line], h.lines[line+removed]
	prefix := shiftSpans(oldFirst, off, truncateAll, 0)
	suffixFrom := off + len(e.Deleted) // dónde empieza el texto que sobrevive en la última línea vieja
	if removed > 0 {
		suffixFrom = len(e.Deleted) - strings.LastIndex(e.Deleted, "\n") - 1
	}
	newBase := off + len(e.Inserted) // dónde empieza ese mismo texto en la última línea nueva
	if added > 0 {
		newBase = len(e.Inserted) - strings.LastIndex(e.Inserted, "\n") - 1
	}
	var suffix []syntax.Span
	for _, sp := range oldLast {
		if sp.End <= suffixFrom {
			continue
		}
		sp.Start, sp.End = max(sp.Start, suffixFrom)-suffixFrom+newBase, sp.End-suffixFrom+newBase
		suffix = append(suffix, sp)
	}

	region := make([][]syntax.Span, 0, added+1)
	if added == 0 {
		region = append(region, append(prefix[:len(prefix):len(prefix)], suffix...))
	} else {
		region = append(region, prefix)
		for i := 1; i < added; i++ {
			region = append(region, nil)
		}
		region = append(region, suffix)
	}
	out := make([][]syntax.Span, 0, len(h.lines)+added-removed)
	out = append(out, h.lines[:line]...)
	out = append(out, region...)
	out = append(out, h.lines[line+removed+1:]...)
	h.lines = out
}

// shiftSpans reajusta los tramos de una línea tras reemplazar delLen bytes en off por otros delLen+delta bytes.
// Los tramos anteriores a la edición no cambian, los posteriores se desplazan y los que la cubren se ajustan.
func shiftSpans(spans []syntax.Span, off, delLen, delta int) []syntax.Span {
	out := spans[:0:0]
	for _, sp := range spans {
		switch {
		case sp.End <= off:
		case sp.Start >= off+delLen:
			sp.Start, sp.End = max(0, sp.Start+delta), max(0, sp.End+delta)
		default: // la edición cae dentro del tramo
			sp.Start = min(sp.Start, off)
			sp.End = max(sp.Start, sp.End+delta)
			if delLen == truncateAll { // «hasta el final»: se recorta en la edición
				sp.End = min(sp.End, off)
			}
		}
		if sp.End > sp.Start {
			out = append(out, sp)
		}
	}
	return out
}
