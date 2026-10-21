package editor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/fs"
	"github.com/Javier-GarciaP/fosforo/internal/fuzzy"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

const (
	maxResults      = 200      // resultados que se muestran como máximo
	syncScoreLimit  = 3000     // hasta este número de elementos se puntúa sin salir del bucle de Bubble Tea
	walkBatch       = 500      // archivos por lote al recorrer el proyecto
	previewLines    = 100      // líneas de la vista previa
	previewBytes    = 64 << 10 // bytes que se leen para la vista previa
	recentFilesKept = 50
)

var nextPickerID atomic.Int64

// pickItem es un elemento de la lista: lo que se ve, lo que se compara con lo escrito y sus datos.
type pickItem struct {
	text   string // lo que se muestra
	match  string // contra lo que se compara lo escrito (por defecto, text)
	detail string // texto a la derecha (atajos, «[+]»)
	data   string // ruta, número de buffer o nombre de comando
}

// picker es una ventana de selección difusa: archivos, buffers, paleta de comandos, atajos o archivos recientes.
// Es la misma para todos; solo cambian la fuente de elementos y la acción al aceptar.
type picker struct {
	id       int64
	kind     ui.PickerKind
	title    string
	prefix   string
	hint     string
	query    []string
	cur      int
	items    []pickItem
	texts    []string // textos de comparación, en paralelo a items
	results  []fuzzy.Result
	sel      int
	scroll   int
	loading  bool // todavía llegan archivos del recorrido
	cancel   context.CancelFunc
	gen      int  // generación de la última puntuación pedida
	scoring  bool // hay una puntuación en segundo plano
	dirty    bool // cambió algo mientras se puntuaba: hay que repetirla
	previews map[string]*ui.PreviewView
	files    <-chan []string // lotes del recorrido de archivos
	keep     bool            // sin texto escrito se conserva el orden de la lista (recientes, buffers, comandos)
	accept   func(m *Model, it pickItem)
}

// filesMsg lleva un lote de archivos del recorrido del proyecto (done indica que no habrá más).
type filesMsg struct {
	id    int64
	batch []string
	done  bool
}

// scoredMsg lleva el resultado de puntuar en segundo plano.
type scoredMsg struct {
	id      int64
	gen     int
	results []fuzzy.Result
}

func (p *picker) queryString() string { return strings.Join(p.query, "") }

func (p *picker) filter(query string, texts []string) []fuzzy.Result {
	return filterTexts(query, texts, p.keep)
}

// filterTexts puntúa y ordena; con keep y sin texto escrito deja los elementos en su orden original.
func filterTexts(query string, texts []string, keep bool) []fuzzy.Result {
	if keep && query == "" {
		n := min(len(texts), maxResults)
		out := make([]fuzzy.Result, n)
		for i := range out {
			out[i] = fuzzy.Result{Index: i}
		}
		return out
	}
	return fuzzy.Filter(query, texts, maxResults)
}

func (p *picker) addItems(items ...pickItem) {
	for _, it := range items {
		if it.match == "" {
			it.match = it.text
		}
		p.items = append(p.items, it)
		p.texts = append(p.texts, it.match)
	}
}

// Picker devuelve si hay una ventana de selección abierta.
func (m Model) PickerOpen() bool { return m.picker != nil }

// PickerTitle devuelve el título de la ventana abierta ("" si no hay).
func (m Model) PickerTitle() string {
	if m.picker == nil {
		return ""
	}
	return m.picker.title
}

// PickerResults devuelve los textos de los resultados actuales, de mejor a peor.
func (m Model) PickerResults() []string {
	if m.picker == nil {
		return nil
	}
	out := make([]string, len(m.picker.results))
	for i, r := range m.picker.results {
		out[i] = m.picker.items[r.Index].text
	}
	return out
}

// PickerLoading indica si el recorrido de archivos sigue en marcha.
func (m Model) PickerLoading() bool { return m.picker != nil && m.picker.loading }

func (m *Model) openPicker(p *picker) tea.Cmd {
	m.closePicker()
	m.resetPending()
	p.id = nextPickerID.Add(1)
	p.previews = map[string]*ui.PreviewView{}
	m.picker = p
	return m.refilter(p)
}

func (m *Model) closePicker() {
	if m.picker != nil && m.picker.cancel != nil {
		m.picker.cancel()
	}
	m.picker = nil
}

