package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// Dialog es un cuadro modal centrado (p. ej. «Cambios sin guardar»).
type Dialog struct {
	Title    string
	Lines    []string // texto del cuerpo; la última línea se atenúa
	Buttons  []string
	Selected int
	Hint     string
}

// Selection es el rango seleccionado en modo Visual: de Start a End, ambos incluidos y ya ordenados.
type Selection struct {
	Start, End buffer.Pos
	Linewise   bool
}

// Anchura del panel del explorador y mínimo de columnas para mostrarlo; en terminales más estrechas se oculta solo.
const (
	ExplorerWidth       = 24
	explorerMinTermCols = 80
)

// ExplorerRow es una fila visible del árbol de archivos.
type ExplorerRow struct {
	Name     string
	Depth    int
	IsDir    bool
	Expanded bool
	Root     bool // la fila del proyecto: nombre en negrita
	Err      bool // no se pudo leer el directorio
}

// ExplorerView es el panel lateral del explorador.
type ExplorerView struct {
	Rows     []ExplorerRow
	Selected int
	Scroll   int // primera fila visible
	Focused  bool
	Icons    IconMode
}

// PanelCols son las columnas que ocupa el panel (con su separador): 0 si el explorador está oculto o la terminal es estrecha.
func PanelCols(w int, show bool) int {
	if !show || w < explorerMinTermCols {
		return 0
	}
	return ExplorerWidth + 1
}

// Prompt es la línea de comandos en edición («:wq», «/patrón»).
type Prompt struct {
	Prefix string
	Text   string
	Cursor int // posición del cursor dentro de Text, en grafemas
}

// EditorView es todo lo que hace falta para dibujar el editor; el dibujo no tiene efectos secundarios.
type EditorView struct {
	Buf       *buffer.Buffer
	Cursor    buffer.Pos
	Top, Left int    // primera línea visible y primera columna visual visible
	Mode      string // etiqueta de modo, p. ej. "INSERT"
	PathLabel string // nombre mostrado; si está vacío se muestra [Sin nombre]
	Message   string // mensaje de la línea de comandos
	IsError   bool
	Pending   string          // teclas pendientes de completar (p. ej. «3d»)
	Info      string          // texto de la statusline en lugar de «utf-8  LF» (p. ej. «3 líneas» en Visual)
	Selection *Selection      // selección de Visual, o nil
	Highlight [][]syntax.Span // tramos de color de cada línea del buffer (puede ser más corto o nil: texto plano)
	Explorer  *ExplorerView
	Markdown  *MarkdownView // vista previa de Markdown (a pantalla completa o dividida), o nil
	PathHint  string        // texto tras el nombre en la statusline (p. ej. «(e: editar)»)
	Right     string        // si no está vacío, sustituye al texto de la derecha de la statusline
	Picker    *PickerView   // ventana de selección (archivos, buffers, paleta), o nil
	Leader    *LeaderView   // panel de ayuda de la tecla líder, o nil
	BufIndex  int           // posición del buffer activo (1..BufCount); solo se muestra con más de uno
	BufCount  int
	Prompt    *Prompt // línea de comandos en edición, o nil
	Dialog    *Dialog
}

// GutterWidth es el ancho de la columna de números (números alineados a la derecha más un espacio).
func GutterWidth(lines int) int { return max(3, len(strconv.Itoa(lines))) + 1 }

// TextRows es cuántas filas de texto caben en una pantalla de h filas (se reservan statusline y línea de comandos).
func TextRows(h int) int { return max(1, h-2) }

// TextCols es cuántas columnas de texto caben en una pantalla de w columnas.
func TextCols(w, lines int) int { return max(1, w-GutterWidth(lines)) }

