package editor

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/fuzzy"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

type promptKind int

const (
	promptEx     promptKind = iota // «:» comandos
	promptSearch                   // «/» búsqueda
	promptCreate                   // explorador: nombre del archivo o carpeta nuevos
	promptRename                   // explorador: nuevo nombre
)

// prompt es la línea de comandos en edición: «:» para comandos, «/» para buscar o un nombre en el explorador.
type prompt struct {
	kind    promptKind
	target  string // carpeta (crear) o ruta (renombrar) sobre la que actúa el explorador
	prefix  string
	text    []string // grafemas
	cur     int      // posición del cursor en grafemas
	hist    []string // historial del tipo de prompt
	histIdx int      // posición en el historial (len(hist) = línea nueva)
	draft   string   // lo escrito antes de navegar el historial

	sugg    []fuzzy.Result // sugerencias de comandos para lo escrito (solo en «:»)
	suggSel int            // sugerencia elegida con ↑/↓ (-1: ninguna, se ejecuta el texto tal cual)
	scroll  int
}

func (p *prompt) String() string { return strings.Join(p.text, "") }

func (p *prompt) insert(s string) {
	s = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(s)
	cl := buffer.Clusters(s)
	p.text = slices.Insert(p.text, p.cur, cl...)
	p.cur += len(cl)
}

func (p *prompt) setText(s string) {
	p.text = buffer.Clusters(s)
	p.cur = len(p.text)
	p.refreshSuggestions()
}

// refreshSuggestions recalcula las sugerencias de comandos para lo escrito en el prompt «:».
// Sin texto no se sugiere nada, y escribir deshace cualquier elección anterior.
func (p *prompt) refreshSuggestions() {
	p.sugg, p.suggSel, p.scroll = nil, -1, 0
	if p.kind != promptEx || len(p.text) == 0 {
		return
	}
	p.sugg = fuzzy.Filter(p.String(), paletteTexts(), maxResults)
}

var paletteTextsCache []string

func paletteTexts() []string {
	if paletteTextsCache == nil {
		for _, it := range palette() {
			paletteTextsCache = append(paletteTextsCache, it.match)
		}
	}
	return paletteTextsCache
}

// suggestionView devuelve la ventana de sugerencias del prompt «:» (nil si no hay).
func (p *prompt) suggestionView(h, w int) *ui.PickerView {
	if len(p.sugg) == 0 {
		return nil
	}
	rows := ui.PickerRows(ui.PickerPalette, w, h)
	if p.suggSel >= 0 {
		if p.suggSel < p.scroll {
			p.scroll = p.suggSel
		}
		if p.suggSel >= p.scroll+rows {
			p.scroll = p.suggSel - rows + 1
		}
	}
	pv := &ui.PickerView{
		Kind: ui.PickerPalette, Title: "Comandos", Prefix: ":", Query: p.String(), Cursor: p.cur,
		Counter: fmt.Sprintf("%d/%d", len(p.sugg), len(palette())), Selected: p.suggSel - p.scroll,
		Hint: "↑↓ elegir   ⏎ ejecutar   Esc cerrar",
	}
	for _, r := range p.sugg[p.scroll:min(len(p.sugg), p.scroll+rows)] {
		it := palette()[r.Index]
		row := ui.PickerRow{Text: it.text, Detail: it.detail}
		for _, pos := range r.Pos {
			if pos < len([]rune(it.text)) {
				row.Match = append(row.Match, pos)
			}
		}
		pv.Rows = append(pv.Rows, row)
	}
	return pv
}

// openPrompt abre la línea de comandos «:» o la búsqueda «/» y entra en modo Command.
func (m *Model) openPrompt(prefix string) {
	kind, hist := promptEx, m.exHistory
	if prefix == "/" {
		kind, hist = promptSearch, m.searchHistory
	}
	m.prompt = &prompt{kind: kind, prefix: prefix, hist: hist, histIdx: len(hist)}
	m.setMode(ModeCommand)
}

