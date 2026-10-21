package ui

import (
	"github.com/Javier-GarciaP/fosforo/internal/markdown"
)

// MarkdownView es la vista previa de un documento Markdown: ocupa toda la zona del editor o su mitad derecha (modo dividido).
type MarkdownView struct {
	Doc    *markdown.Doc // nil mientras se renderiza por primera vez
	Scroll int           // primera línea visible del documento renderizado
	Split  bool
}

// Areas calcula las columnas de cada zona: el panel del explorador (px), el editor (desde px, ew columnas) y la vista
// previa (desde mx, mw columnas). Sin vista previa el editor ocupa todo; a pantalla completa la vista previa ocupa el
// lugar del editor (ew = 0); en modo dividido se reparten el espacio con una columna de separación.
func Areas(w int, explorer bool, md *MarkdownView) (px, ew, mx, mw int) {
	px = PanelCols(w, explorer)
	rest := w - px
	switch {
	case md == nil:
		return px, rest, 0, 0
	case !md.Split:
		return px, 0, px, rest
	}
	ew = rest / 2
	return px, ew, px + ew + 1, rest - ew - 1
}

// EditorWidth es el ancho del área de texto del editor (sin explorador ni vista previa); en vista previa completa vale 0.
func EditorWidth(w int, explorer, mdOn, split bool) int {
	var md *MarkdownView
	if mdOn {
		md = &MarkdownView{Split: split}
	}
	_, ew, _, _ := Areas(w, explorer, md)
	return ew
}

// PreviewWidth es el ancho de la zona de la vista previa (0 si no hay).
func PreviewWidth(w int, explorer, mdOn, split bool) int {
	var md *MarkdownView
	if mdOn {
		md = &MarkdownView{Split: split}
	}
	_, _, _, mw := Areas(w, explorer, md)
	return mw
}

// MarkdownContentWidth es el ancho al que se ajusta el texto en una zona de mw columnas: deja margen y no pasa de 80,
// que es lo que se lee cómodamente (como el ancho máximo de GitHub).
func MarkdownContentWidth(mw int) int { return max(10, min(mw-3, 80)) }

func (t Theme) roleStyle(r markdown.Run) style {
	st := style{fg: t.FG, bg: t.BG, bold: r.Bold, italic: r.Italic, strike: r.Strike, under: r.Underline}
	switch r.Role {
	case markdown.RoleHeading:
		st.bold = true
	case markdown.RoleRule, markdown.RoleTableLine:
		st.fg = t.Border
	case markdown.RoleBullet:
		st.fg = t.Peach
	case markdown.RoleDone:
		st.fg = t.Green
	case markdown.RoleTodo, markdown.RoleQuote:
		st.fg = t.Dim
	case markdown.RoleQuoteBar:
		st.fg = t.Blue
	case markdown.RoleLink:
		st.fg = t.Blue
	case markdown.RoleInlineCode:
		st.fg, st.bg = t.Mauve, t.Code
	case markdown.RoleCode:
		st.bg = t.Code
		if r.Syntax != 0 {
			st.fg = t.syntaxColor(r.Syntax)
		}
	case markdown.RoleDim:
		st.fg, st.italic = t.Dim, true
	case markdown.RoleImage:
		st.fg = t.Teal
	}
	return st
}

// drawMarkdown dibuja la vista previa en la zona de mw columnas desde x0, rows filas de alto.
func drawMarkdown(g *grid, x0, mw, rows int, mv *MarkdownView, th Theme) {
	if mv.Doc == nil {
		g.put(x0+2, 0, "Renderizando…", style{fg: th.Dim, bg: th.BG, italic: true})
		return
	}
	cw := mv.Doc.Width
	for r := 0; r < rows; r++ {
		i := mv.Scroll + r
		if i < 0 || i >= len(mv.Doc.Lines) {
			break
		}
		line := mv.Doc.Lines[i]
		if line.Fill == markdown.FillCode {
			g.fill(x0+2, r, min(cw, mw-2), style{fg: th.FG, bg: th.Code})
		}
		x := x0 + 2
		for _, run := range line.Runs {
			if x >= x0+mw {
				break
			}
			x = g.put(x, r, run.Text, th.roleStyle(run))
		}
	}
}
