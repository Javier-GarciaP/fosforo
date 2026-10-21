package ui

import "github.com/Javier-GarciaP/fosforo/internal/buffer"

// Rect es un rectángulo de celdas de pantalla.
type Rect struct{ X, Y, W, H int }

// Contains indica si la celda (x, y) está dentro del rectángulo.
func (r Rect) Contains(x, y int) bool {
	return r.W > 0 && r.H > 0 && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Layout es la geometría de los paneles de la pantalla. Es la única fuente de verdad: la vista lo usa para dibujar y el
// ratón para saber qué hay bajo el puntero, así no hay desfases entre lo que se ve y lo que se pulsa.
type Layout struct {
	Explorer Rect // vacío si está oculto
	Editor   Rect // área del editor: números de línea y texto (vacío en vista previa completa)
	Preview  Rect // vista previa de Markdown (vacío si no hay)
	Status   Rect
	Command  Rect
	Gutter   int // ancho de la columna de números de línea
	TextX    int // columna de pantalla donde empieza el texto del editor
}

// ComputeLayout calcula la geometría para una pantalla de w×h con el explorador y la vista previa indicados.
// lines es el número de líneas del buffer (el ancho de los números de línea depende de él).
func ComputeLayout(w, h int, explorer bool, md *MarkdownView, lines int) Layout {
	px, ew, mx, mw := Areas(w, explorer, md)
	rows := TextRows(h)
	l := Layout{
		Status:  Rect{0, rows, w, 1},
		Command: Rect{0, rows + 1, w, 1},
		Gutter:  GutterWidth(lines),
	}
	if px > 0 {
		l.Explorer = Rect{0, 0, ExplorerWidth, rows}
	}
	if ew > 0 {
		l.Editor = Rect{px, 0, ew, rows}
	}
	if mw > 0 {
		l.Preview = Rect{mx, 0, mw, rows}
	}
	l.TextX = px + l.Gutter
	return l
}

// Zone es la parte de la interfaz que hay bajo una celda.
type Zone int

const (
	ZoneNone Zone = iota
	ZoneExplorer
	ZoneEditor
	ZonePreview
	ZoneStatus
	ZoneCommand
)

// ZoneAt devuelve qué hay en la celda (x, y).
func (l Layout) ZoneAt(x, y int) Zone {
	switch {
	case l.Explorer.Contains(x, y):
		return ZoneExplorer
	case l.Editor.Contains(x, y):
		return ZoneEditor
	case l.Preview.Contains(x, y):
		return ZonePreview
	case l.Status.Contains(x, y):
		return ZoneStatus
	case l.Command.Contains(x, y):
		return ZoneCommand
	}
	return ZoneNone
}

// PickerGeometry es la geometría de la ventana de selección: la caja, la lista de filas y la vista previa.
type PickerGeometry struct {
	Box      Rect
	ListX    int // columna donde empieza la lista
	ListY    int // fila de la primera fila de la lista
	ListW    int
	ListRows int
	PreviewX int // columna de la vista previa (solo en PickerFiles)
}

// PickerGeom calcula la geometría de la ventana para una pantalla de w×h.
func PickerGeom(p *PickerView, w, h int) PickerGeometry {
	r := pickerRect(p.Kind, w, h, len(p.Rows))
	g := PickerGeometry{Box: Rect{r.x, r.y, r.w, r.h}, ListX: r.x + 1, ListY: r.y + 3, ListW: r.w - 2, ListRows: r.h - 4}
	if p.Kind == PickerFiles {
		g.ListW = (r.w - 3) * 40 / 100
		g.PreviewX = g.ListX + g.ListW + 1
	}
	return g
}

// RowAt devuelve el índice de la fila de la lista bajo la celda (x, y), o -1.
func (g PickerGeometry) RowAt(x, y int) int {
	if x < g.ListX || x >= g.ListX+g.ListW || y < g.ListY || y >= g.ListY+g.ListRows {
		return -1
	}
	return y - g.ListY
}

// DialogGeometry es la geometría de un diálogo: la caja y el rectángulo de cada botón.
type DialogGeometry struct {
	Box     Rect
	Buttons []Rect
	Stacked bool // los botones van uno debajo de otro
	Inner   int
	Y       int // fila de la primera línea de botones
}

// DialogGeom calcula la geometría del diálogo para una pantalla de w×h.
func DialogGeom(d *Dialog, w, h int) DialogGeometry {
	boxW := min(w-2, 62)
	inner := boxW - 2
	rowW := 2
	for _, b := range d.Buttons {
		rowW += textWidth(b) + 4
	}
	stacked := rowW > inner
	height := 7 + len(d.Lines) // bordes, título, huecos, botones y pista
	if stacked {
		height += len(d.Buttons) - 1
	}
	boxH := min(h-2, height+1)
	x0, y0 := (w-boxW)/2, max(0, (h-2-boxH)/2)

	geo := DialogGeometry{Box: Rect{x0, y0, boxW, boxH}, Stacked: stacked, Inner: inner}
	by := y0 + 1 + 2 + len(d.Lines) + 1
	geo.Y = by
	bx := x0 + 2
	for i, b := range d.Buttons {
		bw := textWidth(" " + b + " ")
		if stacked {
			geo.Buttons = append(geo.Buttons, Rect{bx, by + i, bw, 1})
			continue
		}
		geo.Buttons = append(geo.Buttons, Rect{bx, by, bw, 1})
		bx += bw + 2
	}
	return geo
}

// ButtonAt devuelve el índice del botón bajo la celda (x, y), o -1.
func (g DialogGeometry) ButtonAt(x, y int) int {
	for i, b := range g.Buttons {
		if b.Contains(x, y) {
			return i
		}
	}
	return -1
}

// PosAt convierte una celda de pantalla en una posición del texto: línea y grafema. Tiene en cuenta los números de línea,
// los tabuladores, los caracteres anchos (un clic en cualquiera de las celdas de un carácter lo elige) y el desplazamiento
// (top es la primera línea visible y left la primera columna visual visible). Un clic bajo la última línea
// devuelve la última línea, y uno sobre los números de línea o más allá del final de una línea, su inicio o su final.
func PosAt(b *buffer.Buffer, l Layout, top, left, x, y int) buffer.Pos {
	line := max(0, min(top+y, b.LineCount()-1))
	vc := max(0, x-l.TextX+left)
	return buffer.Pos{Line: line, Col: buffer.ColAt(b.Line(line), vc)}
}