// CursorCell devuelve la celda de pantalla donde debe mostrarse el cursor, y false si no es visible.
func CursorCell(v EditorView, w, h int) (x, y int, ok bool) {
	if v.Prompt != nil && v.Dialog == nil { // la línea de comandos (aunque muestre sugerencias) conserva el cursor
		x = textWidth(v.Prompt.Prefix) + buffer.VisualCol(v.Prompt.Text, v.Prompt.Cursor)
		return x, h - 1, x < w
	}
	if v.Picker != nil && v.Dialog == nil {
		x, y = PickerCursorCell(v.Picker, w, h)
		return x, y, x < w
	}
	px, ew, _, _ := Areas(w, v.Explorer != nil, v.Markdown)
	if ew == 0 { // vista previa a pantalla completa: no hay cursor de edición
		return 0, 0, false
	}
	gw := GutterWidth(v.Buf.LineCount())
	x = px + gw + buffer.VisualCol(v.Buf.Line(v.Cursor.Line), v.Cursor.Col) - v.Left
	y = v.Cursor.Line - v.Top
	focusInEditor := v.Explorer == nil || !v.Explorer.Focused
	ok = v.Dialog == nil && focusInEditor && y >= 0 && y < TextRows(h) && x >= px+gw && x < px+ew
	return
}

// RenderEditor dibuja el editor en una pantalla de w×h celdas: texto con números de línea, statusline y línea de comandos.
func RenderEditor(v EditorView, w, h int, th Theme) string {
	base := style{fg: th.FG, bg: th.BG}
	g := newGrid(w, h, base)
	lines := v.Buf.LineCount()
	px, ew, mx, mw := Areas(w, v.Explorer != nil, v.Markdown)
	gw, tw, rows := GutterWidth(lines), TextCols(max(ew, 1), lines), TextRows(h)
	if v.Explorer != nil && px > 0 {
		drawExplorer(g, v.Explorer, rows, th)
	}
	if v.Markdown != nil {
		if v.Markdown.Split {
			for r := 0; r < rows; r++ {
				g.put(px+ew, r, "│", style{fg: th.Border, bg: th.BG})
			}
		}
		drawMarkdown(g, mx, mw, rows, v.Markdown, th)
	}

	for r := 0; r < rows && ew > 0; r++ {
		i := v.Top + r
		if i >= lines {
			g.put(px+gw-2, r, "~", style{fg: th.Dim, bg: th.BG})
			continue
		}
		rowStyle, numStyle := base, style{fg: th.Dim, bg: th.BG}
		if i == v.Cursor.Line {
			rowStyle.bg, numStyle = th.Sel, style{fg: th.Yellow, bg: th.Sel, bold: true}
			g.fill(px, r, ew, rowStyle)
		}
		num := strconv.Itoa(i + 1)
		g.put(px+gw-1-len(num), r, num, numStyle)
		from, to := v.Selection.cols(i, v.Buf.LineLen(i))
		var spans []syntax.Span
		if i < len(v.Highlight) {
			spans = v.Highlight[i]
		}
		drawLine(g, px+gw, r, v.Buf.Line(i), v.Left, tw, rowStyle, th, spans, from, to)
		if to > v.Buf.LineLen(i) { // la selección incluye el salto de línea: se marca una celda tras el texto
			if x := buffer.VisualCol(v.Buf.Line(i), v.Buf.LineLen(i)) - v.Left; x >= 0 && x < tw {
				g.put(px+gw+x, r, " ", style{fg: th.FG, bg: th.Visual})
			}
		}
	}

	drawStatusline(g, rows, v, th)
	drawCommandLine(g, rows+1, v, th)
	if v.Leader != nil {
		drawLeader(g, v.Leader, rows, th)
	}
	if v.Picker != nil {
		g.dim(0.55)
		drawPicker(g, v.Picker, th)
	}
	if v.Dialog != nil {
		drawDialog(g, v.Dialog, th)
	}
	return g.String()
}

