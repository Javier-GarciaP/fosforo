package editor

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// rangeText devuelve el texto de un rango.
func (m *Model) rangeText(r textRange) string {
	if r.linewise {
		return m.buf.Range(buffer.Pos{Line: r.start.Line}, buffer.Pos{Line: r.end.Line, Col: m.buf.LineLen(r.end.Line)})
	}
	return m.buf.Range(r.start, r.end)
}

// setRegister guarda el rango en el registro sin nombre y devuelve su texto.
func (m *Model) setRegister(r textRange) string {
	text := m.rangeText(r)
	m.reg = register{text: text, linewise: r.linewise}
	return text
}

// deleteLines elimina las líneas s..e dejando siempre al menos una línea (vacía) en el buffer.
func (m *Model) deleteLines(s, e int) error {
	n := m.buf.LineCount()
	var err error
	switch {
	case s == 0 && e == n-1:
		_, err = m.buf.Delete(buffer.Pos{}, buffer.Pos{Line: e, Col: m.buf.LineLen(e)})
	case e < n-1:
		_, err = m.buf.Delete(buffer.Pos{Line: s}, buffer.Pos{Line: e + 1})
	default: // hasta la última línea: se quita el salto de línea anterior
		_, err = m.buf.Delete(buffer.Pos{Line: s - 1, Col: m.buf.LineLen(s - 1)}, buffer.Pos{Line: e, Col: m.buf.LineLen(e)})
	}
	return err
}

// applyOperator ejecuta un operador (y, d, c) sobre un rango.
func (m *Model) applyOperator(op string, r textRange) tea.Cmd {
	if !r.linewise && r.start == r.end {
		if op == "c" {
			m.setMode(ModeInsert)
		}
		return nil // rango vacío: no hay nada que copiar ni borrar
	}
	text := m.setRegister(r)
	switch op {
	case "y":
		if r.linewise {
			m.cur.Set(m.buf, buffer.Pos{Line: r.start.Line, Col: m.cur.Col})
		} else {
			m.cur.Set(m.buf, r.start)
		}
		m.setMode(ModeNormal)
		if r.linewise {
			text += "\n"
		}
		return m.clipboard(text)
	case "d":
		m.deleteRange(r)
		m.setMode(ModeNormal)
	case "c":
		if r.linewise {
			if m.edit(m.deleteContent(r.start.Line, r.end.Line)) {
				m.cur.Set(m.buf, buffer.Pos{Line: r.start.Line})
				m.setMode(ModeInsert)
			}
			return nil
		}
		if m.deleteRange(r) {
			m.setMode(ModeInsert)
		}
	}
	return nil
}

// deleteContent borra el contenido de las líneas s..e dejando una sola línea vacía en su lugar.
func (m *Model) deleteContent(s, e int) error {
	_, err := m.buf.Delete(buffer.Pos{Line: s}, buffer.Pos{Line: e, Col: m.buf.LineLen(e)})
	return err
}

// deleteRange borra un rango y coloca el cursor donde empezaba. Devuelve si se pudo.
func (m *Model) deleteRange(r textRange) bool {
	if r.linewise {
		if !m.edit(m.deleteLines(r.start.Line, r.end.Line)) {
			return false
		}
		l := min(r.start.Line, m.lastLine())
		m.cur.Set(m.buf, buffer.Pos{Line: l, Col: firstNonBlank(m.buf.Line(l))})
		return true
	}
	_, err := m.buf.Delete(r.start, r.end)
	if !m.edit(err) {
		return false
	}
	m.cur.Set(m.buf, r.start)
	return true
}

// operatorWithMotion aplica el operador pendiente a la motion indicada.
func (m *Model) operatorWithMotion(op string, def motionDef, count int, hasCount bool) tea.Cmd {
	mot := def.fn(m, count, hasCount, true)
	if mot.failed {
		return nil
	}
	return m.applyOperator(op, m.rangeOf(m.cur.Pos, mot))
}

// operatorLines aplica el operador a count líneas desde la actual (dd, cc, yy).
func (m *Model) operatorLines(op string, count int) tea.Cmd {
	last := min(m.cur.Line+max(1, count)-1, m.lastLine())
	return m.applyOperator(op, textRange{
		start: buffer.Pos{Line: m.cur.Line}, end: buffer.Pos{Line: last, Col: m.buf.LineLen(last)}, linewise: true,
	})
}

// visualRange devuelve el rango seleccionado en modo Visual.
func (m *Model) visualRange() textRange {
	a, b := m.anchor, m.cur.Pos
	if b.Less(a) {
		a, b = b, a
	}
	if m.mode == ModeVisualLine {
		return textRange{start: buffer.Pos{Line: a.Line}, end: buffer.Pos{Line: b.Line, Col: m.buf.LineLen(b.Line)}, linewise: true}
	}
	if b.Col < m.buf.LineLen(b.Line) {
		b.Col++
	}
	return textRange{start: a, end: b}
}

// paste pega el registro sin nombre count veces, después (p) o antes (P) del cursor.
func (m *Model) paste(after bool, count int) {
	if m.reg.empty() {
		m.msg, m.msgErr = "E353: el registro está vacío", true
		return
	}
	n := max(1, count)
	line := m.cur.Line
	if m.reg.linewise {
		body := strings.TrimSuffix(strings.Repeat(m.reg.text+"\n", n), "\n")
		var at buffer.Pos
		var text string
		target := line
		if after {
			at, text, target = buffer.Pos{Line: line, Col: m.buf.LineLen(line)}, "\n"+body, line+1
		} else {
			at, text = buffer.Pos{Line: line}, body+"\n"
		}
		if _, _, err := m.buf.Insert(at, text); m.edit(err) {
			m.cur.Set(m.buf, buffer.Pos{Line: target, Col: firstNonBlank(m.buf.Line(target))})
		}
		return
	}
	text := strings.Repeat(m.reg.text, n)
	at := m.cur.Pos
	if after && m.buf.LineLen(line) > 0 {
		at.Col++
	}
	_, end, err := m.buf.Insert(at, text)
	if !m.edit(err) {
		return
	}
	if strings.Contains(text, "\n") {
		m.cur.Set(m.buf, at)
	} else {
		m.cur.Set(m.buf, buffer.Pos{Line: end.Line, Col: max(0, end.Col-1)})
	}
}
