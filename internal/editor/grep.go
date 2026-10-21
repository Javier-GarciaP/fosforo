package editor

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/fuzzy"
	"github.com/Javier-GarciaP/fosforo/internal/search"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// DefaultGrepDelay es lo que se espera tras la última tecla antes de lanzar la búsqueda en el proyecto.
const DefaultGrepDelay = 150 * time.Millisecond

// grepState es el estado de una búsqueda de texto en el proyecto dentro de la ventana de selección.
type grepState struct {
	root    string
	gen     int // generación de la última búsqueda pedida
	cancel  context.CancelFunc
	ch      <-chan grepMsg
	running bool
	trunc   bool
	backend string
}

// grepStartMsg lo envía el temporizador del debounce: lanza la búsqueda si sigue siendo la última pedida.
type grepStartMsg struct {
	id  int64
	gen int
}

// grepMsg lleva un lote de coincidencias (done indica que no habrá más).
type grepMsg struct {
	id      int64
	gen     int
	batch   []search.Match
	done    bool
	trunc   bool
	backend string
}

// WithGrepDelay cambia la espera antes de lanzar la búsqueda en el proyecto (por defecto DefaultGrepDelay).
func WithGrepDelay(d time.Duration) Option { return func(m *Model) { m.grepWait = d } }

// openGrepPicker abre «Buscar texto en el proyecto». Con initial no vacío busca de inmediato esa palabra.
func (m *Model) openGrepPicker(title, initial string) tea.Cmd {
	g := &grepState{root: m.rootDir()}
	p := &picker{
		kind: ui.PickerSearch, title: title, prefix: "/", grep: g,
		hint: "escribe para buscar   ↑↓ navegar   ⏎ abrir en la línea   Esc cerrar",
		accept: func(m *Model, it pickItem) {
			parts := strings.Split(it.data, "\x00")
			if len(parts) != 3 {
				return
			}
			line, _ := strconv.Atoi(parts[1])
			col, _ := strconv.Atoi(parts[2])
			m.jumpTo(filepath.Join(g.root, filepath.FromSlash(parts[0])), line-1, col)
		},
	}
	if initial != "" {
		p.query = buffer.Clusters(initial)
		p.cur = len(p.query)
	}
	cmd := m.openPicker(p)
	if initial != "" {
		return m.grepSchedule(p, true)
	}
	return cmd
}

// jumpTo abre path y coloca el cursor en (línea, columna), centrando la línea si hace falta.
func (m *Model) jumpTo(path string, line, col int) {
	if !m.openPath(path, m.displayName(path)) {
		return
	}
	m.focus = FocusEditor
	if md := m.mdState(); md != nil && md.view != mdSource {
		md.view = mdSource // un resultado de búsqueda es una línea del fuente
	}
	m.setMode(ModeNormal)
	line = max(0, min(line, m.buf.LineCount()-1))
	m.cur.Set(m.buf, buffer.Pos{Line: line, Col: col})
	m.top = max(0, line-ui.TextRows(m.height)/2)
	m.scroll()
}

// grepSchedule recalcula la búsqueda tras escribir: cancela la anterior y espera a que se deje de teclear.
func (m *Model) grepSchedule(p *picker, immediate bool) tea.Cmd {
	g := p.grep
	g.gen++
	if g.cancel != nil {
		g.cancel()
		g.cancel = nil
	}
	g.running, g.trunc = false, false
	p.items, p.texts, p.results, p.sel, p.scroll = nil, nil, nil, 0, 0
	p.loading = false
	if p.queryString() == "" {
		return nil
	}
	p.loading = true
	id, gen := p.id, g.gen
	if immediate {
		return func() tea.Msg { return grepStartMsg{id: id, gen: gen} }
	}
	return tea.Tick(m.grepWait, func(time.Time) tea.Msg { return grepStartMsg{id: id, gen: gen} })
}

// grepStart lanza la búsqueda en segundo plano; los lotes llegan por un canal, uno a uno, como mensajes.
func (m *Model) grepStart(msg grepStartMsg) tea.Cmd {
	p := m.picker
	if p == nil || p.id != msg.id || p.grep == nil || p.grep.gen != msg.gen {
		return nil
	}
	g := p.grep
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel, g.running = cancel, true
	ch := make(chan grepMsg, 8)
	g.ch = ch
	opts := search.Options{Root: g.root, Query: p.queryString()}
	id, gen := p.id, g.gen
	go func() {
		defer close(ch)
		send := func(x grepMsg) bool {
			select {
			case ch <- x:
				return true
			case <-ctx.Done():
				return false
			}
		}
		res, err := search.Run(ctx, opts, func(batch []search.Match) {
			send(grepMsg{id: id, gen: gen, batch: append([]search.Match(nil), batch...)})
		})
		if ctx.Err() == nil {
			_ = err
			send(grepMsg{id: id, gen: gen, done: true, trunc: res.Truncated, backend: res.Backend})
		}
	}()
	return waitGrep(id, gen, ch)
}

func waitGrep(id int64, gen int, ch <-chan grepMsg) tea.Cmd {
	return func() tea.Msg {
		x, ok := <-ch
		if !ok {
			return grepMsg{id: id, gen: gen, done: true}
		}
		return x
	}
}

// grepArrived añade un lote de coincidencias a la lista.
func (m *Model) grepArrived(msg grepMsg) tea.Cmd {
	p := m.picker
	if p == nil || p.id != msg.id || p.grep == nil || p.grep.gen != msg.gen {
		return nil
	}
	g := p.grep
	for _, mt := range msg.batch {
		text := fmt.Sprintf("%s:%d  %s", mt.Path, mt.Line, strings.TrimSpace(strings.ReplaceAll(mt.Text, "\t", " ")))
		p.addItems(pickItem{text: text, data: fmt.Sprintf("%s\x00%d\x00%d", mt.Path, mt.Line, mt.Col)})
		p.results = append(p.results, fuzzy.Result{Index: len(p.items) - 1})
	}
	if msg.done {
		p.loading, g.running, g.trunc, g.backend = false, false, msg.trunc, msg.backend
		return nil
	}
	return waitGrep(msg.id, msg.gen, g.ch)
}

// grepCounter es el contador de la ventana de búsqueda.
func (p *picker) grepCounter() string {
	c := fmt.Sprintf("%d", len(p.items))
	if p.grep.trunc {
		c += "+ (límite)"
	}
	if p.loading {
		c += " …"
	}
	return c
}

func (m *Model) grepWordUnderCursor() tea.Cmd {
	a, b := m.wordAt(m.cur.Pos)
	cl := buffer.Clusters(m.buf.Line(a.Line))
	word := ""
	if a.Line == b.Line && a.Col < len(cl) {
		word = strings.TrimSpace(strings.Join(cl[a.Col:min(b.Col+1, len(cl))], ""))
	}
	return m.openGrepPicker("Buscar palabra", word)
}