// drawLine dibuja la parte de line visible entre las columnas visuales left y left+width.
// Los grafemas de selFrom a selTo (sin incluir) se pintan con el fondo selBG.
// Los tramos de sintaxis (spans, por bytes de la línea) dan el color del texto; el fondo (línea actual o selección) no se toca.
func drawLine(g *grid, x0, y int, line string, left, width int, st style, th Theme, spans []syntax.Span, selFrom, selTo int) {
	vc, bo, si := 0, 0, 0
	for idx, c := range buffer.Clusters(line) {
		s, cw := displayCluster(c, vc)
		if vc >= left+width {
			return
		}
		cst := st
		if idx >= selFrom && idx < selTo {
			cst.bg = th.Visual
		}
		for si < len(spans) && spans[si].End <= bo {
			si++
		}
		if si < len(spans) && spans[si].Start <= bo {
			cst.fg = th.syntaxColor(spans[si].Kind)
		}
		bo += len(c)
		if buffer.IsControl(c) {
			cst.fg = ""
			cst.bold = false
		}
		switch {
		case cw == 0:
		case vc >= left && vc+cw <= left+width:
			g.put(x0+vc-left, y, s, cst)
		case vc+cw > left: // carácter cortado por un borde: se muestra como espacio
			for i := max(vc, left); i < min(vc+cw, left+width); i++ {
				g.put(x0+i-left, y, " ", cst)
			}
		}
		vc += cw
	}
}

// cols devuelve el rango de grafemas [from,to) de la línea que cubre la selección; to supera la longitud
// de la línea cuando la selección incluye su salto de línea. Sin selección devuelve (0,0).
func (s *Selection) cols(line, lineLen int) (from, to int) {
	if s == nil || line < s.Start.Line || line > s.End.Line {
		return 0, 0
	}
	if s.Linewise {
		return 0, lineLen + 1
	}
	from, to = 0, lineLen+1
	if line == s.Start.Line {
		from = s.Start.Col
	}
	if line == s.End.Line {
		to = s.End.Col + 1
	}
	return from, to
}

func percent(line, total int) int { return line * 100 / max(1, total) }

func drawStatusline(g *grid, y int, v EditorView, th Theme) {
	bar := style{fg: th.FG, bg: th.Panel}
	g.fill(0, y, g.w, bar)

	chip := style{fg: th.Panel, bg: th.modeColor(v.Mode), bold: true}
	x := g.put(0, y, " "+v.Mode+" ", chip)

	// Información de la derecha; en terminales estrechas se reduce a posición y porcentaje.
	pos := fmt.Sprintf("%d:%d  %d%%", v.Cursor.Line+1, v.Cursor.Col+1, percent(v.Cursor.Line+1, v.Buf.LineCount()))
	info := pos
	if v.BufCount > 1 {
		pos = fmt.Sprintf("%d/%d  %s", v.BufIndex, v.BufCount, pos)
		info = pos
	}
	switch {
	case v.Info != "":
		info = v.Info + "  " + pos
	case g.w >= 60:
		eol := "LF"
		if v.Buf.CRLF() {
			eol = "CRLF"
		}
		info = "utf-8  " + eol + "  " + pos
	}
	if v.Right != "" {
		info = v.Right
	}
	g.putRight(g.w-1, y, info, style{fg: th.Dim, bg: th.Panel})

	// Las marcas [+] y [solo lectura] siempre se ven: si el nombre no cabe, se recorta por la izquierda.
	type mark struct {
		text string
		st   style
	}
	var marks []mark
	if v.Buf.Modified() {
		marks = append(marks, mark{"[+]", style{fg: th.Yellow, bg: th.Panel}})
	}
	if ro, _ := v.Buf.ReadOnly(); ro {
		marks = append(marks, mark{"[solo lectura]", style{fg: th.Red, bg: th.Panel}})
	}
	room := g.w - x - 1 - textWidth(info) - 2
	for _, m := range marks {
		room -= textWidth(m.text) + 1
	}
	if v.PathHint != "" { // el aviso se reserva su sitio antes de repartir lo que queda para el nombre
		room -= min(textWidth(v.PathHint), g.w/2) + 1
	}
	name := v.PathLabel
	if name == "" {
		name = "[Sin nombre]"
	}
	x = g.put(x+1, y, truncateLeft(name, room), bar)
	for _, m := range marks {
		x = g.put(x+1, y, m.text, m.st)
	}
	if v.PathHint != "" {
		g.put(x+1, y, truncateRight(v.PathHint, max(0, g.w-x-1-textWidth(info)-2)), style{fg: th.Dim, bg: th.Panel})
	}
}

