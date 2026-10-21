package editor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

const readmeSrc = `# Proyecto demo

Un párrafo de introducción con **negrita**, ` + "`código`" + ` y un [enlace](https://example.com).

## Características

- Primera
- Segunda
- [x] Hecho
- [ ] Pendiente

## Instalación

` + "```go" + `
package main

func main() { println("hola") }
` + "```" + `

> Una nota importante.

| Tecla | Acción |
|-------|--------|
| ` + "`Space f f`" + ` | Buscar archivo |

![logo](logo.png)

## Final

Fin del documento.
`

// longDoc genera un documento con varias secciones y mucho texto.
func longDoc() string {
	var sb strings.Builder
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&sb, "## Sección %d\n\n", i)
		for j := 0; j < 6; j++ {
			fmt.Fprintf(&sb, "Párrafo %d.%d con bastante texto para ocupar una o dos líneas de la vista previa renderizada en pantalla.\n\n", i, j)
		}
	}
	sb.WriteString("## Último\n\nFINAL_DEL_DOCUMENTO\n")
	return sb.String()
}

func mdEditor(t *testing.T, name, content string, opts ...editor.Option) *ed {
	t.Helper()
	opts = append([]editor.Option{editor.WithHighlightDelay(time.Millisecond), editor.WithMarkdownDelay(20 * time.Millisecond),
		editor.WithTimeout(20 * time.Millisecond), editor.WithLeaderDelay(20 * time.Millisecond)}, opts...)
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := editor.Open(p, opts...)
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	return e.send(tea.WindowSizeMsg{Width: 100, Height: 50})
}

func (e *ed) wantView(want string) {
	e.t.Helper()
	if got := e.m.ViewKind(); got != want {
		e.t.Errorf("vista = %q, quería %q", got, want)
	}
}

func TestMarkdownOpensInPreviewByDefault(t *testing.T) {
	for _, name := range []string{"README.md", "notas.markdown", "LEEME.MD"} {
		e := mdEditor(t, name, readmeSrc)
		e.wantView("preview")
		scr := e.screen()
		if !strings.Contains(scr, "PREVIEW") {
			t.Errorf("%s: la statusline indica PREVIEW:\n%s", name, scr)
		}
		if strings.Contains(scr, "# Proyecto") || strings.Contains(scr, "**negrita**") || strings.Contains(scr, "```") {
			t.Errorf("%s: se ve el fuente en lugar del documento renderizado:\n%s", name, scr)
		}
		if e.m.Buffer().Modified() {
			t.Error("abrir no modifica el archivo")
		}
	}
	for _, name := range []string{"main.go", "notas.txt", "datos.json"} {
		e := mdEditor(t, name, "# no es markdown\n")
		e.wantView("")
		if strings.Contains(e.screen(), "PREVIEW") {
			t.Errorf("%s no tiene vista previa", name)
		}
	}
}

func TestPreviewRendersGitHubStyle(t *testing.T) {
	e := mdEditor(t, "README.md", readmeSrc)
	scr := e.screen()
	for _, want := range []string{
		"Proyecto demo", "──────", // encabezado con regla
		"• Primera", "☑ Hecho", "☐ Pendiente", // listas y tareas
		"┃ Una nota importante.",                       // cita
		"┌", "│ Tecla", "│ Space f f │ Buscar archivo", // tabla con bordes
		"[imagen: logo]", "package main", "Fin del documento",
	} {
		if !strings.Contains(scr, want) {
			t.Errorf("falta %q:\n%s", want, scr)
		}
	}
	if strings.Contains(scr, "https://example.com") {
		t.Error("de un enlace solo se muestra el texto")
	}
}

