package editor

import (
	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// motion es el resultado de un movimiento: adónde lleva y cómo debe interpretarse cuando un operador lo usa.
// Toda motion devuelve un rango, así d, c e y funcionan con cualquiera sin código extra.
type motion struct {
	pos       buffer.Pos
	linewise  bool // el rango abarca líneas completas (j, k, gg, G)
	inclusive bool // incluye el carácter de destino (e, $)
	vertical  int  // movimiento vertical relativo: conserva la columna deseada al mover el cursor
	failed    bool // no pudo moverse (p. ej. h en la columna 0)
}

// motionFn calcula una motion desde el cursor. op indica que se usa con un operador pendiente.
type motionFn func(m *Model, count int, hasCount bool, op bool) motion

type motionDef struct {
	name string
	desc string
	keys []string // teclas por defecto, separadas por espacios en cada secuencia
	fn   motionFn
}

func (m *Model) lastLine() int { return m.buf.LineCount() - 1 }

func clampLine(m *Model, n int) int { return max(0, min(n, m.lastLine())) }

// motionDefs es la lista cerrada de movimientos de este cambio.
var motionDefs = []motionDef{
	{"cursor.left", "Mover un carácter a la izquierda", []string{"h", "left"}, func(m *Model, n int, _, _ bool) motion {
		p := m.cur.Pos
		if p.Col == 0 {
			if prev, ok := m.buf.Prev(p); ok && m.mode == ModeInsert { // en Insert las flechas cruzan de línea
				return motion{pos: prev}
			}
			return motion{pos: p, failed: true}
		}
		return motion{pos: buffer.Pos{Line: p.Line, Col: max(0, p.Col-n)}}
	}},
	{"cursor.right", "Mover un carácter a la derecha", []string{"l", "right"}, func(m *Model, n int, _, op bool) motion {
		p := m.cur.Pos
		limit := m.buf.LineLen(p.Line) // con operador o en Insert se puede llegar al final de la línea
		if !op && m.mode != ModeInsert {
			limit = max(0, limit-1)
		}
		if p.Col >= limit {
			if next, ok := m.buf.Next(p); ok && m.mode == ModeInsert {
				return motion{pos: next}
			}
			return motion{pos: p, failed: true}
		}
		return motion{pos: buffer.Pos{Line: p.Line, Col: min(p.Col+n, limit)}}
	}},
	{"cursor.down", "Bajar una línea", []string{"j", "down"}, func(m *Model, n int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line+n)
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
	{"cursor.up", "Subir una línea", []string{"k", "up"}, func(m *Model, n int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line-n)
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
	{"cursor.word_forward", "Ir al inicio de la palabra siguiente", []string{"w"}, func(m *Model, n int, _, op bool) motion {
		p := m.cur.Pos
		if op && m.op == "c" { // «cw» actúa como «ce» cuando el cursor está sobre una palabra
			if cl := buffer.Clusters(m.buf.Line(p.Line)); p.Col < len(cl) && charClass(cl[p.Col]) != classBlank {
				for i := 0; i < n; i++ {
					if i == 0 {
						p = endOfCurrentWord(m.buf, p) // la primera «palabra» llega hasta el final de la actual
						continue
					}
					p = wordEnd(m.buf, p)
				}
				return motion{pos: p, inclusive: true}
			}
		}
		for i := 0; i < n; i++ {
			p = nextWordStart(m.buf, p)
		}
		return motion{pos: p}
	}},
	{"cursor.word_back", "Ir al inicio de la palabra anterior", []string{"b"}, func(m *Model, n int, _, _ bool) motion {
		p := m.cur.Pos
		for i := 0; i < n; i++ {
			p = prevWordStart(m.buf, p)
		}
		return motion{pos: p}
	}},
	{"cursor.word_end", "Ir al final de la palabra", []string{"e"}, func(m *Model, n int, _, _ bool) motion {
		p := m.cur.Pos
		for i := 0; i < n; i++ {
			p = wordEnd(m.buf, p)
		}
		return motion{pos: p, inclusive: true}
	}},
	{"cursor.line_start", "Ir al inicio de la línea", []string{"0", "home"}, func(m *Model, _ int, _, _ bool) motion {
		return motion{pos: buffer.Pos{Line: m.cur.Line}}
	}},
	{"cursor.first_nonblank", "Ir al primer carácter no blanco de la línea", []string{"^"}, func(m *Model, _ int, _, _ bool) motion {
		return motion{pos: buffer.Pos{Line: m.cur.Line, Col: firstNonBlank(m.buf.Line(m.cur.Line))}}
	}},
	{"cursor.line_end", "Ir al final de la línea", []string{"$", "end"}, func(m *Model, n int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line+n-1)
		if m.mode == ModeInsert { // en Insert se puede llegar justo después del último carácter
			return motion{pos: buffer.Pos{Line: l, Col: m.buf.LineLen(l)}}
		}
		return motion{pos: buffer.Pos{Line: l, Col: max(0, m.buf.LineLen(l)-1)}, inclusive: true}
	}},
	{"cursor.first_line", "Ir a la primera línea (o a la línea {n})", []string{"g g"}, func(m *Model, n int, has, _ bool) motion {
		l := 0
		if has {
			l = clampLine(m, n-1)
		}
		return motion{pos: buffer.Pos{Line: l, Col: firstNonBlank(m.buf.Line(l))}, linewise: true}
	}},
	{"cursor.last_line", "Ir a la última línea (o a la línea {n})", []string{"G"}, func(m *Model, n int, has, _ bool) motion {
		l := m.lastLine()
		if has {
			l = clampLine(m, n-1)
		}
		return motion{pos: buffer.Pos{Line: l, Col: firstNonBlank(m.buf.Line(l))}, linewise: true}
	}},
	{"cursor.half_down", "Bajar media pantalla", []string{"ctrl+d"}, func(m *Model, _ int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line+max(1, ui.TextRows(m.height)/2))
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
	{"cursor.half_up", "Subir media pantalla", []string{"ctrl+u"}, func(m *Model, _ int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line-max(1, ui.TextRows(m.height)/2))
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
	{"cursor.page_down", "Bajar una pantalla", []string{"pgdown"}, func(m *Model, n int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line+n*m.pageSize())
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
	{"cursor.page_up", "Subir una pantalla", []string{"pgup"}, func(m *Model, n int, _, _ bool) motion {
		l := clampLine(m, m.cur.Line-n*m.pageSize())
		return motion{pos: buffer.Pos{Line: l}, linewise: true, vertical: l - m.cur.Line, failed: l == m.cur.Line}
	}},
}

// motionByCommand busca la motion registrada con un nombre de comando.
func motionByCommand(name string) (motionDef, bool) {
	for _, d := range motionDefs {
		if d.name == name {
			return d, true
		}
	}
	return motionDef{}, false
}

// textRange es un rango de texto sobre el que actúan los operadores. En modo carácter end no se incluye;
// en modo línea start.Line..end.Line se incluyen enteras.
type textRange struct {
	start, end buffer.Pos
	linewise   bool
}

// rangeOf convierte una motion desde from en un rango, con las reglas de Vim para motions exclusivas.
func (m *Model) rangeOf(from buffer.Pos, mot motion) textRange {
	a, b := from, mot.pos
	if b.Less(a) {
		a, b = b, a
	}
	if mot.linewise {
		return textRange{start: buffer.Pos{Line: a.Line}, end: buffer.Pos{Line: b.Line, Col: m.buf.LineLen(b.Line)}, linewise: true}
	}
	if mot.inclusive {
		if b.Col < m.buf.LineLen(b.Line) {
			b.Col++
		}
	} else if b.Col == 0 && b.Line > a.Line {
		// Motion exclusiva que termina en la columna 0: acaba al final de la línea anterior (así «dw» no une líneas).
		b = buffer.Pos{Line: b.Line - 1, Col: m.buf.LineLen(b.Line - 1)}
	}
	return textRange{start: a, end: b}
}
