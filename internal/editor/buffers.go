package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// bufState es un buffer abierto con lo que hay que recordar de él: cursor, desplazamiento, nombre mostrado y
// los metadatos del archivo en disco cuando se cargó o guardó (para detectar cambios hechos desde fuera).
// El buffer activo vive también en los campos buf/cur/top/left/label del Model; stash() los vuelca aquí.
type bufState struct {
	buf       *buffer.Buffer
	cur       buffer.Cursor
	top, left int
	label     string
	mtime     time.Time
	size      int64
	onDisk    bool
	md        *mdState // estado de Markdown (solo en archivos .md y .markdown)
	id        int64    // identificador estable (los resultados asíncronos del resaltado lo usan para encontrar su buffer)
	hl        *hlState // resaltado de sintaxis
}

// snapshot guarda la fecha y el tamaño actuales del archivo en disco.
func (bs *bufState) snapshot() {
	bs.mtime, bs.size, bs.onDisk = time.Time{}, 0, false
	if p := bs.buf.Path(); p != "" {
		if info, err := os.Stat(p); err == nil {
			bs.mtime, bs.size, bs.onDisk = info.ModTime(), info.Size(), true
		}
	}
}

// changedOnDisk indica si el archivo cambió en disco desde que se cargó o guardó. Un archivo que ya no existe no cuenta.
func (bs *bufState) changedOnDisk() bool {
	p := bs.buf.Path()
	if p == "" || !bs.onDisk {
		return false
	}
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.ModTime().Equal(bs.mtime) || info.Size() != bs.size
}

func newBufState(b *buffer.Buffer, label string) *bufState {
	bs := &bufState{buf: b, cur: buffer.NewCursor(buffer.Pos{}), label: label, id: nextBufID.Add(1)}
	bs.snapshot()
	bs.initHighlight()
	bs.initMarkdown(true)
	return bs
}