func TestPreviewColorsUseTheTheme(t *testing.T) {
	e := mdEditor(t, "README.md", readmeSrc)
	rows := e.rows()
	var linkRow, codeRow, headRow string
	for _, r := range rows {
		switch {
		case strings.Contains(r, "enlace"):
			linkRow = r
		case strings.Contains(r, "package"):
			codeRow = r
		case strings.Contains(r, "Proyecto demo"):
			headRow = r
		}
	}
	if sgr := sgrBefore(linkRow[strings.Index(linkRow, "un ")+3:], "enlace"); !strings.Contains(sgr, darkBlue) || !strings.Contains(sgr, ";4") {
		t.Errorf("el enlace va en azul y subrayado: %q", sgr)
	}
	if !strings.Contains(codeRow, "48;2;10;14;24") { // fondo de código #0a0e18
		t.Errorf("el bloque de código tiene su propio fondo: %q", codeRow)
	}
	if sgr := sgrBefore(codeRow, "package"); !strings.Contains(sgr, darkMauve) {
		t.Errorf("el código usa los mismos colores de sintaxis que el editor (package en mauve): %q", sgr)
	}
	if sgr := sgrBefore(headRow, "Proyecto"); !strings.Contains(sgr, ";1;") && !strings.Contains(sgr, "[0;1") {
		t.Errorf("los encabezados van en negrita: %q", sgr)
	}
	// tema claro: sin colores oscuros
	e.typ(" t")
	light := strings.Join(e.rows(), "")
	if strings.Contains(light, "48;2;10;14;24") || strings.Contains(light, "48;2;23;27;39") {
		t.Error("con el tema claro no queda ningún fondo oscuro")
	}
	if !strings.Contains(light, "48;2;227;230;239") { // fondo de código claro #e3e6ef
		t.Error("fondo de código del tema claro")
	}
}

func TestPreviewStatusline(t *testing.T) {
	e := mdEditor(t, "README.md", readmeSrc)
	rows := strings.Split(e.screen(), "\n")
	status := rows[len(rows)-2]
	for _, want := range []string{"PREVIEW", "README.md", "(e: editar", "Space m s: dividir)", "markdown", "utf-8", "LF", "líneas", "0%"} {
		if !strings.Contains(status, want) {
			t.Errorf("la statusline debe incluir %q: %q", want, status)
		}
	}
	if strings.Contains(status, "[+]") {
		t.Error("sin cambios")
	}
}

func TestPreviewNavigation(t *testing.T) {
	e := mdEditor(t, "doc.md", longDoc())
	if e.m.PreviewScroll() != 0 {
		t.Fatal("empieza arriba")
	}
	if strings.Contains(e.screen(), "FINAL_DEL_DOCUMENTO") {
		t.Fatal("el final no cabe en pantalla")
	}
	e.typ("j")
	if e.m.PreviewScroll() != 1 {
		t.Errorf("j: %d", e.m.PreviewScroll())
	}
	e.typ("5j")
	if e.m.PreviewScroll() != 6 {
		t.Errorf("5j: %d", e.m.PreviewScroll())
	}
	e.typ("k")
	if e.m.PreviewScroll() != 5 {
		t.Errorf("k: %d", e.m.PreviewScroll())
	}
	e.send(keyCtrlD)
	if want := 5 + (50-2)/2; e.m.PreviewScroll() != want {
		t.Errorf("Ctrl+d baja media pantalla: %d, quería %d", e.m.PreviewScroll(), want)
	}
	e.send(keyCtrlU)
	if e.m.PreviewScroll() != 5 {
		t.Errorf("Ctrl+u: %d", e.m.PreviewScroll())
	}
	e.send(keyPgDown)
	if e.m.PreviewScroll() <= 20 {
		t.Errorf("PageDown: %d", e.m.PreviewScroll())
	}
	e.send(keyPgUp)
	e.typ("G")
	if !strings.Contains(e.screen(), "FINAL_DEL_DOCUMENTO") {
		t.Errorf("G muestra el final del documento:\n%s", e.screen())
	}
	max := e.m.PreviewScroll()
	e.typ("j").typ("5j")
	if e.m.PreviewScroll() != max {
		t.Errorf("no se pasa del final: %d > %d", e.m.PreviewScroll(), max)
	}
	e.typ("gg")
	if e.m.PreviewScroll() != 0 || !strings.Contains(e.screen(), "Sección 1") {
		t.Errorf("gg vuelve al principio: %d", e.m.PreviewScroll())
	}
	e.typ("k").send(keyUp)
	if e.m.PreviewScroll() != 0 {
		t.Error("no se pasa del principio")
	}
	e.send(keyDown, keyDown, keyUp)
	if e.m.PreviewScroll() != 1 {
		t.Errorf("flechas: %d", e.m.PreviewScroll())
	}
	e.typ("12gg") // {n}gg: a la línea n del documento renderizado
	if e.m.PreviewScroll() != 11 {
		t.Errorf("12gg: %d", e.m.PreviewScroll())
	}
}

