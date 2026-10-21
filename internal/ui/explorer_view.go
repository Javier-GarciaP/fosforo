package ui

// drawExplorer dibuja el panel del explorador en las columnas 0..ExplorerWidth-1, con un separador a la derecha.
// La jerarquía se lee solo por la sangría (2 columnas por nivel): sin flechas ni líneas guía.
func drawExplorer(g *grid, ex *ExplorerView, rows int, th Theme) {
	panel := style{fg: th.FG, bg: th.Panel}
	sep := style{fg: th.Border, bg: th.Panel}
	if ex.Focused {
		sep.fg = th.Blue
	}
	for r := 0; r < rows; r++ {
		g.fill(0, r, ExplorerWidth, panel)
		g.put(ExplorerWidth, r, "│", sep)
		idx := ex.Scroll + r
		if idx < 0 || idx >= len(ex.Rows) {
			continue
		}
		drawExplorerRow(g, r, ex.Rows[idx], idx == ex.Selected, ex.Focused, ex.Icons, th)
	}
}

func drawExplorerRow(g *grid, y int, row ExplorerRow, selected, focused bool, icons IconMode, th Theme) {
	bg := th.Panel
	if selected {
		bg = th.Sel
		g.fill(0, y, ExplorerWidth, style{fg: th.FG, bg: bg})
	}
	x := 1 + row.Depth*2
	name := row.Name
	nameStyle := style{fg: th.FG, bg: bg, bold: selected && focused}
	switch {
	case row.Root:
		nameStyle = style{fg: th.FG, bg: bg, bold: true}
	case row.IsDir:
		nameStyle.fg = th.Blue
		if icons == IconsNone {
			name += "/"
			if !row.Expanded {
				nameStyle.fg = th.Dim
			}
		}
	}
	if row.Err {
		nameStyle.fg = th.Red
	}
	if icons == IconsNerd {
		ic := ResolveIcon(row.Name, row.IsDir, row.Expanded || row.Root)
		g.put(x, y, ic.Glyph, style{fg: th.color(ic.Color), bg: bg})
		x += 2 // el glifo ocupa una celda más un espacio
	}
	avail := ExplorerWidth - x
	if avail <= 0 {
		return
	}
	g.put(x, y, truncateRight(name, avail), nameStyle)
}

// ExplorerScroll devuelve la primera fila visible para que la fila seleccionada quede dentro de rows filas.
func ExplorerScroll(scroll, selected, total, rows int) int {
	if selected < scroll {
		scroll = selected
	}
	if selected >= scroll+rows {
		scroll = selected - rows + 1
	}
	return max(0, min(scroll, max(0, total-rows)))
}
