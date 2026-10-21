package buffer

import (
	"strings"

	"github.com/rivo/uniseg"
)

// TabWidth es el ancho de un tabulador: avanza hasta la siguiente parada múltiplo de este valor.
const TabWidth = 4

// Pos es una posición en el texto: línea y grafema (carácter percibido) dentro de la línea, ambos desde 0.
// Col puede valer hasta el número de grafemas de la línea (justo después del último).
type Pos struct{ Line, Col int }

// Less indica si p va antes que q.
func (p Pos) Less(q Pos) bool {
	return p.Line < q.Line || (p.Line == q.Line && p.Col < q.Col)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// Clusters divide s en grafemas (caracteres tal como los percibe el usuario).
func Clusters(s string) []string {
	if s == "" {
		return nil
	}
	if isASCII(s) {
		out := make([]string, len(s))
		for i := range s {
			out[i] = s[i : i+1]
		}
		return out
	}
	out := make([]string, 0, len(s))
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		out = append(out, g.Str())
	}
	return out
}

// Count devuelve el número de grafemas de s.
func Count(s string) int {
	if isASCII(s) {
		return len(s)
	}
	n := 0
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		n++
	}
	return n
}

func isControl(r rune) bool { return r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0) }

// IsControl indica si el grafema es un carácter de control (se dibuja como un punto para no enviar secuencias al terminal).
func IsControl(cluster string) bool {
	for _, r := range cluster {
		return isControl(r)
	}
	return false
}

// CellWidth devuelve cuántas celdas de pantalla ocupa el grafema si empieza en la columna visual vcol.
// Los tabuladores avanzan hasta la siguiente parada y los caracteres de control ocupan una celda.
func CellWidth(cluster string, vcol int) int {
	if cluster == "\t" {
		return TabWidth - vcol%TabWidth
	}
	if IsControl(cluster) {
		return 1
	}
	return uniseg.StringWidth(cluster)
}

// VisualCol devuelve la columna de pantalla donde empieza el grafema col de la línea.
func VisualCol(line string, col int) int {
	v := 0
	for i, c := range Clusters(line) {
		if i >= col {
			break
		}
		v += CellWidth(c, v)
	}
	return v
}

// ColAt devuelve el grafema que cubre la columna visual v. Si v pasa del final de la línea devuelve el número de grafemas.
func ColAt(line string, v int) int {
	acc := 0
	for i, c := range Clusters(line) {
		w := CellWidth(c, acc)
		if v < acc+w {
			return i
		}
		acc += w
	}
	return Count(line)
}

// byteOffset convierte el índice de grafema col en un desplazamiento en bytes dentro de line.
func byteOffset(line string, col int) int {
	if col <= 0 {
		return 0
	}
	if isASCII(line) {
		return min(col, len(line))
	}
	off := 0
	for i, c := range Clusters(line) {
		if i >= col {
			break
		}
		off += len(c)
	}
	return off
}

// endOf devuelve la posición que alcanza el texto text si empieza en start.
func endOf(start Pos, text string) Pos {
	if !strings.Contains(text, "\n") {
		return Pos{start.Line, start.Col + Count(text)}
	}
	parts := strings.Split(text, "\n")
	return Pos{start.Line + len(parts) - 1, Count(parts[len(parts)-1])}
}
