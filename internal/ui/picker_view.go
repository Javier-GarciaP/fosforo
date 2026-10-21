package ui

import (
	"strconv"
	"strings"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// PickerKind elige la disposición de la ventana de selección.
type PickerKind int

const (
	// PickerFiles es la ventana grande centrada con lista a la izquierda y vista previa a la derecha (maqueta 05).
	PickerFiles PickerKind = iota
	// PickerList es una lista sin vista previa, centrada y más estrecha (buffers).
	PickerList
	// PickerPalette es la lista ancha anclada arriba, con los atajos a la derecha (maqueta 06).
	PickerPalette
)

// PickerRow es una fila de la lista. Match son los índices de runa de Text que coincidieron con lo escrito.
type PickerRow struct {
	Text   string
	Match  []int
	Detail string // texto a la derecha (atajos, «[+]»...)
}

// PreviewView es la vista previa de un archivo: sus primeras líneas con colores de sintaxis, o una nota.
type PreviewView struct {
	Lines []string
	Spans [][]syntax.Span
	Note  string // si no está vacía, se muestra en lugar de las líneas («archivo binario»...)
}

// PickerView es la ventana de selección (archivos, buffers, paleta de comandos).
type PickerView struct {
	Kind     PickerKind
	Title    string
	Prefix   string // «>» en archivos, «:» en la paleta
	Query    string
	Cursor   int // posición del cursor en Query, en grafemas
	Rows     []PickerRow
	Selected int // índice de la fila seleccionada dentro de Rows
	Counter  string
	Preview  *PreviewView
	Hint     string
	Empty    string // texto cuando no hay filas («Sin resultados»)
}

// LeaderEntry es una opción del menú de la tecla líder.
type LeaderEntry struct {
	Key   string
	Label string
	Group bool // lleva más teclas debajo: se muestra con «+»
}

// LeaderView es el panel de ayuda de la tecla líder (maqueta 04).
type LeaderView struct {
	Title   string
	Entries []LeaderEntry
}

type rect struct{ x, y, w, h int }

// pickerRect calcula la caja de la ventana para una pantalla de w×h.
func pickerRect(kind PickerKind, w, h, rows int) rect {
	switch kind {
	case PickerPalette:
		bw := min(w-2, max(50, w*58/100))
		bh := min(h-4, rows+4) // título, aire, entrada, filas y pie
		return rect{(w - bw) / 2, max(1, h/12), bw, bh}
	case PickerList:
		bw := min(w-2, max(44, w*50/100))
		bh := min(h-4, rows+4)
		return rect{(w - bw) / 2, max(1, (h-bh)/3), bw, bh}
	}
	bw := min(w-2, max(60, w*83/100))
	bh := min(h-4, max(12, h*72/100))
	return rect{(w - bw) / 2, max(1, (h-2)/9), bw, bh}
}

// PickerRows devuelve cuántas filas de lista caben en una pantalla de w×h para ese tipo de ventana.
func PickerRows(kind PickerKind, w, h int) int {
	return max(1, pickerRect(kind, w, h, 8).h-4)
}

// PickerCursorCell devuelve la celda del cursor de la entrada de texto de la ventana.
func PickerCursorCell(p *PickerView, w, h int) (x, y int) {
	r := pickerRect(p.Kind, w, h, len(p.Rows))
	return r.x + 2 + textWidth(p.Prefix) + 1 + buffer.VisualCol(p.Query, p.Cursor), r.y + 2
}

// dim atenúa todo lo dibujado hasta ahora mezclando el color del texto con el del fondo: una terminal no mezcla transparencias.
func (g *grid) dim(amount float64) {
	cache := map[string]string{}
	for _, row := range g.cells {
		for i := range row {
			c := &row[i]
			key := c.st.fg + "|" + c.st.bg
			m, ok := cache[key]
			if !ok {
				m = mixHex(c.st.fg, c.st.bg, amount)
				cache[key] = m
			}
			c.st.fg = m
		}
	}
}

// mixHex mezcla a hacia b: amount 0 deja a y 1 da b.
func mixHex(a, b string, amount float64) string {
	ar, ag, ab, ok1 := hexRGB(a)
	br, bg, bb, ok2 := hexRGB(b)
	if !ok1 || !ok2 {
		return a
	}
	mix := func(x, y int) int { return int(float64(x)*(1-amount) + float64(y)*amount + 0.5) }
	return "#" + hex2(mix(ar, br)) + hex2(mix(ag, bg)) + hex2(mix(ab, bb))
}

func hex2(v int) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[v>>4&15], digits[v&15]})
}