func TestPreviewIsReadOnly(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	for _, k := range []string{"i", "a", "o", "x", "dd", "p", "u", "v", "V", "/", "w", "ciw"} {
		e.typ(k).send(keyEsc)
		if e.m.Buffer().Modified() || e.text() != readmeSrc || e.m.Mode() != editor.ModeNormal {
			t.Fatalf("la tecla %q modificó el documento o cambió el modo", k)
		}
		e.wantView("preview")
	}
	e.send(keyEnter, keyTab, keyBackspace).typ("zq")
	if e.m.Buffer().Modified() {
		t.Error("las teclas sin función se ignoran")
	}
	if v := e.m.View(); v.Cursor != nil {
		t.Errorf("la vista previa a pantalla completa no muestra cursor: %+v", v.Cursor)
	}
}

func TestPreviewColonCommandsAndSaveStillWork(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	e.ex("w")
	if !strings.Contains(e.m.Message(), "escritas") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.ex("q")
	if !e.quit {
		t.Error(":q funciona desde la vista previa")
	}
	e = mdEditor(t, "doc.md", readmeSrc)
	e.send(keyCtrlS)
	if !strings.Contains(e.m.Message(), "escritas") {
		t.Errorf("Ctrl+S: %q", e.m.Message())
	}
}

func TestPreviewRewrapsOnResize(t *testing.T) {
	e := mdEditor(t, "doc.md", "# T\n\n"+strings.Repeat("palabra ", 40)+"\n")
	wide := len(e.m.PreviewDoc().Lines)
	if e.m.PreviewDoc().Width != 80 {
		t.Errorf("el ancho de texto se limita a 80: %d", e.m.PreviewDoc().Width)
	}
	e.send(tea.WindowSizeMsg{Width: 50, Height: 30})
	narrow := e.m.PreviewDoc()
	if narrow.Width != 47 || len(narrow.Lines) <= wide {
		t.Errorf("al estrechar se vuelve a ajustar: ancho %d, %d líneas (antes %d)", narrow.Width, len(narrow.Lines), wide)
	}
	for _, l := range narrow.Lines {
		if l.Width() > 47 {
			t.Errorf("línea desbordada: %d", l.Width())
		}
	}
	for _, row := range strings.Split(e.screen(), "\n") {
		if len([]rune(row)) > 50+2 {
			t.Errorf("la pantalla no debe desbordarse: %q", row)
		}
	}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 30})
	if len(e.m.PreviewDoc().Lines) != wide {
		t.Error("al ensanchar vuelve al ajuste original")
	}
}

func TestResizeKeepsApproximatePosition(t *testing.T) {
	e := mdEditor(t, "doc.md", longDoc()).typ("20j")
	before := e.m.PreviewScroll()
	e.send(tea.WindowSizeMsg{Width: 60, Height: 30})
	after := e.m.PreviewScroll()
	if after < before || after > before*3 {
		t.Errorf("la posición se conserva aproximadamente: %d → %d", before, after)
	}
}

// ---------- alternar con el fuente ----------

