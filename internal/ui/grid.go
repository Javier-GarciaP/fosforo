package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// style es el estilo de una celda: colores en hexadecimal ("#rrggbb", vacío = por defecto) y negrita.
type style struct {
	fg, bg string
	bold   bool
	italic bool
	under  bool
	strike bool
}

// cell es una celda de pantalla. Los caracteres anchos ocupan una celda con texto seguida de celdas de continuación.
type cell struct {
	s    string
	st   style
	cont bool
}

// grid es la pantalla como una cuadrícula de celdas; permite superponer capas (diálogos) y recortar por columnas.
type grid struct {
	w, h  int
	cells [][]cell
}

func newGrid(w, h int, base style) *grid {
	g := &grid{w: w, h: h, cells: make([][]cell, h)}
	for y := range g.cells {
		row := make([]cell, w)
		for x := range row {
			row[x] = cell{s: " ", st: base}
		}
		g.cells[y] = row
	}
	return g
}

// blank deja la celda (x,y) como un espacio, rompiendo el carácter ancho al que pertenezca.
func (g *grid) blank(x, y int, st style) {
	if x < 0 || x >= g.w || y < 0 || y >= g.h {
		return
	}
	row := g.cells[y]
	if row[x].cont { // mitad derecha de un carácter ancho: se borra el carácter entero
		for i := x; i >= 0; i-- {
			wasCont := row[i].cont
			row[i] = cell{s: " ", st: st}
			if !wasCont {
				break
			}
		}
	} else {
		for i := x + 1; i < g.w && row[i].cont; i++ { // el carácter ancho que empezaba aquí pierde su continuación
			row[i] = cell{s: " ", st: st}
		}
	}
	row[x] = cell{s: " ", st: st}
}

// displayCluster devuelve lo que se dibuja para un grafema y su ancho: los controles se muestran como un punto,
// para no enviar nunca secuencias de escape del archivo al terminal.
func displayCluster(c string, vcol int) (string, int) {
	switch {
	case c == "\t":
		w := buffer.CellWidth(c, vcol)
		return strings.Repeat(" ", w), w
	case buffer.IsControl(c):
		return "·", 1
	}
	return c, uniseg.StringWidth(c)
}

// put escribe text desde (x,y) con el estilo st, recortando en el borde derecho. Devuelve la columna siguiente.
func (g *grid) put(x, y int, text string, st style) int {
	text = strings.ToValidUTF8(text, "�")
	for _, c := range buffer.Clusters(text) {
		s, w := displayCluster(c, x)
		if w == 0 {
			continue
		}
		if x < 0 || x+w > g.w || y < 0 || y >= g.h {
			if x+w > g.w {
				break
			}
			x += w
			continue
		}
		for i := 0; i < w; i++ {
			g.blank(x+i, y, st)
		}
		g.cells[y][x] = cell{s: s, st: st}
		for i := 1; i < w; i++ {
			g.cells[y][x+i] = cell{st: st, cont: true}
		}
		x += w
	}
	return x
}

// fill pinta w celdas desde (x,y) con espacios del estilo st.
func (g *grid) fill(x, y, w int, st style) {
	for i := 0; i < w; i++ {
		g.blank(x+i, y, st)
		if x+i >= 0 && x+i < g.w && y >= 0 && y < g.h {
			g.cells[y][x+i] = cell{s: " ", st: st}
		}
	}
}

// putRight escribe text alineado a la derecha, terminando en la columna end (excluida).
func (g *grid) putRight(end, y int, text string, st style) {
	g.put(end-textWidth(text), y, text, st)
}

func textWidth(s string) int {
	w := 0
	for _, c := range buffer.Clusters(strings.ToValidUTF8(s, "�")) {
		_, cw := displayCluster(c, w)
		w += cw
	}
	return w
}

func hexRGB(h string) (r, g, b int, ok bool) {
	if len(h) != 7 || h[0] != '#' {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(h[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(v >> 16), int(v >> 8 & 0xff), int(v & 0xff), true
}

func (s style) sgr() string {
	var sb strings.Builder
	sb.WriteString("\x1b[0")
	if s.bold {
		sb.WriteString(";1")
	}
	if s.italic {
		sb.WriteString(";3")
	}
	if s.under {
		sb.WriteString(";4")
	}
	if s.strike {
		sb.WriteString(";9")
	}
	if r, g, b, ok := hexRGB(s.fg); ok {
		fmt.Fprintf(&sb, ";38;2;%d;%d;%d", r, g, b)
	}
	if r, g, b, ok := hexRGB(s.bg); ok {
		fmt.Fprintf(&sb, ";48;2;%d;%d;%d", r, g, b)
	}
	sb.WriteByte('m')
	return sb.String()
}

// String convierte la cuadrícula en texto con secuencias de color, una fila por línea.
func (g *grid) String() string {
	var sb strings.Builder
	for y, row := range g.cells {
		if y > 0 {
			sb.WriteByte('\n')
		}
		var cur style
		first := true
		for _, c := range row {
			if c.cont {
				continue
			}
			if first || c.st != cur {
				sb.WriteString(c.st.sgr())
				cur, first = c.st, false
			}
			sb.WriteString(c.s)
		}
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}
