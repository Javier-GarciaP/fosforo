package editor

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// prompt es la línea de comandos en edición: «:» para comandos o «/» para buscar.
type prompt struct {
	prefix  string
	text    []string // grafemas
	cur     int      // posición del cursor en grafemas
	hist    []string // historial del tipo de prompt
	histIdx int      // posición en el historial (len(hist) = línea nueva)
	draft   string   // lo escrito antes de navegar el historial
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
}

// openPrompt abre la línea de comandos con el prefijo dado y entra en modo Command.
func (m *Model) openPrompt(prefix string) {
	hist := m.exHistory
	if prefix == "/" {
		hist = m.searchHistory
	}
	m.prompt = &prompt{prefix: prefix, hist: hist, histIdx: len(hist)}
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
		line := p.String()
		prefix := p.prefix
		if line != "" {
			if prefix == "/" {
				m.searchHistory = appendHistory(m.searchHistory, line)
			} else {
				m.exHistory = appendHistory(m.exHistory, line)
			}
		}
		m.closePrompt()
		if prefix == "/" {
			m.search(line, true)
		} else {
			m.runEx(line)
		}
	case "backspace":
		if len(p.text) == 0 {
			m.closePrompt() // Backspace en una línea vacía cancela, como en Vim
			return
		}
		if p.cur > 0 {
			p.text = slices.Delete(p.text, p.cur-1, p.cur)
			p.cur--
		}
	case "delete":
		if p.cur < len(p.text) {
			p.text = slices.Delete(p.text, p.cur, p.cur+1)
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
	case "up", "ctrl+p":
		if p.histIdx == len(p.hist) {
			p.draft = p.String()
		}
		if p.histIdx > 0 {
			p.histIdx--
			p.setText(p.hist[p.histIdx])
		}
	case "down", "ctrl+n":
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