// openPromptKind abre un prompt del explorador con un texto inicial.
func (m *Model) openPromptKind(kind promptKind, prefix, initial, target string) {
	m.prompt = &prompt{kind: kind, prefix: prefix, target: target}
	m.prompt.setText(initial)
	m.setMode(ModeCommand)
}

func (m *Model) closePrompt() {
	m.prompt = nil
	m.setMode(ModeNormal)
}

// promptKey procesa una tecla mientras la línea de comandos está abierta.
func (m *Model) promptKey(k tea.KeyPressMsg) {
	p := m.prompt
	switch k.String() {
	case "esc", "ctrl+c":
		m.closePrompt()
	case "enter":
		if p.kind == promptEx && p.suggSel >= 0 && p.suggSel < len(p.sugg) {
			name := palette()[p.sugg[p.suggSel].Index].data // se eligió una sugerencia con ↑/↓
			m.exHistory = appendHistory(m.exHistory, p.String())
			m.closePrompt()
			m.runPaletteCommand(name)
			return
		}
		line, kind, target := p.String(), p.kind, p.target
		if line != "" {
			switch kind {
			case promptSearch:
				m.searchHistory = appendHistory(m.searchHistory, line)
			case promptEx:
				m.exHistory = appendHistory(m.exHistory, line)
			}
		}
		m.closePrompt()
		switch kind {
		case promptSearch:
			m.search(line, true)
		case promptEx:
			m.runEx(line)
		case promptCreate:
			if strings.TrimSpace(line) != "" {
				m.finishCreate(target, line)
			}
		case promptRename:
			if strings.TrimSpace(line) != "" {
				m.finishRename(target, line)
			}
		}
	case "backspace":
		if len(p.text) == 0 {
			m.closePrompt() // Backspace en una línea vacía cancela, como en Vim
			return
		}
		if p.cur > 0 {
			p.text = slices.Delete(p.text, p.cur-1, p.cur)
			p.cur--
			p.refreshSuggestions()
		}
	case "delete":
		if p.cur < len(p.text) {
			p.text = slices.Delete(p.text, p.cur, p.cur+1)
			p.refreshSuggestions()
		}
	case "left":
		p.cur = max(0, p.cur-1)
	case "right":
		p.cur = min(len(p.text), p.cur+1)
	case "home", "ctrl+a":
		p.cur = 0
	case "end", "ctrl+e":
		p.cur = len(p.text)
	case "ctrl+u":
		p.text, p.cur = nil, 0
		p.refreshSuggestions()
	case "up", "ctrl+p":
		if len(p.sugg) > 0 && p.histIdx == len(p.hist) { // con sugerencias a la vista, ↑/↓ las recorren
			if p.suggSel <= 0 {
				p.suggSel = len(p.sugg) - 1
			} else {
				p.suggSel--
			}
			return
		}
		if p.histIdx == len(p.hist) {
			p.draft = p.String()
		}
		if p.histIdx > 0 {
			p.histIdx--
			p.setText(p.hist[p.histIdx])
		}
	case "down", "ctrl+n":
		if len(p.sugg) > 0 && p.histIdx == len(p.hist) {
			p.suggSel = (p.suggSel + 1) % len(p.sugg)
			return
		}
		if p.histIdx < len(p.hist) {
			p.histIdx++
			if p.histIdx == len(p.hist) {
				p.setText(p.draft)
			} else {
				p.setText(p.hist[p.histIdx])
			}
		}
	default:
		if k.Text != "" && k.Mod&(tea.ModCtrl|tea.ModAlt) == 0 {
			p.insert(k.Text)
			p.refreshSuggestions()
		}
	}
}

// appendHistory añade una línea al historial evitando repetir la última.
func appendHistory(h []string, line string) []string {
	if len(h) > 0 && h[len(h)-1] == line {
		return h
	}
	h = append(h[:len(h):len(h)], line)
	if len(h) > 100 {
		h = h[len(h)-100:]
	}
	return h
}
