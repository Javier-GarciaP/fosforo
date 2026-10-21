package ui_test

import (
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/markdown"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

func TestComputeLayout(t *testing.T) {
	doc := markdown.Render("x\n", 40)
	split := &ui.MarkdownView{Doc: &doc, Split: true}
	full := &ui.MarkdownView{Doc: &doc}
	for _, tc := range []struct {
		name      string
		w, h      int
		explorer  bool
		md        *ui.MarkdownView
		lines     int
		wantZones map[[2]int]ui.Zone
	}{
		{"solo editor", 100, 30, false, nil, 20, map[[2]int]ui.Zone{{0, 0}: ui.ZoneEditor, {99, 27}: ui.ZoneEditor, {50, 28}: ui.ZoneStatus, {50, 29}: ui.ZoneCommand}},
		{"con explorador", 100, 30, true, nil, 20, map[[2]int]ui.Zone{{0, 0}: ui.ZoneExplorer, {23, 5}: ui.ZoneExplorer, {25, 5}: ui.ZoneEditor, {24, 5}: ui.ZoneNone}},
		{"vista previa completa", 100, 30, true, full, 20, map[[2]int]ui.Zone{{10, 3}: ui.ZoneExplorer, {60, 3}: ui.ZonePreview}},
		{"dividida", 101, 30, false, split, 20, map[[2]int]ui.Zone{{10, 3}: ui.ZoneEditor, {49, 3}: ui.ZoneEditor, {50, 3}: ui.ZoneNone, {51, 3}: ui.ZonePreview}},
		{"fuera de la pantalla", 100, 30, false, nil, 20, map[[2]int]ui.Zone{{100, 0}: ui.ZoneNone, {0, 30}: ui.ZoneNone, {-1, 3}: ui.ZoneNone}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := ui.ComputeLayout(tc.w, tc.h, tc.explorer, tc.md, tc.lines)
			for pos, want := range tc.wantZones {
				if got := l.ZoneAt(pos[0], pos[1]); got != want {
					t.Errorf("ZoneAt%v = %v, quería %v", pos, got, want)
				}
			}
		})
	}
	l := ui.ComputeLayout(100, 30, true, nil, 1234)
	if l.Gutter != 5 || l.TextX != 25+5 { // 4 cifras + 1
		t.Errorf("gutter %d, texto desde %d", l.Gutter, l.TextX)
	}
	if l.Editor.W != 75 || l.Explorer.W != 24 || l.Status.Y != 28 || l.Command.Y != 29 {
		t.Errorf("rectángulos: %+v", l)
	}
	// terminal estrecha: el explorador se oculta solo
	if n := ui.ComputeLayout(79, 30, true, nil, 5); n.Explorer.W != 0 || n.Editor.X != 0 || n.Editor.W != 79 {
		t.Errorf("sin explorador a 79 columnas: %+v", n)
	}
}

// El dibujo y el ratón comparten geometría: cada zona del layout es exactamente lo que se dibuja.
func TestLayoutMatchesWhatIsDrawn(t *testing.T) {
	src := "uno\ndos\ntres\n"
	b := buffer.FromString("", src)
	v := ui.EditorView{Buf: b, Mode: "NORMAL", Explorer: fosforoTree(ui.IconsNone)}
	rows := strings.Split(plain(v, 100, 12), "\n")
	l := ui.ComputeLayout(100, 12, true, nil, b.LineCount())
	// el texto del editor empieza exactamente en TextX
	row := []rune(rows[0])
	if string(row[l.TextX:l.TextX+3]) != "uno" {
		t.Errorf("el texto no empieza en TextX=%d: %q", l.TextX, rows[0])
	}
	// el separador está justo antes del editor
	if row[l.Editor.X-1] != '│' {
		t.Errorf("separador: %q", rows[0])
	}
}