// drawPicker dibuja la ventana de selección sobre lo que haya detrás.
func drawPicker(g *grid, p *PickerView, th Theme) {
	r := pickerRect(p.Kind, g.w, g.h, len(p.Rows))
	panel := style{fg: th.FG, bg: th.Panel}
	bar := style{fg: th.FG, bg: th.Sel}
	for y := 0; y < r.h; y++ {
		g.fill(r.x, r.y+y, r.w, panel)
	}

	// título y contador
	g.fill(r.x, r.y, r.w, bar)
	g.put(r.x+1, r.y, truncateRight(p.Title, r.w-12), style{fg: th.Blue, bg: th.Sel, bold: true})
	g.putRight(r.x+r.w-1, r.y, p.Counter, style{fg: th.Dim, bg: th.Sel})

	// entrada (la fila r.y+1 queda en blanco para dar aire, como en la maqueta)
	g.put(r.x+2, r.y+2, p.Prefix, style{fg: th.Blue, bg: th.Panel, bold: true})
	g.put(r.x+2+textWidth(p.Prefix)+1, r.y+2, truncateRight(p.Query, r.w-6), panel)

	listTop, listH := r.y+3, r.h-4
	listW := r.w - 2
	var prevX int
	if p.Kind == PickerFiles {
		listW = (r.w - 3) * 40 / 100
		prevX = r.x + 1 + listW + 1
		for y := 0; y < listH; y++ {
			g.put(prevX-1, listTop+y, "│", style{fg: th.Border, bg: th.Panel})
		}
	}
	// lista
	if len(p.Rows) == 0 && p.Empty != "" {
		g.put(r.x+3, listTop, truncateRight(p.Empty, listW-2), style{fg: th.Dim, bg: th.Panel})
	}
	for i := 0; i < len(p.Rows) && i < listH; i++ {
		drawPickerRow(g, r.x+1, listTop+i, listW, p.Rows[i], i == p.Selected, th)
	}
	// vista previa
	if p.Kind == PickerFiles && p.Preview != nil {
		drawPreview(g, prevX, listTop, r.x+r.w-1-prevX, listH, p.Preview, th)
	}

	// pie
	foot := r.y + r.h - 1
	g.fill(r.x, foot, r.w, bar)
	g.put(r.x+1, foot, truncateRight(p.Hint, r.w-2), style{fg: th.Dim, bg: th.Sel})

	// borde fino alrededor para separar la ventana del fondo
	border := style{fg: th.Border, bg: th.Panel}
	for y := 1; y < r.h-1; y++ {
		g.put(r.x, r.y+y, "│", border)
		g.put(r.x+r.w-1, r.y+y, "│", border)
	}
}

func drawPickerRow(g *grid, x, y, w int, row PickerRow, selected bool, th Theme) {
	bg := th.Panel
	if selected {
		bg = th.Sel
		g.fill(x, y, w, style{fg: th.FG, bg: bg})
		g.put(x, y, "▸", style{fg: th.Blue, bg: bg})
	}
	detailW := textWidth(row.Detail)
	if detailW > 0 {
		g.putRight(x+w-1, y, row.Detail, style{fg: th.Dim, bg: bg})
		detailW += 2
	}
	avail := max(1, w-3-detailW)
	base := style{fg: th.FG, bg: bg}
	hit := style{fg: th.Yellow, bg: bg, bold: true}
	cx := x + 2
	matched := map[int]bool{}
	for _, m := range row.Match {
		matched[m] = true
	}
	text := []rune(row.Text)
	if textWidth(row.Text) > avail { // se recorta por la izquierda, que es donde está lo menos importante de una ruta
		cut := len(text)
		w := 1
		for cut > 0 && w+textWidth(string(text[cut-1])) <= avail {
			cut--
			w += textWidth(string(text[cut]))
		}
		g.put(cx, y, "…", base)
		cx++
		text = text[cut:]
		shift := cut
		m2 := map[int]bool{}
		for k := range matched {
			m2[k-shift] = true
		}
		matched = m2
	}
	for i, r := range text {
		st := base
		if matched[i] {
			st = hit
		}
		cx = g.put(cx, y, string(r), st)
	}
}