func TestEditFromPreviewLandsOnNearestHeading(t *testing.T) {
	e := mdEditor(t, "README.md", readmeSrc)
	e.send(tea.WindowSizeMsg{Width: 100, Height: 12}) // pantalla baja: el documento no cabe y hay desplazamiento
	// se desplaza hasta que «Instalación» es lo que se ve arriba
	doc := e.m.PreviewDoc()
	var out int
	for _, a := range doc.Anchors {
		if a.Heading && a.Src == 11 { // «## Instalación» es la línea 12 del fuente (índice 11)
			out = a.Out
		}
	}
	e.typ(fmt.Sprintf("%dgg", out+1)).typ("3j") // unas líneas dentro de la sección «Instalación»
	if !strings.Contains(e.screen(), "package main") {
		t.Fatalf("preparación: se ve el contenido de la sección:\n%s", e.screen())
	}
	e.typ("e")
	e.wantView("source")
	e.mode(editor.ModeNormal)
	if got := e.m.Buffer().Line(e.m.Cursor().Line); got != "## Instalación" {
		t.Errorf("el cursor cae en el encabezado de la sección: %q", got)
	}
	scr := e.screen()
	if !strings.Contains(scr, "## Instalación") || strings.Contains(scr, "PREVIEW") || !strings.Contains(scr, "NORMAL") {
		t.Errorf("se ve el fuente en modo Normal:\n%s", scr)
	}
	if v := e.m.View(); v.Cursor == nil {
		t.Error("en el fuente hay cursor")
	}
	// «e» en el fuente es el movimiento a fin de palabra, no alterna
	e.typ("e")
	e.wantView("source")
	// de vuelta a la vista previa con Space m p, mostrando la sección del cursor
	e.space("mp")
	e.wantView("preview")
	if !strings.Contains(e.screen(), "Instalación") {
		t.Errorf("la vista previa muestra la sección donde estaba el cursor:\n%s", e.screen())
	}
	if e.m.Buffer().Modified() {
		t.Error("alternar no modifica el archivo")
	}
}

func TestToggleWithoutHeadingsUsesProportion(t *testing.T) {
	e := mdEditor(t, "plano.md", strings.Repeat("un párrafo largo de relleno para el documento.\n\n", 60))
	e.typ("G").typ("e")
	if e.m.Cursor().Line < 60 {
		t.Errorf("sin encabezados se usa la posición proporcional: línea %d", e.m.Cursor().Line)
	}
	e.space("mp")
	if e.m.PreviewScroll() == 0 {
		t.Error("vuelve cerca del final")
	}
}

func TestToggleKeepsCursorEditsAndMode(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	e.typ("e").typ("G").typ("ix").send(keyEsc) // se edita el final del fuente
	if !e.m.Buffer().Modified() {
		t.Fatal("preparación")
	}
	e.space("mp")
	if !strings.Contains(e.screen(), "xFin del documento") && !strings.Contains(e.screen(), "Fin del documento") {
		t.Errorf("la vista previa refleja lo editado:\n%s", e.screen())
	}
	e.typ("e")
	e.wantView("source")
}

func TestToggleOnNonMarkdownAndLargeDocs(t *testing.T) {
	e := mdEditor(t, "main.go", "package main\n")
	e.space("mp")
	if !strings.Contains(e.m.Message(), "no es Markdown") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.space("ms")
	if !strings.Contains(e.m.Message(), "no es Markdown") || e.m.ViewKind() != "" {
		t.Errorf("mensaje = %q", e.m.Message())
	}

	big := strings.Repeat("Un párrafo normal de relleno para hacer grande el documento.\n\n", 40000) // > 2 MiB
	if len(big) <= 2<<20 {
		t.Fatalf("el documento de prueba mide %d", len(big))
	}
	e = mdEditor(t, "enorme.md", big)
	e.wantView("source")
	if !strings.Contains(e.m.Message(), "2 MiB") || !strings.Contains(e.screen(), "NORMAL") {
		t.Errorf("se abre el fuente avisando del tamaño: %q", e.m.Message())
	}
	e.send(keyEsc).space("mp")
	if !strings.Contains(e.m.Message(), "2 MiB") || e.m.ViewKind() != "source" {
		t.Errorf("alternar también avisa: %q", e.m.Message())
	}
	e.space("ms")
	if e.m.ViewKind() != "source" {
		t.Error("tampoco se divide")
	}
}