// refilter recalcula los resultados. Con pocos elementos se hace al momento; con muchos, en segundo plano.
func (m *Model) refilter(p *picker) tea.Cmd {
	if len(p.items) <= syncScoreLimit {
		p.results = p.filter(p.queryString(), p.texts)
		p.sel, p.scroll = 0, 0
		return nil
	}
	if p.scoring {
		p.dirty = true
		return nil
	}
	return m.scoreAsync(p)
}

func (m *Model) scoreAsync(p *picker) tea.Cmd {
	p.scoring, p.dirty = true, false
	p.gen++
	id, gen, query, keep := p.id, p.gen, p.queryString(), p.keep
	texts := p.texts[:len(p.texts):len(p.texts)]
	return func() tea.Msg {
		return scoredMsg{id: id, gen: gen, results: filterTexts(query, texts, keep)}
	}
}

func (m *Model) pickerScored(msg scoredMsg) tea.Cmd {
	p := m.picker
	if p == nil || p.id != msg.id {
		return nil
	}
	p.scoring = false
	if msg.gen == p.gen && !p.dirty {
		p.results = msg.results
		p.sel, p.scroll = 0, 0
		return nil
	}
	return m.scoreAsync(p) // mientras se puntuaba cambió la búsqueda o llegaron archivos: se repite con lo último
}

// ---- fuentes ----

func (m *Model) rootDir() string {
	if m.root != "" {
		return m.root
	}
	return absPath(".")
}

// openFilePicker abre «Buscar archivos»: el recorrido del proyecto va llenando la lista en segundo plano.
func (m *Model) openFilePicker() tea.Cmd {
	root := m.rootDir()
	ctx, cancel := context.WithCancel(context.Background())
	p := &picker{
		kind: ui.PickerFiles, title: "Buscar archivos", prefix: ">", loading: true, cancel: cancel,
		hint: "↑↓ navegar   ⏎ abrir   Ctrl+n/p mover   Esc cerrar",
		accept: func(m *Model, it pickItem) {
			if m.openPath(it.data, m.displayName(it.data)) {
				m.focus = FocusEditor
			}
		},
	}
	ch := make(chan []string, 8)
	p.files = ch
	cmd := m.openPicker(p)
	go fs.Walk(ctx, root, nil, walkBatch, ch)
	return tea.Batch(cmd, waitFiles(p.id, ch))
}

// waitFiles espera el siguiente lote del recorrido.
func waitFiles(id int64, ch <-chan []string) tea.Cmd {
	return func() tea.Msg {
		batch, ok := <-ch
		if !ok {
			return filesMsg{id: id, done: true}
		}
		return filesMsg{id: id, batch: batch}
	}
}

// filesArrived añade un lote de archivos a la lista, sin interrumpir lo que el usuario esté escribiendo.
func (m *Model) filesArrived(msg filesMsg) tea.Cmd {
	p := m.picker
	if p == nil || p.id != msg.id {
		return nil
	}
	root := m.rootDir()
	for _, rel := range msg.batch {
		p.addItems(pickItem{text: rel, data: filepath.Join(root, filepath.FromSlash(rel))})
	}
	var cmds []tea.Cmd
	if len(msg.batch) > 0 {
		sel := p.sel
		cmds = append(cmds, m.refilter(p))
		if len(p.items) <= syncScoreLimit { // se conserva la fila seleccionada mientras llegan más
			p.sel = min(sel, max(0, len(p.results)-1))
		}
	}
	if msg.done {
		p.loading = false
	} else {
		cmds = append(cmds, waitFiles(p.id, p.files))
	}
	return tea.Batch(cmds...)
}

// bufferItems lista los buffers abiertos, marcando los modificados.
func (m *Model) openBufferPicker() tea.Cmd {
	m.stash()
	p := &picker{
		kind: ui.PickerList, title: "Buffers", prefix: ">", keep: true, hint: "↑↓ navegar   ⏎ cambiar   Esc cerrar",
		accept: func(m *Model, it pickItem) {
			var i int
			fmt.Sscanf(it.data, "%d", &i)
			m.activate(i)
		},
	}
	for i, bs := range m.bufs {
		name := bs.label
		if name == "" {
			name = "[Sin nombre]"
		}
		detail := ""
		if bs.buf.Modified() {
			detail = "[+]"
		}
		p.addItems(pickItem{text: name, detail: detail, data: fmt.Sprint(i)})
	}
	return m.openPicker(p)
}