func absPath(p string) string {
	if p == "" {
		return ""
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return filepath.Clean(p)
}

// stash vuelca el estado de trabajo (cursor, scroll, nombre) en el buffer activo de la lista.
func (m *Model) stash() {
	bs := m.bufs[m.active]
	bs.buf, bs.cur, bs.top, bs.left, bs.label = m.buf, m.cur, m.top, m.left, m.label
}

// load hace activo el buffer i copiando su estado a los campos de trabajo.
func (m *Model) load(i int) {
	m.active = i
	bs := m.bufs[i]
	m.buf, m.cur, m.top, m.left, m.label = bs.buf, bs.cur, bs.top, bs.left, bs.label
}

// activate cambia al buffer i recordando el cursor y el scroll del que se deja.
func (m *Model) activate(i int) {
	if i == m.active || i < 0 || i >= len(m.bufs) {
		return
	}
	m.stash()
	m.load(i)
	m.resetPending()
	m.setMode(ModeNormal)
	m.checkExternal()
}

// modifiedBuffers devuelve los buffers con cambios sin guardar.
func (m *Model) modifiedBuffers() []*bufState {
	m.stash()
	var out []*bufState
	for _, bs := range m.bufs {
		if bs.buf.Modified() {
			out = append(out, bs)
		}
	}
	return out
}

// names devuelve los nombres cortos de los buffers (el nombre del archivo; la ruta completa si hay nombres repetidos).
func names(list []*bufState) []string {
	out := make([]string, len(list))
	count := map[string]int{}
	for i, bs := range list {
		out[i] = filepath.Base(bs.label)
		if bs.label == "" {
			out[i] = "[Sin nombre]"
		}
		count[out[i]]++
	}
	for i, bs := range list {
		if count[out[i]] > 1 && bs.label != "" {
			out[i] = bs.label
		}
	}
	return out
}

// openPath abre un archivo reutilizando el buffer si ya está abierto (con su cursor y scroll). label es el nombre mostrado.
func (m *Model) openPath(path, label string) bool {
	abs := absPath(path)
	m.stash()
	for i, bs := range m.bufs {
		if absPath(bs.buf.Path()) == abs {
			m.rememberFile(path)
			m.activate(i)
			return true
		}
	}
	b, err := buffer.Open(path)
	if err != nil {
		m.msg, m.msgErr = "E484: no se puede abrir el archivo: "+err.Error(), true
		return false
	}
	bs := newBufState(b, label)
	if len(m.bufs) == 1 && isPristineEmpty(m.bufs[0].buf) { // el buffer vacío inicial se sustituye en lugar de acumularse
		m.bufs[0] = bs
		m.load(0)
	} else {
		m.bufs = append(m.bufs, bs)
		m.stash()
		m.load(len(m.bufs) - 1)
	}
	m.resetPending()
	m.setMode(ModeNormal)
	m.rememberFile(path)
	m.msg, m.msgErr = fmt.Sprintf("%q %dL", label, b.LineCount()), false
	if ro, why := b.ReadOnly(); ro {
		m.msg, m.msgErr = "Solo lectura: "+why, true
	}
	m.noticeMarkdown()
	return true
}

func isPristineEmpty(b *buffer.Buffer) bool {
	return b.Path() == "" && !b.Modified() && b.LineCount() == 1 && b.Line(0) == ""
}

// displayName es el nombre que se muestra de una ruta: relativa a la raíz del proyecto si está dentro de ella.
func (m *Model) displayName(path string) string {
	if m.root != "" {
		if rel, err := filepath.Rel(m.root, path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

// ---- cambios externos ----

// checkExternal revisa si el archivo del buffer activo cambió en disco: un buffer sin modificar se recarga con un aviso;
// uno modificado solo avisa (al guardar se pedirá confirmación).
func (m *Model) checkExternal() {
	bs := m.bufs[m.active]
	if !bs.changedOnDisk() {
		return
	}
	if !m.buf.Modified() {
		m.reloadActive(fmt.Sprintf("%q cambió en disco: recargado", m.label0()))
		return
	}
	m.msg, m.msgErr = fmt.Sprintf("W12: %q cambió en disco y tiene cambios sin guardar", m.label0()), true
}

// reloadActive vuelve a leer del disco el buffer activo, conservando el cursor en lo posible.
func (m *Model) reloadActive(notice string) {
	b, err := buffer.Open(m.buf.Path())
	if err != nil {
		m.msg, m.msgErr = "E484: no se puede recargar: "+err.Error(), true
		return
	}
	m.buf = b
	m.bufs[m.active].buf = b
	m.bufs[m.active].snapshot()
	m.bufs[m.active].initHighlight()
	m.bufs[m.active].initMarkdown(false)
	if md := m.bufs[m.active].md; md != nil {
		md.doc, md.docRev = nil, -1 // el texto recargado se vuelve a renderizar
	}
	m.cur.Set(m.buf, m.cur.Pos)
	m.clampCursor()
	if notice != "" {
		m.msg, m.msgErr = notice, false
	}
	if ro, why := b.ReadOnly(); ro {
		m.msg, m.msgErr = "Solo lectura: "+why, true
	}
}

// ---- cerrar y salir ----

func (m *Model) nextBuffer(step int) {
	if len(m.bufs) < 2 {
		m.msg = "No hay otros buffers abiertos"
		return
	}
	m.activate((m.active + step + len(m.bufs)) % len(m.bufs))
}

// closeActive cierra el buffer activo. Si era el último, queda un buffer vacío sin nombre y la aplicación sigue.
func (m *Model) closeActive(bang bool) {
	if m.buf.Modified() && !bang {
		d := unsavedDialog([]string{m.label0()},
			func(m *Model) {
				if m.save("") {
					m.removeActive()
				}
			},
			func(m *Model) { m.removeActive() })
		m.dialog = d
		return
	}
	m.removeActive()
}

func (m *Model) removeActive() {
	name := m.label0()
	i := m.active
	m.bufs = append(m.bufs[:i:i], m.bufs[i+1:]...)
	if len(m.bufs) == 0 {
		m.bufs = []*bufState{newBufState(buffer.New(""), "")}
	}
	m.load(max(0, i-1)) // se pasa al buffer anterior (o al primero)
	m.resetPending()
	m.setMode(ModeNormal)
	m.msg, m.msgErr = fmt.Sprintf("%q cerrado", name), false
	m.checkExternal()
}

// requestQuit sale si no hay cambios (o bang); si los hay, abre el diálogo con todos los buffers modificados.
func (m *Model) requestQuit(bang bool) {
	if bang {
		m.quit = true
		return
	}
	mod := m.modifiedBuffers()
	if len(mod) == 0 {
		m.quit = true
		return
	}
	m.msg, m.msgErr = "E37: cambios sin guardar (usa :q! para descartarlos)", true
	m.dialog = unsavedDialog(names(mod), func(m *Model) { m.saveAllAndQuit() }, func(m *Model) { m.quit = true })
}

// saveAllAndQuit guarda todos los buffers modificados y sale. Si alguno falla (o cambió en disco) se queda en él con el aviso.
func (m *Model) saveAllAndQuit() {
	for _, bs := range m.modifiedBuffers() {
		for i, other := range m.bufs {
			if other == bs {
				m.activate(i)
			}
		}
		m.batchSave = true
		ok := m.save("")
		m.batchSave = false
		if !ok {
			return
		}
	}
	m.quit = true
}