// truncateRight recorta s por la derecha con «…» para que ocupe como máximo max celdas.
func truncateRight(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if textWidth(s) <= max {
		return s
	}
	w := 1 // 1 celda para «…»
	var sb strings.Builder
	for _, c := range buffer.Clusters(s) {
		cw := textWidth(c)
		if w+cw > max {
			break
		}
		w += cw
		sb.WriteString(c)
	}
	return sb.String() + "…"
}

// truncateLeft recorta s por la izquierda con «…» para que ocupe como máximo max celdas.
func truncateLeft(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if textWidth(s) <= max {
		return s
	}
	cl := buffer.Clusters(s)
	w, i := 1, len(cl) // 1 celda para «…»
	for i > 0 {
		cw := textWidth(cl[i-1])
		if w+cw > max {
			break
		}
		w += cw
		i--
	}
	return "…" + strings.Join(cl[i:], "")
}

func drawCommandLine(g *grid, y int, v EditorView, th Theme) {
	switch {
	case v.Prompt != nil:
		x := g.put(0, y, v.Prompt.Prefix, style{fg: th.Peach, bg: th.BG, bold: true})
		g.put(x, y, v.Prompt.Text, style{fg: th.FG, bg: th.BG})
	case v.Message != "" && v.IsError:
		g.put(0, y, v.Message, style{fg: th.Red, bg: th.BG, bold: true})
	case v.Message != "":
		g.put(0, y, v.Message, style{fg: th.FG, bg: th.BG})
	case v.Pending != "":
		g.put(0, y, v.Pending, style{fg: th.Dim, bg: th.BG})
	default:
		if banner := modeBanner(v.Mode); banner != "" {
			g.put(0, y, banner, style{fg: th.modeColor(v.Mode), bg: th.BG, bold: true})
		}
	}
}

// modeBanner es el texto «-- MODO --» de la línea de comandos; Normal no muestra ninguno.
func modeBanner(mode string) string {
	switch mode {
	case "INSERT", "VISUAL":
		return "-- " + mode + " --"
	case "V-LINE":
		return "-- VISUAL LINE --"
	}
	return ""
}

// drawDialog superpone el cuadro modal, centrado y con borde (ver docs/design/exports/10-dialogo-cambios-sin-guardar.png).
func drawDialog(g *grid, d *Dialog, th Theme) {
	geo := DialogGeom(d, g.w, g.h)
	x0, y0, boxW, boxH, inner := geo.Box.X, geo.Box.Y, geo.Box.W, geo.Box.H, geo.Inner
	panel := style{fg: th.FG, bg: th.Panel}
	border := style{fg: th.Border, bg: th.Panel}

	for y := 0; y < boxH; y++ {
		g.fill(x0, y0+y, boxW, panel)
	}
	g.put(x0, y0, "┌"+repeat("─", inner)+"┐", border)
	g.put(x0, y0+boxH-1, "└"+repeat("─", inner)+"┘", border)
	for y := 1; y < boxH-1; y++ {
		g.put(x0, y0+y, "│", border)
		g.put(x0+boxW-1, y0+y, "│", border)
	}

	y := y0 + 1
	g.fill(x0+1, y, inner, style{bg: th.Sel})
	g.put(x0+2, y, truncateRight(d.Title, inner-2), style{fg: th.Peach, bg: th.Sel, bold: true})
	y += 2
	for i, l := range d.Lines {
		st := panel
		if i > 0 {
			st.fg = th.Dim
		}
		g.put(x0+2, y, truncateRight(l, inner-2), st)
		y++
	}
	y++
	for i, b := range d.Buttons {
		st := style{fg: th.FG, bg: th.Sel}
		if i == d.Selected {
			st = style{fg: th.Panel, bg: th.Blue, bold: true}
		}
		r := geo.Buttons[i]
		g.put(r.X, r.Y, " "+b+" ", st)
	}
	if geo.Stacked {
		y += len(d.Buttons) - 1
	}
	if d.Hint != "" && y+2 < y0+boxH {
		g.put(x0+2, y+2, truncateRight(d.Hint, inner-2), style{fg: th.Dim, bg: th.Panel})
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