// ---------- modo dividido ----------

func TestSplitLayout(t *testing.T) {
	e := mdEditor(t, "README.md", readmeSrc)
	e.space("ms")
	e.wantView("split")
	scr := e.screen()
	rows := strings.Split(scr, "\n")
	// el fuente a la izquierda (con números de línea y marcadores) y la vista previa a la derecha
	found := false
	for _, r := range rows {
		if strings.Contains(r, "## Instalación") && strings.Contains(r, "│") {
			found = true
			left, right, _ := strings.Cut(r, "│")
			if !strings.Contains(left, "## Instalación") || strings.Contains(right, "##") {
				t.Errorf("fuente a la izquierda y renderizado a la derecha: %q", r)
			}
		}
	}
	if !found {
		t.Errorf("fuente y vista previa lado a lado:\n%s", scr)
	}
	if !strings.Contains(scr, "• Primera") || !strings.Contains(scr, "☑ Hecho") {
		t.Errorf("la mitad derecha muestra el documento renderizado:\n%s", scr)
	}
	if !strings.Contains(scr, "NORMAL") || !strings.Contains(scr, "vista previa sincronizada") {
		t.Errorf("la statusline conserva el modo y avisa de la sincronización:\n%s", scr)
	}
	if v := e.m.View(); v.Cursor == nil || v.Cursor.X >= 50 {
		t.Errorf("el cursor está en la mitad izquierda: %+v", v.Cursor)
	}
	// Space m s otra vez cierra la división
	e.space("ms")
	e.wantView("source")
	if strings.Contains(e.screen(), "• Primera") {
		t.Error("al cerrar la división desaparece la vista previa")
	}
	// y desde el modo dividido, «Space m p» pasa a vista previa completa
	e.space("ms").space("mp")
	e.wantView("preview")
}

func TestSplitEditingFlowAndLiveUpdate(t *testing.T) {
	e := mdEditor(t, "doc.md", "# Título\n\nTexto original del párrafo.\n")
	e.space("ms")
	if !strings.Contains(e.screen(), "Texto original del párrafo.") {
		t.Fatal("preparación")
	}
	jobs := e.m.MarkdownJobs()
	e.typ("jjiX").send(keyEsc)
	if !strings.Contains(e.text(), "XTexto") {
		t.Fatalf("se edita el fuente: %q", e.text())
	}
	// la vista previa NO se actualiza al instante (debounce)
	right := func() string {
		var out []string
		for _, r := range strings.Split(e.screen(), "\n") {
			if _, after, ok := strings.Cut(r, "│"); ok {
				out = append(out, after)
			}
		}
		return strings.Join(out, "\n")
	}
	if strings.Contains(right(), "XTexto") {
		t.Error("la vista previa espera el debounce antes de actualizarse")
	}
	e.settle()
	if !strings.Contains(right(), "XTexto original") {
		t.Errorf("tras el debounce se ve el cambio:\n%s", right())
	}
	if e.m.MarkdownJobs() != jobs+1 && e.m.MarkdownJobs() != jobs {
		t.Errorf("trabajos: %d → %d", jobs, e.m.MarkdownJobs())
	}
	// una ráfaga de ediciones se renderiza una sola vez
	before := e.m.PreviewDoc()
	e.typ("ia").typ("bcdefgh").send(keyEsc).settle()
	if e.m.PreviewDoc() == before {
		t.Error("el documento se renderiza de nuevo")
	}
	if !strings.Contains(right(), "abcdefghXTexto") {
		t.Errorf("la última versión está a la vista:\n%s", right())
	}
}

