package ui

import (
	"strings"
	"testing"
)

// startGridText devuelve el texto plano de la rejilla (sin estilos), una fila por línea.
func startGridText(g *grid) string {
	var sb strings.Builder
	for y, row := range g.cells {
		if y > 0 {
			sb.WriteByte('\n')
		}
		for _, c := range row {
			if !c.cont {
				sb.WriteString(c.s)
			}
		}
	}
	return sb.String()
}

// findCell busca la primera celda con el texto dado y devuelve su posición.
func findCell(g *grid, s string) (int, int, bool) {
	for y, row := range g.cells {
		for x, c := range row {
			if c.s == s && !c.cont {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

func newStartGrid(w, h int, th Theme) *grid {
	return newGrid(w, h, style{fg: th.FG, bg: th.BG})
}

// Con un área grande se dibuja el logo, las ocho acciones con su atajo, todo centrado.
func TestStartScreenBigAreaShowsLogoAndActions(t *testing.T) {
	th := Dark()
	g := newStartGrid(100, 30, th)
	if !drawStart(g, 0, 0, 100, 30, th) {
		t.Fatal("en un área grande debe dibujarse la pantalla de inicio")
	}
	out := startGridText(g)
	if !strings.Contains(out, "█") {
		t.Error("falta el logo de bloques")
	}
	if len(StartEntries) != 8 {
		t.Errorf("la lista debe tener 8 acciones, hay %d", len(StartEntries))
	}
	for _, e := range StartEntries {
		if !strings.Contains(out, e.Label) || !strings.Contains(out, e.Keys) {
			t.Errorf("falta la acción %q con su atajo %q", e.Label, e.Keys)
		}
	}
	// centrado: el logo empieza en la fila y la columna que deja el mismo margen a ambos lados
	x, y, ok := findCell(g, "█")
	if !ok {
		t.Fatal("no se encuentra el logo")
	}
	if y != (30-(len(startLogo)+1+len(StartEntries)))/2 {
		t.Errorf("el logo empieza en la fila %d, quiere %d", y, (30-(len(startLogo)+1+len(StartEntries)))/2)
	}
	if x != (100-textWidth(startLogo[0]))/2 {
		t.Errorf("el logo empieza en la columna %d, quiere %d", x, (100-textWidth(startLogo[0]))/2)
	}
	// los atajos terminan todos alineados a la derecha: la última celda con texto de cada fila de acción coincide
	var ends []int
	for _, e := range StartEntries {
		ly := findLabelRow(g, e.Label)
		last := -1
		for x, c := range g.cells[ly] {
			if !c.cont && c.s != " " {
				last = x
			}
		}
		ends = append(ends, last)
	}
	for _, c := range ends {
		if c != ends[0] {
			t.Errorf("los atajos no terminan en la misma columna: %v", ends)
			break
		}
	}
}

// findLabelRow devuelve la fila de la rejilla donde aparece el texto dado.
func findLabelRow(g *grid, text string) int {
	out := startGridText(g)
	for i, r := range strings.Split(out, "\n") {
		if strings.Contains(r, text) {
			return i
		}
	}
	return -1
}

// Con un área baja se oculta el logo y queda solo la lista, centrada.
func TestStartScreenShortAreaShowsOnlyList(t *testing.T) {
	th := Dark()
	h := len(StartEntries) + 2 // menos que logo + hueco + lista, pero cabe la lista
	g := newStartGrid(100, h, th)
	if !drawStart(g, 0, 0, 100, h, th) {
		t.Fatal("si cabe la lista debe dibujarse algo")
	}
	out := startGridText(g)
	if strings.Contains(out, "█") {
		t.Error("con el área baja no debe verse el logo")
	}
	for _, e := range StartEntries {
		if !strings.Contains(out, e.Label) || !strings.Contains(out, e.Keys) {
			t.Errorf("falta la acción %q", e.Label)
		}
	}
	// la lista se centra verticalmente: primera fila (h - 8) / 2
	_, y, ok := findCell(g, "B") // «Buscar archivo»
	if !ok {
		t.Fatal("no se encuentra la lista")
	}
	if want := (h - len(StartEntries)) / 2; y != want {
		t.Errorf("la lista empieza en la fila %d, quiere %d", y, want)
	}
}

// Con un área mínima (ni filas ni columnas para la lista) no se dibuja nada.
func TestStartScreenMinimumDrawsNothing(t *testing.T) {
	th := Dark()
	// sin filas: una menos de las que necesita la lista
	h := len(StartEntries) - 1
	g := newStartGrid(100, max(h, 1), th)
	if drawStart(g, 0, 0, 100, h, th) {
		t.Error("sin sitio para la lista no debe dibujarse nada")
	}
	if strings.Contains(startGridText(g), "█") || strings.Contains(startGridText(g), "Space") {
		t.Error("la rejilla debe quedar intacta")
	}
	// sin columnas: el ancho de la lista menos uno
	w := startListWidth() - 1
	g2 := newStartGrid(100, 30, th)
	if drawStart(g2, 0, 0, w, 30, th) {
		t.Errorf("con %d columnas no cabe la lista (%d)", w, startListWidth())
	}
	if strings.TrimSpace(startGridText(g2)) != "" {
		t.Error("la rejilla debe quedar intacta")
	}
}

// Nada se dibuja fuera del área pedida: el área empieza en x=60 y es estrecha.
func TestStartScreenDoesNotDrawOutsideArea(t *testing.T) {
	th := Dark()
	g := newStartGrid(100, 30, th)
	if !drawStart(g, 60, 0, 40, 30, th) {
		t.Fatal("40 columnas y 30 filas caben la pantalla completa")
	}
	for y, row := range g.cells {
		for x, c := range row {
			if x < 60 && c.s != " " {
				t.Fatalf("celda (%d,%d) fuera del área: %q", x, y, c.s)
			}
		}
	}
}

// Colores del tema (decisión 7): logo en Brand y negrita, etiquetas en Blue, atajos en Peach, todo sobre BG.
func TestStartScreenUsesThemeColors(t *testing.T) {
	th := Dark()
	g := newStartGrid(100, 30, th)
	drawStart(g, 0, 0, 100, 30, th)
	x, y, ok := findCell(g, "█")
	if !ok {
		t.Fatal("no se encuentra el logo")
	}
	st := g.cells[y][x].st
	if st.fg != th.Brand || !st.bold || st.bg != th.BG {
		t.Errorf("logo: %+v", st)
	}
	x, y, ok = findCell(g, "B")
	if !ok {
		t.Fatal("no se encuentra la lista")
	}
	if st := g.cells[y][x].st; st.fg != th.Blue || st.bg != th.BG {
		t.Errorf("etiqueta: %+v", st)
	}
	// la «S» de un atajo «Space ...» (la etiqueta «Salir» empieza también por S: se busca tras un espacio)
	for yy, row := range g.cells {
		for xx, c := range row {
			if c.s == "S" && xx > 0 && g.cells[yy][xx-1].s == " " && strings.HasPrefix(collectFrom(g, xx, yy), "Space") {
				if st := g.cells[yy][xx].st; st.fg != th.Peach || st.bg != th.BG {
					t.Errorf("atajo: %+v", st)
				}
				return
			}
		}
	}
	t.Error("no se encuentra ningún atajo")
}

func collectFrom(g *grid, x, y int) string {
	var sb strings.Builder
	for _, c := range g.cells[y][x:] {
		sb.WriteString(c.s)
	}
	return sb.String()
}

// El nivel completo necesita logo + hueco + lista; el intermedio solo la lista (decisión 4).
func TestStartScreenLevels(t *testing.T) {
	n := len(StartEntries)
	for _, tc := range []struct {
		h    int
		full bool
	}{
		{len(startLogo) + 1 + n, true},
		{len(startLogo) + n, false}, // una fila menos: el logo ya no cabe
	} {
		th := Dark()
		g := newStartGrid(100, tc.h, th)
		drawStart(g, 0, 0, 100, tc.h, th)
		if full := strings.Contains(startGridText(g), "█"); full != tc.full {
			t.Errorf("h=%d: ¿logo? %v, se esperaba %v", tc.h, full, tc.full)
		}
	}
}

// La sombra del logo (los trazos de caja como ╗ ║) va en el color atenuado, no en el de la marca.
func TestStartScreenShadowIsDim(t *testing.T) {
	th := Dark()
	g := newStartGrid(100, 30, th)
	drawStart(g, 0, 0, 100, 30, th)
	x, y, ok := findCell(g, "╗")
	if !ok {
		t.Fatal("no se encuentra la sombra del logo")
	}
	if st := g.cells[y][x].st; st.fg != th.Dim || st.bg != th.BG {
		t.Errorf("sombra: %+v", st)
	}
}