var (
	paletteOnce  sync.Once
	paletteCache []pickItem
)

// palette devuelve los elementos de la paleta, construidos una vez (el registro no cambia mientras el programa corre).
// Es perezoso para no crear un ciclo de inicialización con el registro de comandos.
func palette() []pickItem {
	paletteOnce.Do(func() { paletteCache = paletteItems() })
	return paletteCache
}

// paletteItems construye la lista de comandos: descripción, y los atajos a la derecha.
func paletteItems() []pickItem {
	cmds := registry.List()
	items := make([]pickItem, 0, len(cmds))
	for _, c := range cmds {
		keys := strings.Join(c.Keys, "  ")
		items = append(items, pickItem{text: c.Desc, match: c.Desc + " " + keys, detail: keys, data: c.Name})
	}
	return items
}

func (m *Model) openPalette(title string) tea.Cmd {
	p := &picker{
		kind: ui.PickerPalette, title: title, prefix: ":", keep: true, hint: "↑↓ navegar   ⏎ ejecutar   Esc cerrar",
		accept: func(m *Model, it pickItem) { m.runPaletteCommand(it.data) },
	}
	p.addItems(palette()...)
	return m.openPicker(p)
}

// runPaletteCommand ejecuta un comando elegido en la paleta o en las sugerencias, igual que si se hubiera pulsado su atajo.
func (m *Model) runPaletteCommand(name string) {
	cmd := m.dispatchByName(name, command.Args{})
	if cmd != nil {
		m.pendingCmds = append(m.pendingCmds, cmd)
	}
}

// openRecentPicker lista los archivos abiertos en esta sesión, el más reciente primero.
func (m *Model) openRecentPicker() tea.Cmd {
	p := &picker{
		kind: ui.PickerList, title: "Archivos recientes", prefix: ">", keep: true, hint: "↑↓ navegar   ⏎ abrir   Esc cerrar",
		accept: func(m *Model, it pickItem) { m.openPath(it.data, m.displayName(it.data)) },
	}
	for _, path := range m.recent {
		p.addItems(pickItem{text: m.displayName(path), data: path})
	}
	return m.openPicker(p)
}

// rememberFile anota un archivo abierto en la lista de recientes.
func (m *Model) rememberFile(path string) {
	abs := absPath(path)
	if abs == "" {
		return
	}
	out := []string{abs}
	for _, p := range m.recent {
		if p != abs && len(out) < recentFilesKept {
			out = append(out, p)
		}
	}
	m.recent = out
}

// ---- teclas ----

func (m *Model) pickerKey(k tea.KeyPressMsg) tea.Cmd {
	p := m.picker
	rows := ui.PickerRows(p.kind, m.width, m.height)
	switch key := k.String(); key {
	case "esc", "ctrl+c":
		m.closePicker()
	case "enter":
		m.pickerAccept()
	case "up", "ctrl+p":
		p.sel = max(0, p.sel-1)
	case "down", "ctrl+n":
		p.sel = min(max(0, len(p.results)-1), p.sel+1)
	case "pgup":
		p.sel = max(0, p.sel-rows)
	case "pgdown":
		p.sel = min(max(0, len(p.results)-1), p.sel+rows)
	case "left":
		p.cur = max(0, p.cur-1)
	case "right":
		p.cur = min(len(p.query), p.cur+1)
	case "home", "ctrl+a":
		p.cur = 0
	case "end", "ctrl+e":
		p.cur = len(p.query)
	case "backspace":
		if p.cur > 0 {
			p.query = append(p.query[:p.cur-1:p.cur-1], p.query[p.cur:]...)
			p.cur--
			return m.refilter(p)
		}
	case "delete":
		if p.cur < len(p.query) {
			p.query = append(p.query[:p.cur:p.cur], p.query[p.cur+1:]...)
			return m.refilter(p)
		}
	case "ctrl+u":
		if len(p.query) > 0 {
			p.query, p.cur = nil, 0
			return m.refilter(p)
		}
	case "ctrl+w": // borra la palabra anterior
		if p.cur > 0 {
			i := p.cur
			for i > 0 && p.query[i-1] == " " {
				i--
			}
			for i > 0 && p.query[i-1] != " " {
				i--
			}
			p.query = append(p.query[:i:i], p.query[p.cur:]...)
			p.cur = i
			return m.refilter(p)
		}
	default:
		if k.Text != "" && k.Mod&(tea.ModCtrl|tea.ModAlt) == 0 {
			return m.pickerInsert(k.Text)
		}
	}
	return nil
}

