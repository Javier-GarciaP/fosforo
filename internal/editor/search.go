package editor

import (
	"fmt"
	"regexp"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// findMatch busca la siguiente coincidencia de re desde from, hacia delante o hacia atrás, dando la vuelta al archivo.
// Las columnas son índices de grafema. Solo se buscan coincidencias dentro de una línea.
func findMatch(b *buffer.Buffer, re *regexp.Regexp, from buffer.Pos, forward bool) (buffer.Pos, bool) {
	n := b.LineCount()
	cols := func(line int) []int {
		text := b.Line(line)
		var out []int
		for _, loc := range re.FindAllStringIndex(text, -1) {
			out = append(out, buffer.Count(text[:loc[0]]))
		}
		return out
	}
	if forward {
		for i := 0; i <= n; i++ {
			line := (from.Line + i) % n
			for _, c := range cols(line) {
				if i == 0 && c <= from.Col {
					continue
				}
				if i == n && c > from.Col { // vuelta completa: solo la propia línea hasta el cursor
					continue
				}
				return buffer.Pos{Line: line, Col: c}, true
			}
		}
		return from, false
	}
	for i := 0; i <= n; i++ {
		line := ((from.Line-i)%n + n) % n
		cs := cols(line)
		for j := len(cs) - 1; j >= 0; j-- {
			c := cs[j]
			if i == 0 && c >= from.Col {
				continue
			}
			if i == n && c < from.Col {
				continue
			}
			return buffer.Pos{Line: line, Col: c}, true
		}
	}
	return from, false
}

// search ejecuta una búsqueda y mueve el cursor a la coincidencia. Si pattern está vacío repite la última.
func (m *Model) search(pattern string, forward bool) {
	if pattern == "" {
		pattern = m.lastSearch
	}
	if pattern == "" {
		m.msg, m.msgErr = "E35: no hay una expresión regular previa", true
		return
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		m.msg, m.msgErr = fmt.Sprintf("E383: patrón no válido: %s", pattern), true
		return
	}
	m.lastSearch, m.lastForward = pattern, forward
	pos, ok := findMatch(m.buf, re, m.cur.Pos, forward)
	if !ok {
		m.msg, m.msgErr = "E486: patrón no encontrado: "+pattern, true
		return
	}
	m.cur.Set(m.buf, pos)
	m.msg = "/" + pattern
}