func drawPreview(g *grid, x, y, w, h int, pv *PreviewView, th Theme) {
	panel := style{fg: th.FG, bg: th.Panel}
	if pv.Note != "" {
		g.put(x+1, y, truncateRight(pv.Note, w-2), style{fg: th.Dim, bg: th.Panel})
		return
	}
	gw := max(3, len(strconv.Itoa(len(pv.Lines)))) + 1
	for i := 0; i < len(pv.Lines) && i < h; i++ {
		num := strconv.Itoa(i + 1)
		g.put(x+gw-1-len(num), y+i, num, style{fg: th.Dim, bg: th.Panel})
		var spans []syntax.Span
		if i < len(pv.Spans) {
			spans = pv.Spans[i]
		}
		drawLineIn(g, x+gw, y+i, pv.Lines[i], x+w, panel, th, spans)
	}
}

// drawLineIn dibuja una línea con colores de sintaxis recortándola en la columna limit (excluida).
func drawLineIn(g *grid, x0, y int, line string, limit int, st style, th Theme, spans []syntax.Span) {
	x, bo, si := x0, 0, 0
	for _, c := range buffer.Clusters(line) {
		s, cw := displayCluster(c, x-x0)
		if x+cw > limit {
			return
		}
		cst := st
		for si < len(spans) && spans[si].End <= bo {
			si++
		}
		if si < len(spans) && spans[si].Start <= bo {
			cst.fg = th.syntaxColor(spans[si].Kind)
		}
		bo += len(c)
		if buffer.IsControl(c) {
			cst.fg = ""
		}
		if cw > 0 {
			g.put(x, y, s, cst)
		}
		x += cw
	}
}

// drawLeader dibuja el panel de ayuda de la tecla líder encima de la statusline, como en la maqueta 04.
func drawLeader(g *grid, lv *LeaderView, bottom int, th Theme) {
	n := len(lv.Entries)
	cols := 3
	if g.w < 80 {
		cols = 2
	}
	if g.w < 50 {
		cols = 1
	}
	rows := max(1, (n+cols-1)/cols)
	rows = min(rows, max(1, bottom-4)) // sin salirse de la pantalla
	top := bottom - rows - 3
	if top < 0 {
		top = 0
	}
	panel := style{fg: th.FG, bg: th.Panel}
	for y := top; y < bottom; y++ {
		g.fill(0, y, g.w, panel)
	}
	g.put(0, top, strings.Repeat("─", g.w), style{fg: th.Border, bg: th.Panel})
	g.put(1, top+1, lv.Title, style{fg: th.Mauve, bg: th.Panel, bold: true})
	g.put(0, top+2, strings.Repeat("─", g.w), style{fg: th.Border, bg: th.Panel})
	colW := g.w / cols
	for i, e := range lv.Entries {
		c, r := i/rows, i%rows
		if c >= cols {
			break
		}
		x, y := 2+c*colW, top+3+r
		keyW := 5 // «Space»
		g.put(x, y, e.Key, style{fg: th.Mauve, bg: th.Panel, bold: true})
		g.put(x+keyW+1, y, "→", style{fg: th.Dim, bg: th.Panel})
		label := e.Label
		st := style{fg: th.FG, bg: th.Panel}
		if e.Group {
			label, st.fg = "+"+label, th.Blue
		}
		g.put(x+keyW+3, y, truncateRight(label, colW-keyW-6), st)
	}
}
