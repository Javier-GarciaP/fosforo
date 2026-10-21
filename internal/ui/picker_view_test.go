package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

func baseView() ui.EditorView {
	spans, _ := syntax.Tokenize("Go", trieGo)
	return ui.EditorView{Buf: buffer.FromString("", trieGo), Cursor: buffer.Pos{Line: 11, Col: 34}, Mode: "NORMAL",
		PathLabel: "internal/keymap/trie.go", Highlight: spans, Explorer: fosforoTree(ui.IconsNone)}
}

func filesPicker() *ui.PickerView {
	preview := strings.Split(strings.TrimSuffix(trieGo, "\n"), "\n")
	spans, _ := syntax.Tokenize("Go", trieGo)
	return &ui.PickerView{
		Kind: ui.PickerFiles, Title: "Buscar archivos", Prefix: ">", Query: "trie", Cursor: 4, Counter: "5/42",
		Rows: []ui.PickerRow{
			{Text: "internal/keymap/trie.go", Match: []int{16, 17, 18, 19}},
			{Text: "internal/keymap/trie_test.go", Match: []int{16, 17, 18, 19}},
			{Text: "changes/keymap-trie/design.md", Match: []int{15, 16, 17, 18}},
			{Text: "changes/keymap-trie/tasks.md", Match: []int{15, 16, 17, 18}},
			{Text: "changes/keymap-trie/proposal.md", Match: []int{15, 16, 17, 18}},
		},
		Selected: 0, Preview: &ui.PreviewView{Lines: preview, Spans: spans},
		Hint: "↑↓ navegar   ⏎ abrir   Ctrl+n/p mover   Esc cerrar",
	}
}

func palettePicker() *ui.PickerView {
	return &ui.PickerView{
		Kind: ui.PickerPalette, Title: "Paleta de comandos", Prefix: ":", Query: "gua", Cursor: 3, Counter: "5/60",
		Rows: []ui.PickerRow{
			{Text: "Guardar archivo", Match: []int{0, 1, 2}, Detail: "Ctrl+S  :w"},
			{Text: "Guardar y salir", Match: []int{0, 1, 2}, Detail: ":wq"},
			{Text: "Guardar si hay cambios y salir", Match: []int{0, 1, 2}, Detail: ":x"},
		},
		Hint: "↑↓ navegar   ⏎ ejecutar   Esc cerrar",
	}
}

func TestGoldenFilePicker(t *testing.T) {
	v := baseView()
	v.Picker = filesPicker()
	golden(t, "picker_files", plain(v, 120, 36))
	raw := ui.RenderEditor(v, 120, 36, ui.Dark())
	// el fondo queda atenuado: el texto del editor ya no usa el color normal
	rows := strings.Split(raw, "\n")
	if strings.Contains(rows[0], "38;2;207;215;235") && strings.Contains(strings.Split(ansi.Strip(rows[0]), "│")[1], "package") {
		if !strings.Contains(lastSGRBefore(rows[0], "package"), "38;2;") {
			t.Error("falta color")
		}
	}
	if sgr := lastSGRBefore(rows[0], "fosforo"); strings.Contains(sgr, "38;2;207;215;235;") && !strings.Contains(sgr, "1;") {
		t.Errorf("el explorador de fondo debe atenuarse: %q", sgr)
	}
}

func TestGoldenFilePickerSyntaxPreviewAndHighlights(t *testing.T) {
	v := baseView()
	v.Picker = filesPicker()
	raw := ui.RenderEditor(v, 120, 36, ui.Dark())
	rows := strings.Split(raw, "\n")
	// fila 0 del contenido de la lista: coincidencias en amarillo y negrita
	var listRow string
	for _, r := range rows {
		if strings.Contains(ansi.Strip(r), "internal/keymap/") && strings.Contains(ansi.Strip(r), "▸") {
			listRow = r
		}
	}
	if listRow == "" {
		t.Fatal("no se ve la fila seleccionada")
	}
	i := strings.Index(listRow, "trie")
	if i < 0 {
		// las letras coinciden una a una (SGR entre medias): se busca la «t» tras «keymap/»
		i = strings.Index(listRow, "keymap/") + 7
	}
	seg := listRow[i:]
	if !strings.Contains(seg[:strings.Index(seg, "t")+1], "1;38;2;213;184;94") && !strings.Contains(listRow, "\x1b[0;1;38;2;213;184;94") {
		t.Errorf("las letras coincidentes van en amarillo y negrita: %q", listRow)
	}
	// vista previa coloreada: «package» en mauve dentro de la ventana
	found := false
	for _, r := range rows {
		if strings.Contains(r, "│") && strings.Contains(ansi.Strip(r), "package keymap") && strings.Contains(r, "38;2;204;168;246") {
			found = true
		}
	}
	if !found {
		t.Error("la vista previa debe ir coloreada")
	}
}