// pickerAccept acepta el elemento seleccionado: cierra la ventana y ejecuta la acción.
func (m *Model) pickerAccept() {
	p := m.picker
	if p == nil || len(p.results) == 0 {
		return
	}
	it := p.items[p.results[p.sel].Index]
	accept := p.accept
	m.closePicker()
	accept(m, it)
}

func (m *Model) pickerInsert(text string) tea.Cmd {
	p := m.picker
	cl := buffer.Clusters(strings.NewReplacer("\r", "", "\n", "", "\t", " ").Replace(text))
	if len(cl) == 0 {
		return nil
	}
	rest := append([]string(nil), p.query[p.cur:]...)
	p.query = append(append(p.query[:p.cur:p.cur], cl...), rest...)
	p.cur += len(cl)
	return m.refilter(p)
}

// ---- vista ----

// pickerView prepara la ventana para dibujarla: solo las filas visibles y la vista previa del archivo seleccionado.
func (m *Model) pickerView() *ui.PickerView {
	p := m.picker
	if p == nil {
		return nil
	}
	rows := ui.PickerRows(p.kind, m.width, m.height)
	if p.sel < p.scroll {
		p.scroll = p.sel
	}
	if p.sel >= p.scroll+rows {
		p.scroll = p.sel - rows + 1
	}
	end := min(len(p.results), p.scroll+rows)
	pv := &ui.PickerView{
		Kind: p.kind, Title: p.title, Prefix: p.prefix, Query: p.queryString(), Cursor: p.cur, Hint: p.hint,
		Counter: fmt.Sprintf("%d/%d", len(p.results), len(p.items)), Selected: p.sel - p.scroll,
	}
	if p.loading {
		pv.Counter += " …"
	}
	if len(p.results) == 0 {
		pv.Empty = "Sin resultados"
		if p.loading {
			pv.Empty = "Cargando…"
		}
	}
	for _, r := range p.results[p.scroll:end] {
		it := p.items[r.Index]
		row := ui.PickerRow{Text: it.text, Detail: it.detail}
		limit := len([]rune(it.text))
		for _, pos := range r.Pos {
			if pos < limit { // las coincidencias en el texto de apoyo (atajos) no se resaltan
				row.Match = append(row.Match, pos)
			}
		}
		pv.Rows = append(pv.Rows, row)
	}
	if p.kind == ui.PickerFiles && p.sel < len(p.results) {
		pv.Preview = m.preview(p, p.items[p.results[p.sel].Index].data)
	}
	return pv
}

// preview lee las primeras líneas de un archivo y las colorea. Los binarios y los archivos ilegibles muestran una nota.
func (m *Model) preview(p *picker, path string) *ui.PreviewView {
	if pv, ok := p.previews[path]; ok {
		return pv
	}
	pv := buildPreview(path)
	p.previews[path] = pv
	return pv
}

func buildPreview(path string) *ui.PreviewView {
	f, err := os.Open(path)
	if err != nil {
		return &ui.PreviewView{Note: "(no se puede leer el archivo)"}
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil || info.IsDir() {
		return &ui.PreviewView{Note: "(no es un archivo)"}
	}
	data := make([]byte, previewBytes)
	n, _ := f.Read(data)
	data = data[:n]
	probe := data[:min(len(data), 8192)]
	if strings.IndexByte(string(probe), 0) >= 0 {
		return &ui.PreviewView{Note: "(archivo binario: sin vista previa)"}
	}
	if !utf8.Valid(data) {
		// el corte de lectura puede partir un carácter al final: se prueba sin los últimos bytes
		if !utf8.Valid(data[:max(0, len(data)-3)]) {
			return &ui.PreviewView{Note: "(no es UTF-8: sin vista previa)"}
		}
		data = data[:max(0, len(data)-3)]
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) > previewLines {
		lines = lines[:previewLines]
	}
	pv := &ui.PreviewView{Lines: lines}
	if lang := syntax.Detect(path); lang != "" {
		pv.Spans, _ = syntax.Tokenize(lang, strings.Join(lines, "\n"))
	}
	return pv
}