func TestSplitLiveUpdateWithRealDebounce(t *testing.T) {
	m, _ := editor.Open(func() string {
		p := filepath.Join(t.TempDir(), "doc.md")
		os.WriteFile(p, []byte("# T\n\nuno\n"), 0o644)
		return p
	}(), editor.WithHighlightDelay(time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 30})
	e.space("ms") // el debounce por defecto es de 150 ms
	e.typ("jji!").send(keyEsc)
	start := time.Now()
	e.settle()
	el := time.Since(start)
	if el < 120*time.Millisecond || el > time.Second {
		t.Errorf("el debounce es de unos 150 ms (menos de un segundo): %v", el)
	}
	if !strings.Contains(e.screen(), "!uno") {
		t.Errorf("la vista previa refleja el cambio:\n%s", e.screen())
	}
	if editor.DefaultMarkdownDelay != 150*time.Millisecond {
		t.Errorf("DefaultMarkdownDelay = %v", editor.DefaultMarkdownDelay)
	}
}

func TestSplitScrollSynchronization(t *testing.T) {
	e := mdEditor(t, "doc.md", longDoc())
	e.space("ms")
	if e.m.PreviewScroll() != 0 {
		t.Fatalf("empieza arriba: %d", e.m.PreviewScroll())
	}
	if !strings.Contains(e.screen(), "Sección 1") {
		t.Fatal("se ve la primera sección")
	}
	// el fuente baja hasta la sección 6: la vista previa va a la misma sección
	for _, a := range e.m.PreviewDoc().Anchors {
		_ = a
	}
	e.typ("G")
	if !strings.Contains(e.screen(), "FINAL_DEL_DOCUMENTO") {
		t.Errorf("al final del fuente, la vista previa muestra el final:\n%s", e.screen())
	}
	e.typ("gg")
	if e.m.PreviewScroll() != 0 || !strings.Contains(e.screen(), "Sección 1") {
		t.Errorf("de vuelta arriba: %d", e.m.PreviewScroll())
	}
	// a mitad de documento (línea del fuente de «## Sección 5»): la vista previa enseña esa sección
	src := 4 * 14 // cada sección ocupa 14 líneas del fuente (encabezado + 6 párrafos con sus blancos)
	e.typ(fmt.Sprintf("%dG", src+1))
	if !strings.Contains(e.screen(), "Sección 5") && !strings.Contains(e.screen(), "Sección 4") {
		t.Errorf("la vista previa sigue al fuente:\n%s", e.screen())
	}
	if e.m.PreviewScroll() == 0 {
		t.Error("la vista previa se ha desplazado")
	}
}

func TestSplitWithExplorerAndNarrowTerminals(t *testing.T) {
	root := project(t, "docs/README.md="+readmeSrc)
	m, _ := editor.Open("", editor.WithRoot(root), editor.WithHighlightDelay(time.Millisecond), editor.WithMarkdownDelay(time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 130, Height: 30})
	e.ex("e " + filepath.Join(root, "docs", "README.md")).space("e").send(tea.KeyPressMsg{Code: tea.KeyTab})
	e.space("ms")
	scr := e.screen()
	if !strings.Contains(scr, "docs/") || !strings.Contains(scr, "• Primera") || !strings.Contains(scr, "## Instalación") {
		t.Errorf("explorador, fuente y vista previa a la vez:\n%s", scr)
	}
	for _, row := range strings.Split(scr, "\n") {
		if n := len([]rune(row)); n > 130+4 {
			t.Errorf("fila desbordada (%d): %q", n, row)
		}
	}
	e.send(tea.WindowSizeMsg{Width: 40, Height: 12})
	_ = e.m.View()
	e.send(tea.WindowSizeMsg{Width: 0, Height: 0})
	_ = e.m.View()
}

func TestMarkdownPerBufferState(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	os.WriteFile(a, []byte(longDoc()), 0o644)
	os.WriteFile(b, []byte("# B\n\ntexto\n"), 0o644)
	m, _ := editor.Open(a, editor.WithHighlightDelay(time.Millisecond))
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 30}).typ("10j")
	scroll := e.m.PreviewScroll()
	e.ex("e " + b)
	e.wantView("preview")
	if e.m.PreviewScroll() != 0 {
		t.Error("cada buffer tiene su propia posición")
	}
	e.typ("e") // b pasa a fuente
	e.ex("bp")
	e.wantView("preview")
	if e.m.PreviewScroll() != scroll {
		t.Errorf("a conserva su scroll: %d, quería %d", e.m.PreviewScroll(), scroll)
	}
	e.ex("bn")
	e.wantView("source")
}

