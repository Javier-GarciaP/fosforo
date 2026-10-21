package markdown

import (
	"strings"

	"github.com/rivo/uniseg"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// cell es el contenido de una celda de tabla en una sola línea.
type cell struct {
	runs  []Run
	width int
}

func runsWidth(runs []Run) int {
	w := 0
	for _, r := range runs {
		w += uniseg.StringWidth(r.Text)
	}
	return w
}

// truncateRuns recorta una lista de fragmentos a w celdas terminando en «…».
func truncateRuns(runs []Run, w int) []Run {
	if runsWidth(runs) <= w {
		return runs
	}
	var out []Run
	acc := 0
	for _, r := range runs {
		var sb strings.Builder
		for _, g := range graphemes(r.Text) {
			gw := uniseg.StringWidth(g)
			if acc+gw > w-1 {
				if sb.Len() > 0 {
					rr := r
					rr.Text = sb.String()
					out = append(out, rr)
				}
				out = append(out, Run{Text: "…", Role: r.Role})
				return mergeRuns(out)
			}
			sb.WriteString(g)
			acc += gw
		}
		rr := r
		rr.Text = sb.String()
		out = append(out, rr)
	}
	return mergeRuns(out)
}

func (r *renderer) table(t *east.Table, width int) []Line {
	var rows [][]cell
	var header []cell
	for row := t.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []cell
		for c := row.FirstChild(); c != nil; c = c.NextSibling() {
			st := style{}
			if _, ok := row.(*east.TableHeader); ok {
				st.bold = true
			}
			atoms := r.inline(c, st)
			var runs []Run
			for _, a := range atoms {
				switch {
				case a.newline:
					runs = append(runs, a.st.run(" "))
				default:
					runs = append(runs, a.st.run(a.text))
				}
			}
			runs = mergeRuns(runs)
			cells = append(cells, cell{runs: runs, width: runsWidth(runs)})
		}
		if _, ok := row.(*east.TableHeader); ok {
			header = cells
		} else {
			rows = append(rows, cells)
		}
	}
	cols := len(t.Alignments)
	if header != nil && len(header) > cols {
		cols = len(header)
	}
	if cols == 0 {
		return nil
	}
	widths := make([]int, cols)
	measure := func(cells []cell) {
		for i, c := range cells {
			if i < cols {
				widths[i] = max(widths[i], c.width)
			}
		}
	}
	measure(header)
	for _, row := range rows {
		measure(row)
	}
	for i := range widths {
		widths[i] = max(widths[i], 1)
	}
	// ancho total: bordes (cols+1) y un espacio de relleno a cada lado de cada columna
	overhead := cols + 1 + 2*cols
	for sum(widths)+overhead > width {
		// se estrecha la columna más ancha (hasta un mínimo de 3 celdas) y sus celdas se recortan con «…»
		wi, wv := -1, 3
		for i, w := range widths {
			if w > wv {
				wi, wv = i, w
			}
		}
		if wi < 0 {
			break
		}
		widths[wi]--
	}

	border := func(l, m, rr string) Line {
		var sb strings.Builder
		sb.WriteString(l)
		for i, w := range widths {
			sb.WriteString(strings.Repeat("─", w+2))
			if i < cols-1 {
				sb.WriteString(m)
			}
		}
		sb.WriteString(rr)
		return Line{Runs: []Run{{Text: sb.String(), Role: RoleTableLine}}}
	}
	rowLine := func(cells []cell) Line {
		var runs []Run
		runs = append(runs, Run{Text: "│", Role: RoleTableLine})
		for i := 0; i < cols; i++ {
			var c cell
			if i < len(cells) {
				c = cells[i]
			}
			content := truncateRuns(c.runs, widths[i])
			cw := runsWidth(content)
			pad := widths[i] - cw
			left := 0
			switch alignment(t, i) {
			case east.AlignRight:
				left = pad
			case east.AlignCenter:
				left = pad / 2
			}
			runs = append(runs, Run{Text: strings.Repeat(" ", 1+left)})
			runs = append(runs, content...)
			runs = append(runs, Run{Text: strings.Repeat(" ", pad-left+1)})
			runs = append(runs, Run{Text: "│", Role: RoleTableLine})
		}
		return Line{Runs: mergeRuns(runs)}
	}

	out := []Line{border("┌", "┬", "┐")}
	if header != nil {
		out = append(out, rowLine(header), border("├", "┼", "┤"))
	}
	for _, row := range rows {
		out = append(out, rowLine(row))
	}
	out = append(out, border("└", "┴", "┘"))
	return out
}

func alignment(t *east.Table, i int) east.Alignment {
	if i < len(t.Alignments) {
		return t.Alignments[i]
	}
	return east.AlignNone
}

func sum(v []int) int {
	s := 0
	for _, x := range v {
		s += x
	}
	return s
}

var _ ast.Node = (*east.Table)(nil)