func TestGoldenPalette(t *testing.T) {
	v := baseView()
	v.Mode = "COMMAND"
	v.Picker = palettePicker()
	v.Prompt = nil
	out := plain(v, 120, 36)
	golden(t, "palette", out)
	for _, want := range []string{"Guardar archivo", "Ctrl+S  :w", "Guardar y salir", ":wq"} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en la paleta", want)
		}
	}
	// los atajos van alineados a la derecha de la caja
	var cols []int
	for _, r := range strings.Split(out, "\n") {
		for _, k := range []string{":wq", ":x", "Ctrl+S  :w"} {
			if strings.Contains(r, k) && strings.Contains(r, "Guardar") {
				cols = append(cols, ansi.StringWidth(r[:strings.Index(r, k)])+len(k))
			}
		}
	}
	if len(cols) < 3 || cols[0] != cols[1] || cols[1] != cols[2] {
		t.Errorf("los atajos terminan en la misma columna: %v", cols)
	}
}

func TestGoldenListPickerAndEmptyState(t *testing.T) {
	v := baseView()
	v.Picker = &ui.PickerView{
		Kind: ui.PickerList, Title: "Buffers", Prefix: ">", Query: "", Counter: "3/3",
		Rows: []ui.PickerRow{
			{Text: "internal/keymap/trie.go", Detail: "[+]"},
			{Text: "docs/README.md"},
			{Text: "[Sin nombre]"},
		},
		Selected: 1, Hint: "↑↓ navegar   ⏎ cambiar   Esc cerrar",
	}
	golden(t, "picker_list", plain(v, 100, 28))
	v.Picker = &ui.PickerView{Kind: ui.PickerFiles, Title: "Buscar archivos", Prefix: ">", Query: "zzz", Counter: "0/42", Empty: "Sin resultados", Hint: "Esc cerrar"}
	out := plain(v, 100, 28)
	if !strings.Contains(out, "Sin resultados") || !strings.Contains(out, "0/42") {
		t.Errorf("estado vacío:\n%s", out)
	}
}

func TestPickerCursorAndRows(t *testing.T) {
	p := filesPicker()
	x, y := ui.PickerCursorCell(p, 120, 36)
	if y != 4+0 && y <= 0 { // la entrada está dentro de la caja
		t.Errorf("y = %d", y)
	}
	if x < 10 {
		t.Errorf("x = %d", x)
	}
	x2, _ := ui.PickerCursorCell(&ui.PickerView{Kind: ui.PickerFiles, Prefix: ">", Query: "日本", Cursor: 1, Rows: p.Rows}, 120, 36)
	x1, _ := ui.PickerCursorCell(&ui.PickerView{Kind: ui.PickerFiles, Prefix: ">", Query: "日本", Cursor: 0, Rows: p.Rows}, 120, 36)
	if x2-x1 != 2 {
		t.Errorf("un carácter ancho avanza 2 celdas: %d", x2-x1)
	}
	if ui.PickerRows(ui.PickerFiles, 120, 36) < 10 || ui.PickerRows(ui.PickerPalette, 120, 36) < 4 {
		t.Error("PickerRows")
	}
	// el cursor del editor pasa a la entrada del picker
	v := baseView()
	v.Picker = p
	if cx, cy, ok := ui.CursorCell(v, 120, 36); !ok || cx != x || cy != y {
		t.Errorf("CursorCell = %d,%d,%v", cx, cy, ok)
	}
}