func TestPosAt(t *testing.T) {
	l := ui.ComputeLayout(80, 20, false, nil, 3) // gutter de 4 columnas
	b := buffer.FromString("", "abc\n日本語x\n\tz\n")
	for _, tc := range []struct {
		name      string
		top, left int
		x, y      int
		want      buffer.Pos
	}{
		{"primer carácter", 0, 0, 4, 0, buffer.Pos{Line: 0, Col: 0}},
		{"tercer carácter", 0, 0, 6, 0, buffer.Pos{Line: 0, Col: 2}},
		{"más allá del final de la línea", 0, 0, 40, 0, buffer.Pos{Line: 0, Col: 3}},
		{"sobre los números de línea", 0, 0, 1, 1, buffer.Pos{Line: 1, Col: 0}},
		{"primera celda de un CJK", 0, 0, 4, 1, buffer.Pos{Line: 1, Col: 0}},
		{"segunda celda de un CJK", 0, 0, 5, 1, buffer.Pos{Line: 1, Col: 0}}, // la misma letra
		{"segundo CJK, primera celda", 0, 0, 6, 1, buffer.Pos{Line: 1, Col: 1}},
		{"segundo CJK, segunda celda", 0, 0, 7, 1, buffer.Pos{Line: 1, Col: 1}},
		{"tras los tres CJK", 0, 0, 10, 1, buffer.Pos{Line: 1, Col: 3}},
		{"tabulador: dentro de él", 0, 0, 6, 2, buffer.Pos{Line: 2, Col: 0}},
		{"tras el tabulador", 0, 0, 8, 2, buffer.Pos{Line: 2, Col: 1}},
		{"bajo el último texto", 0, 0, 5, 10, buffer.Pos{Line: 2, Col: 0}},
		{"con scroll vertical", 1, 0, 4, 0, buffer.Pos{Line: 1, Col: 0}},
		{"con scroll horizontal", 0, 1, 4, 0, buffer.Pos{Line: 0, Col: 1}},
		{"scroll horizontal sobre un CJK partido", 0, 1, 4, 1, buffer.Pos{Line: 1, Col: 0}},
		{"y negativa", 0, 0, 5, -3, buffer.Pos{Line: 0, Col: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ui.PosAt(b, l, tc.top, tc.left, tc.x, tc.y); got != tc.want {
				t.Errorf("PosAt = %+v, quería %+v", got, tc.want)
			}
		})
	}
	// coherencia con el cursor: para cada posición, la celda donde se dibuja el cursor se convierte de vuelta en la misma
	for line := 0; line < b.LineCount(); line++ {
		for col := 0; col < b.LineLen(line); col++ {
			v := ui.EditorView{Buf: b, Cursor: buffer.Pos{Line: line, Col: col}, Mode: "NORMAL"}
			x, y, ok := ui.CursorCell(v, 80, 20)
			if !ok {
				t.Fatalf("sin celda para %d:%d", line, col)
			}
			if got := ui.PosAt(b, l, 0, 0, x, y); got != (buffer.Pos{Line: line, Col: col}) {
				t.Errorf("CursorCell(%d,%d) = (%d,%d) → PosAt %+v", line, col, x, y, got)
			}
		}
	}
}

func TestPickerAndDialogGeometry(t *testing.T) {
	p := filesPicker()
	g := ui.PickerGeom(p, 120, 36)
	if !g.Box.Contains(g.ListX, g.ListY) || g.ListRows != ui.PickerRows(ui.PickerFiles, 120, 36) {
		t.Errorf("%+v", g)
	}
	if g.RowAt(g.ListX, g.ListY) != 0 || g.RowAt(g.ListX, g.ListY+2) != 2 || g.RowAt(g.ListX-1, g.ListY) != -1 ||
		g.RowAt(g.PreviewX+1, g.ListY) != -1 || g.RowAt(g.ListX, g.ListY-1) != -1 || g.RowAt(g.ListX, g.ListY+g.ListRows) != -1 {
		t.Errorf("RowAt: %+v", g)
	}
	// la geometría coincide con lo dibujado: la fila 0 de la lista está en ListY y empieza con la marca de selección
	v := baseView()
	v.Picker = p
	rows := strings.Split(plain(v, 120, 36), "\n")
	if r := []rune(rows[g.ListY]); string(r[g.ListX]) != "▸" {
		t.Errorf("la fila seleccionada no está donde dice la geometría: %q", rows[g.ListY])
	}

	d := dialog()
	dg := ui.DialogGeom(d, 100, 30)
	if len(dg.Buttons) != 3 || dg.ButtonAt(dg.Buttons[1].X, dg.Buttons[1].Y) != 1 || dg.ButtonAt(0, 0) != -1 {
		t.Errorf("%+v", dg)
	}
	drawn := strings.Split(plain(ui.EditorView{Buf: buffer.FromString("", "x\n"), Mode: "NORMAL", Dialog: d}, 100, 30), "\n")
	for i, label := range []string{"Guardar (y)", "Descartar (n)", "Cancelar (Esc)"} {
		b := dg.Buttons[i]
		seg := string([]rune(drawn[b.Y])[b.X : b.X+b.W])
		if strings.TrimSpace(seg) != label {
			t.Errorf("botón %d: se dibuja %q donde la geometría dice %q", i, seg, label)
		}
	}
	// estrecho: botones apilados
	narrow := ui.DialogGeom(d, 40, 20)
	if !narrow.Stacked || narrow.Buttons[0].Y == narrow.Buttons[1].Y {
		t.Errorf("apilados: %+v", narrow)
	}
	for i, b := range narrow.Buttons {
		if narrow.ButtonAt(b.X+1, b.Y) != i {
			t.Errorf("ButtonAt apilado %d", i)
		}
	}
}

func TestRectContains(t *testing.T) {
	r := ui.Rect{X: 2, Y: 3, W: 4, H: 2}
	for _, tc := range []struct {
		x, y int
		want bool
	}{{2, 3, true}, {5, 4, true}, {6, 3, false}, {1, 3, false}, {2, 5, false}, {2, 2, false}} {
		if r.Contains(tc.x, tc.y) != tc.want {
			t.Errorf("Contains(%d,%d)", tc.x, tc.y)
		}
	}
	if (ui.Rect{}).Contains(0, 0) {
		t.Error("un rectángulo vacío no contiene nada")
	}
}