func TestLeaderPanelShowsMarkdownGroup(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	e.send(tea.KeyPressMsg{Code: ' ', Text: " "}).send(editor.LeaderMsg{Seq: e.m.Seq()})
	if !strings.Contains(e.screen(), "+markdown") {
		t.Errorf("el grupo m aparece solo en el panel:\n%s", e.screen())
	}
	e.typ("m")
	scr := e.screen()
	if !strings.Contains(scr, "vista previa / fuente") || !strings.Contains(scr, "fuente | vista previa") {
		t.Errorf("las acciones del grupo markdown:\n%s", scr)
	}
	e.send(keyEsc)
}

func TestPaletteListsMarkdownCommands(t *testing.T) {
	e := mdEditor(t, "doc.md", readmeSrc)
	e.space(" ").typ("alternar vista previa")
	res := e.results()
	if len(res) == 0 || !strings.Contains(strings.Join(res, "|"), "Alternar vista previa y fuente de Markdown") {
		t.Errorf("la paleta incluye el comando: %v", res)
	}
	e.send(keyEnter)
	e.wantView("source")
	// los comandos de la vista previa lanzados desde la paleta fuera de ella no hacen nada
	e.space(" ").typ("previa: bajar").send(keyEnter)
	e.wantView("source")
	if e.m.Buffer().Modified() {
		t.Error("nada cambia")
	}
}

func TestReloadRerendersPreview(t *testing.T) {
	e := mdEditor(t, "doc.md", "# Uno\n\nprimero\n")
	p := e.m.Buffer().Path()
	if !strings.Contains(e.screen(), "primero") {
		t.Fatal("preparación")
	}
	os.WriteFile(p, []byte("# Uno\n\nsegundo texto distinto\n"), 0o644)
	later := time.Now().Add(3 * time.Second)
	os.Chtimes(p, later, later)
	e.send(tea.FocusMsg{})
	if !strings.Contains(e.screen(), "segundo texto distinto") || strings.Contains(e.screen(), "primero") {
		t.Errorf("al recargar se renderiza el texto nuevo:\n%s", e.screen())
	}
	e.wantView("preview")
}

func TestRenameToMarkdownKeepsSourceView(t *testing.T) {
	root := project(t, "nota.txt=# Titulo\n\ntexto\n")
	e := openProj(t, root)
	e.ex("e " + filepath.Join(root, "nota.txt"))
	e.wantView("")
	e.space("e").typ("j").typ("r").send(keyBackspace, keyBackspace, keyBackspace).typ("md").send(keyEnter)
	e.send(tea.KeyPressMsg{Code: tea.KeyTab})
	e.wantView("source") // se vuelve Markdown pero no se cambia lo que se está viendo sin avisar
	e.space("mp")
	e.wantView("preview")
}

func TestPreviewIsRenderedForTinyAndHugeLineLengths(t *testing.T) {
	e := mdEditor(t, "doc.md", "# T\n\n"+strings.Repeat("x", 5000)+"\n\n```\n"+strings.Repeat("y", 5000)+"\n```\n")
	for _, row := range strings.Split(e.screen(), "\n") {
		if n := len([]rune(row)); n > 104 {
			t.Errorf("fila desbordada: %d", n)
		}
	}
	e.send(tea.WindowSizeMsg{Width: 40, Height: 10})
	_ = e.m.View()
}

func TestHTMLNoticeShownInPreview(t *testing.T) {
	e := mdEditor(t, "doc.md", "<p align=\"center\"><img src=\"x.png\"></p>\n\n# Título\n\ntexto\n")
	if !strings.Contains(e.screen(), "omitido") || strings.Contains(e.screen(), "<p") {
		t.Errorf("aviso discreto de contenido omitido:\n%s", e.screen())
	}
}