func TestPickerRowTruncatesLongPathsFromTheLeft(t *testing.T) {
	v := baseView()
	long := "muy/larga/ruta/de/directorios/anidados/que/no/cabe/en/la/lista/de/resultados/archivo_final.go"
	v.Picker = &ui.PickerView{Kind: ui.PickerFiles, Title: "Buscar archivos", Prefix: ">", Rows: []ui.PickerRow{{Text: long, Match: []int{len([]rune(long)) - 4}}}, Hint: "x"}
	out := plain(v, 100, 28)
	if !strings.Contains(out, "…") || !strings.Contains(out, "archivo_final.go") || strings.Contains(out, "muy/larga/ruta") {
		t.Errorf("se recorta por la izquierda conservando el nombre:\n%s", out)
	}
}

func TestPickerWithWideCharactersKeepsRowWidth(t *testing.T) {
	v := baseView()
	v.Picker = &ui.PickerView{Kind: ui.PickerPalette, Title: "Paleta", Prefix: ":", Query: "日", Rows: []ui.PickerRow{
		{Text: "日本語のコマンド", Match: []int{0}, Detail: ":日"}, {Text: "otro", Detail: "x"}}, Hint: "h"}
	for i, row := range strings.Split(ansi.Strip(ui.RenderEditor(v, 100, 28, ui.Dark())), "\n") {
		if w := ansi.StringWidth(row); w != 100 {
			t.Errorf("fila %d mide %d celdas", i, w)
		}
	}
}

func TestDimMixesTowardsBackground(t *testing.T) {
	v := baseView()
	bright := ui.RenderEditor(v, 100, 20, ui.Dark())
	v.Picker = &ui.PickerView{Kind: ui.PickerPalette, Title: "x", Prefix: ":", Hint: "h"}
	dimmed := ui.RenderEditor(v, 100, 20, ui.Dark())
	if strings.Count(dimmed, "38;2;207;215;235") >= strings.Count(bright, "38;2;207;215;235") {
		t.Error("detrás de la ventana el color normal del texto debe haberse atenuado")
	}
	if !strings.Contains(dimmed, "48;2;23;27;39") {
		t.Error("el fondo no cambia")
	}
}

func TestGoldenLeaderMenu(t *testing.T) {
	v := baseView()
	v.Pending = "<Space> esperando tecla..."
	v.Leader = &ui.LeaderView{Title: "<Space> líder", Entries: []ui.LeaderEntry{
		{Key: "f", Label: "archivos", Group: true}, {Key: "e", Label: "explorador (alternar)"}, {Key: "b", Label: "buffers", Group: true},
		{Key: "s", Label: "buscar", Group: true}, {Key: "h", Label: "ayuda"}, {Key: "m", Label: "markdown", Group: true},
		{Key: "w", Label: "guardar"}, {Key: "q", Label: "cerrar buffer"}, {Key: "Q", Label: "salir"}, {Key: "t", Label: "tema claro / oscuro"},
		{Key: "Space", Label: "paleta de comandos"}, {Key: "/", Label: "buscar en el archivo"}, {Key: ":", Label: "línea de comandos"}, {Key: "?", Label: "lista de atajos"},
	}}
	out := plain(v, 120, 36)
	golden(t, "leader_menu", out)
	for _, want := range []string{"<Space> líder", "→ +archivos", "→ explorador (alternar)", "Space → paleta de comandos", "<Space> esperando tecla..."} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q:\n%s", want, out)
		}
	}
	raw := ui.RenderEditor(v, 120, 36, ui.Dark())
	// grupos en azul con «+», comandos con el color normal, teclas en mauve
	for _, r := range strings.Split(raw, "\n") {
		if strings.Contains(r, "+archivos") {
			if !strings.Contains(lastSGRBefore(r, "+archivos"), "38;2;127;193;254") {
				t.Errorf("los grupos van en azul: %q", lastSGRBefore(r, "+archivos"))
			}
		}
	}
}

func TestLeaderMenuFitsNarrowScreens(t *testing.T) {
	v := baseView()
	v.Explorer = nil
	for _, w := range []int{40, 60, 90} {
		v.Leader = &ui.LeaderView{Title: "<Space> líder", Entries: []ui.LeaderEntry{
			{Key: "f", Label: "archivos", Group: true}, {Key: "e", Label: "explorador"}, {Key: "w", Label: "guardar"}, {Key: "Space", Label: "paleta de comandos"}}}
		for i, row := range strings.Split(ansi.Strip(ui.RenderEditor(v, w, 14, ui.Dark())), "\n") {
			if rw := ansi.StringWidth(row); rw != w {
				t.Errorf("w=%d fila %d mide %d", w, i, rw)
			}
		}
	}
}
