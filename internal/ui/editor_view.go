package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// Dialog es un cuadro modal centrado (p. ej. «Cambios sin guardar»).
type Dialog struct {
	Title    string
	Lines    []string // texto del cuerpo; la última línea se atenúa
	Buttons  []string
	Selected int
	Hint     string
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
	x = GutterWidth(v.Buf.LineCount()) + buffer.VisualCol(v.Buf.Line(v.Cursor.Line), v.Cursor.Col) - v.Left
	y = v.Cursor.Line - v.Top
	ok = v.Dialog == nil && y >= 0 && y < TextRows(h) && x >= GutterWidth(v.Buf.LineCount()) && x < w
	return
}

// RenderEditor dibuja el editor en una pantalla de w×h celdas: texto con números de línea, statusline y línea de comandos.
func RenderEditor(v EditorView, w, h int, th Theme) string {
	base := style{fg: th.FG, bg: th.BG}
	g := newGrid(w, h, base)
	lines := v.Buf.LineCount()
	gw, tw, rows := GutterWidth(lines), TextCols(w, lines), TextRows(h)

	for r := 0; r < rows; r++ {
		i := v.Top + r
		if i >= lines {
			g.put(gw-2, r, "~", style{fg: th.Dim, bg: th.BG})
			continue
		}
		rowStyle, numStyle := base, style{fg: th.Dim, bg: th.BG}
		if i == v.Cursor.Line {
			rowStyle.bg, numStyle = th.Sel, style{fg: th.Yellow, bg: th.Sel, bold: true}
			g.fill(0, r, w, rowStyle)
		}
		num := strconv.Itoa(i + 1)
		g.put(gw-1-len(num), r, num, numStyle)
		drawLine(g, gw, r, v.Buf.Line(i), v.Left, tw, rowStyle)
	}

	drawStatusline(g, rows, v, th)
	drawCommandLine(g, rows+1, v, th)
	if v.Dialog != nil {
		drawDialog(g, v.Dialog, th)
	}
	return g.String()
}

// drawLine dibuja la parte de line visible entre las columnas visuales left y left+width.
func drawLine(g *grid, x0, y int, line string, left, width int, st style) {
	vc := 0
	for _, c := range buffer.Clusters(line) {
		s, cw := displayCluster(c, vc)
		if vc >= left+width {
			return
		}
		cst := st
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
				g.put(x0+i-left, y, " ", st)
			}
		}
		vc += cw
	}
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
	if g.w >= 60 {
		eol := "LF"
		if v.Buf.CRLF() {
			eol = "CRLF"
		}
		info = "utf-8  " + eol + "  " + pos
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
	name := v.PathLabel
	if name == "" {
		name = "[Sin nombre]"
	}
	x = g.put(x+1, y, truncateLeft(name, room), bar)
	for _, m := range marks {
		x = g.put(x+1, y, m.text, m.st)
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
	case v.Message != "" && v.IsError:
		g.put(0, y, v.Message, style{fg: th.Red, bg: th.BG, bold: true})
	case v.Message != "":
		g.put(0, y, v.Message, style{fg: th.FG, bg: th.BG})
	case v.Mode != "":
		g.put(0, y, "-- "+v.Mode+" --", style{fg: th.modeColor(v.Mode), bg: th.BG, bold: true})
	}
}

// drawDialog superpone el cuadro modal, centrado y con borde (ver docs/design/exports/10-dialogo-cambios-sin-guardar.png).
func drawDialog(g *grid, d *Dialog, th Theme) {
	boxW := min(g.w-2, 62)
	inner := boxW - 2
	panel := style{fg: th.FG, bg: th.Panel}
	border := style{fg: th.Border, bg: th.Panel}

	// ¿caben los botones en una fila?
	rowW := 2
	for _, b := range d.Buttons {
		rowW += textWidth(b) + 4
	}
	stacked := rowW > inner
	height := 7 + len(d.Lines) // bordes, título, huecos, botones y pista
	if stacked {
		height += len(d.Buttons) - 1
	}
	boxH := min(g.h-2, height+1)
	x0, y0 := (g.w-boxW)/2, max(0, (g.h-2-boxH)/2)

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
	bx := x0 + 2
	for i, b := range d.Buttons {
		st := style{fg: th.FG, bg: th.Sel}
		if i == d.Selected {
			st = style{fg: th.Panel, bg: th.Blue, bold: true}
		}
		label := " " + b + " "
		if stacked {
			g.put(bx, y+i, label, st)
			continue
		}
		g.put(bx, y, label, st)
		bx += textWidth(label) + 2
	}
	if stacked {
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
