package editor

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/command"
	"github.com/Javier-GarciaP/fosforo/internal/keymap"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// registry y defaultKeymap se construyen una vez: toda acción del editor es un comando con nombre en el registro,
// y las teclas solo apuntan a esos nombres.
var (
	registry      *command.Registry[*Model]
	defaultKeymap *keymap.Keymap
)

// Se construyen en init (y no en la declaración) porque algunos comandos, a través de la paleta, vuelven a consultar el registro.
func init() {
	registry = buildRegistry()
	defaultKeymap = buildKeymap()
}

// Registry devuelve el registro de comandos del editor (para la paleta, la ayuda y las pruebas).
func Registry() *command.Registry[*Model] { return registry }

func displayKeys(keys ...string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = strings.ReplaceAll(k, " ", "")
	}
	return out
}

// shortLabels son las etiquetas cortas que usa el menú de la tecla líder (si un comando no tiene, se usa su descripción).
var shortLabels = map[string]string{
	"explorer.toggle": "explorador (alternar)", "file.save": "guardar", "buffer.close": "cerrar buffer", "app.quit": "salir",
	"theme.toggle": "tema claro / oscuro", "palette.open": "paleta de comandos", "mode.search": "buscar en el archivo",
	"mode.command": "línea de comandos", "help.keys": "lista de atajos", "file.find": "buscar archivos",
	"file.recent": "abrir reciente", "file.new": "nuevo archivo", "buffer.pick": "buscar buffers",
	"buffer.next": "buffer siguiente", "buffer.prev": "buffer anterior",
	"markdown.toggle": "vista previa / fuente", "markdown.split": "fuente | vista previa",
}

func buildRegistry() *command.Registry[*Model] {
	r := command.NewRegistry[*Model]()
	reg := func(name, desc string, keys []string, run func(m *Model, a command.Args) tea.Cmd) {
		r.MustRegister(command.Command[*Model]{Name: name, Desc: desc, Short: shortLabels[name], Keys: keys, Run: run})
	}
	none := func(f func(m *Model, a command.Args)) func(*Model, command.Args) tea.Cmd {
		return func(m *Model, a command.Args) tea.Cmd { f(m, a); return nil }
	}

	for _, d := range motionDefs {
		d := d
		reg(d.name, d.desc, displayKeys(d.keys...), none(func(m *Model, a command.Args) { m.moveBy(d, a) }))
	}
	reg("cursor.goto_line", "Ir a la línea {n} (:n)", []string{":n"}, none(func(m *Model, a command.Args) {
		l := clampLine(m, a.Count-1)
		m.cur.Set(m.buf, buffer.Pos{Line: l, Col: firstNonBlank(m.buf.Line(l))})
	}))

	// modos
	reg("mode.normal", "Volver al modo Normal", []string{"Esc", "Ctrl+C"}, none(func(m *Model, _ command.Args) { m.escape() }))
	reg("mode.insert", "Insertar antes del cursor", []string{"i"}, none(func(m *Model, _ command.Args) { m.setMode(ModeInsert) }))
	reg("mode.insert_after", "Insertar después del cursor", []string{"a"}, none(func(m *Model, _ command.Args) {
		m.cur.Set(m.buf, buffer.Pos{Line: m.cur.Line, Col: min(m.cur.Col+1, m.buf.LineLen(m.cur.Line))})
		m.setMode(ModeInsert)
	}))
	reg("mode.insert_line_start", "Insertar al inicio de la línea", []string{"I"}, none(func(m *Model, _ command.Args) {
		m.cur.Set(m.buf, buffer.Pos{Line: m.cur.Line, Col: firstNonBlankOrStart(m.buf.Line(m.cur.Line))})
		m.setMode(ModeInsert)
	}))
	reg("mode.insert_line_end", "Insertar al final de la línea", []string{"A"}, none(func(m *Model, _ command.Args) {
		m.cur.End(m.buf)
		m.setMode(ModeInsert)
	}))
	reg("mode.open_below", "Abrir una línea debajo e insertar", []string{"o"}, none(func(m *Model, _ command.Args) { m.openLine(true) }))
	reg("mode.open_above", "Abrir una línea encima e insertar", []string{"O"}, none(func(m *Model, _ command.Args) { m.openLine(false) }))
	reg("mode.visual", "Modo Visual por caracteres", []string{"v"}, none(func(m *Model, _ command.Args) { m.toggleVisual(ModeVisual) }))
	reg("mode.visual_line", "Modo Visual por líneas", []string{"V"}, none(func(m *Model, _ command.Args) { m.toggleVisual(ModeVisualLine) }))
	reg("mode.command", "Abrir la línea de comandos", []string{":"}, none(func(m *Model, _ command.Args) { m.openPrompt(":") }))
	reg("mode.search", "Buscar en el archivo", []string{"/"}, none(func(m *Model, _ command.Args) { m.openPrompt("/") }))
	reg("visual.swap_end", "Intercambiar el extremo de la selección", []string{"o"}, none(func(m *Model, _ command.Args) {
		if m.mode.isVisual() {
			m.anchor, m.cur.Pos = m.cur.Pos, m.anchor
		}
	}))

	// escritura (Insert)
	// Los comandos de escritura solo tienen sentido en Insert; si se lanzan desde la paleta en otro modo no hacen nada.
	reg("edit.newline", "Salto de línea conservando la sangría", []string{"Enter"}, none(func(m *Model, _ command.Args) {
		if m.mode == ModeInsert {
			m.insertText("\n" + m.indentBeforeCursor())
		}
	}))
	reg("edit.tab", "Insertar un tabulador", []string{"Tab"}, none(func(m *Model, _ command.Args) {
		if m.mode == ModeInsert {
			m.insertText("\t")
		}
	}))
	reg("edit.backspace", "Borrar el carácter anterior", []string{"Backspace"}, none(func(m *Model, _ command.Args) {
		if m.mode == ModeInsert {
			m.backspace()
		}
	}))
	reg("edit.delete_forward", "Borrar el carácter siguiente", []string{"Delete"}, none(func(m *Model, _ command.Args) {
		if m.mode == ModeInsert {
			m.deleteForward()
		}
	}))

	// edición en Normal y Visual
	reg("edit.undo", "Deshacer", []string{"u"}, none(func(m *Model, _ command.Args) { m.undo(false) }))
	reg("edit.redo", "Rehacer", []string{"Ctrl+R"}, none(func(m *Model, _ command.Args) { m.undo(true) }))
	reg("edit.paste_after", "Pegar después del cursor", []string{"p"}, none(func(m *Model, a command.Args) { m.paste(true, a.Count) }))
	reg("edit.paste_before", "Pegar antes del cursor", []string{"P"}, none(func(m *Model, a command.Args) { m.paste(false, a.Count) }))
	reg("edit.delete_char", "Borrar el carácter bajo el cursor", []string{"x"}, func(m *Model, a command.Args) tea.Cmd {
		if m.mode.isVisual() {
			return m.applyOperator("d", m.visualRange())
		}
		l := m.cur.Line
		n := min(m.cur.Col+a.N(), m.buf.LineLen(l))
		return m.applyOperator("d", textRange{start: m.cur.Pos, end: buffer.Pos{Line: l, Col: n}})
	})
	reg("edit.delete_to_end", "Borrar hasta el final de la línea", []string{"D"}, func(m *Model, _ command.Args) tea.Cmd {
		return m.applyOperator("d", textRange{start: m.cur.Pos, end: buffer.Pos{Line: m.cur.Line, Col: m.buf.LineLen(m.cur.Line)}})
	})
	reg("edit.change_to_end", "Cambiar hasta el final de la línea", []string{"C"}, func(m *Model, _ command.Args) tea.Cmd {
		return m.applyOperator("c", textRange{start: m.cur.Pos, end: buffer.Pos{Line: m.cur.Line, Col: m.buf.LineLen(m.cur.Line)}})
	})
	for _, op := range []struct{ name, key, desc string }{
		{"operator.delete", "d", "Borrar (con un movimiento o una selección)"},
		{"operator.change", "c", "Cambiar (con un movimiento o una selección)"},
		{"operator.yank", "y", "Copiar (con un movimiento o una selección)"},
	} {
		op := op
		reg(op.name, op.desc, []string{op.key, op.key + op.key}, func(m *Model, a command.Args) tea.Cmd {
			if m.mode.isVisual() {
				return m.applyOperator(op.key, m.visualRange())
			}
			m.op, m.opCount = op.key, a.Count
			return nil
		})
	}

	// búsqueda
	reg("search.next", "Siguiente coincidencia", []string{"n"}, none(func(m *Model, _ command.Args) { m.search("", m.lastForward) }))
	reg("search.prev", "Coincidencia anterior", []string{"N"}, none(func(m *Model, _ command.Args) { m.search("", !m.lastForward) }))

	registerExplorerCommands(r)
	registerMarkdownCommands(r)
	reg("buffer.next", "Ir al buffer siguiente", []string{"Space b n", ":bn"}, none(func(m *Model, _ command.Args) { m.nextBuffer(1) }))
	reg("buffer.prev", "Ir al buffer anterior", []string{"Space b p", ":bp"}, none(func(m *Model, _ command.Args) { m.nextBuffer(-1) }))
	reg("buffer.close", "Cerrar buffer", []string{"Space b d", ":bd"}, none(func(m *Model, a command.Args) { m.closeActive(a.Bang) }))

	// archivo y aplicación
	reg("file.save", "Guardar archivo", []string{"Ctrl+S", ":w"}, none(func(m *Model, a command.Args) { m.save(a.Text) }))
	reg("file.save_quit", "Guardar y salir", []string{":wq"}, none(func(m *Model, a command.Args) {
		if m.save(a.Text) {
			m.requestQuit(false) // sale salvo que otros buffers tengan cambios
		}
	}))
	reg("file.update_quit", "Guardar si hay cambios y salir", []string{":x"}, none(func(m *Model, a command.Args) {
		if !m.buf.Modified() || m.save(a.Text) {
			m.requestQuit(false)
		}
	}))
	reg("theme.toggle", "Alternar entre el tema oscuro y el claro", []string{"Space t"}, none(func(m *Model, _ command.Args) { m.toggleTheme() }))
	reg("file.find", "Buscar archivos del proyecto", []string{"Space f f"}, func(m *Model, _ command.Args) tea.Cmd { return m.openFilePicker() })
	reg("file.recent", "Abrir un archivo reciente", []string{"Space f r"}, func(m *Model, _ command.Args) tea.Cmd { return m.openRecentPicker() })
	reg("file.new", "Nuevo archivo", []string{"Space f n"}, none(func(m *Model, _ command.Args) {
		m.openPrompt(":")
		m.prompt.insert("e ")
	}))
	reg("buffer.pick", "Buscar entre los buffers abiertos", []string{"Space b b"}, func(m *Model, _ command.Args) tea.Cmd { return m.openBufferPicker() })
	reg("palette.open", "Paleta de comandos", []string{"Space Space"}, func(m *Model, _ command.Args) tea.Cmd { return m.openPalette("Paleta de comandos") })
	reg("help.keys", "Lista de atajos", []string{"Space ?", "Space h"}, func(m *Model, _ command.Args) tea.Cmd { return m.openPalette("Atajos") })
	reg("file.open", "Abrir archivo", []string{":e"}, none(func(m *Model, a command.Args) { m.openFile(a) }))
	reg("app.quit", "Salir", []string{":q"}, none(func(m *Model, a command.Args) { m.requestQuit(a.Bang) }))
	return r
}

func firstNonBlankOrStart(line string) int {
	if strings.TrimSpace(line) == "" {
		return buffer.Count(line)
	}
	return firstNonBlank(line)
}

// specialKeys son las teclas de movimiento que también funcionan en Insert (el resto de teclas escriben texto).
var specialKeys = map[string]bool{"left": true, "right": true, "up": true, "down": true, "home": true, "end": true, "pgup": true, "pgdown": true}

func buildKeymap() *keymap.Keymap {
	km := keymap.New()
	for _, d := range motionDefs {
		for _, k := range d.keys {
			for _, mode := range []string{"normal", "visual", modeOperator} {
				km.MustBind(mode, k, d.name)
			}
			if specialKeys[k] {
				km.MustBind("insert", k, d.name)
			}
		}
	}
	bind := func(modes []string, key, cmd string) {
		for _, mode := range modes {
			km.MustBind(mode, key, cmd)
		}
	}
	normal, visual, both := []string{"normal"}, []string{"visual"}, []string{"normal", "visual"}
	for key, cmd := range map[string]string{
		"i": "mode.insert", "a": "mode.insert_after", "I": "mode.insert_line_start", "A": "mode.insert_line_end",
		"o": "mode.open_below", "O": "mode.open_above", ":": "mode.command", "/": "mode.search",
		"u": "edit.undo", "ctrl+r": "edit.redo", "p": "edit.paste_after", "P": "edit.paste_before",
		"D": "edit.delete_to_end", "C": "edit.change_to_end", "n": "search.next", "N": "search.prev",
	} {
		bind(normal, key, cmd)
	}
	bind(visual, "o", "visual.swap_end")
	for key, cmd := range map[string]string{
		"v": "mode.visual", "V": "mode.visual_line", "x": "edit.delete_char",
		"d": "operator.delete", "c": "operator.change", "y": "operator.yank", "ctrl+s": "file.save",
	} {
		bind(both, key, cmd)
	}
	// La tecla líder (Space): las mismas secuencias en el editor y en el explorador. El panel de ayuda se genera de este árbol.
	for key, cmd := range map[string]string{
		"space space": "palette.open", "space f f": "file.find", "space f r": "file.recent", "space f n": "file.new",
		"space e": "explorer.toggle", "space b b": "buffer.pick", "space b n": "buffer.next", "space b p": "buffer.prev",
		"space b d": "buffer.close", "space w": "file.save", "space q": "buffer.close", "space Q": "app.quit",
		"space t": "theme.toggle", "space /": "mode.search", "space :": "mode.command", "space h": "help.keys", "space ?": "help.keys",
		"space m p": "markdown.toggle", "space m s": "markdown.split",
	} {
		bind([]string{"normal", modeExplorer, modePreview}, key, cmd)
	}
	// vista previa de Markdown a pantalla completa: solo lectura, con su propio juego de teclas
	for key, cmd := range map[string]string{
		"j": "preview.down", "down": "preview.down", "k": "preview.up", "up": "preview.up",
		"ctrl+d": "preview.half_down", "ctrl+u": "preview.half_up", "pgdown": "preview.page_down", "pgup": "preview.page_up",
		"g g": "preview.top", "home": "preview.top", "G": "preview.bottom", "end": "preview.bottom",
		"e": "markdown.toggle", ":": "mode.command", "ctrl+s": "file.save", "tab": "focus.toggle", "ctrl+w": "focus.toggle",
	} {
		bind([]string{modePreview}, key, cmd)
	}
	for key, cmd := range map[string]string{"tab": "focus.toggle", "ctrl+w": "focus.toggle"} {
		bind(normal, key, cmd)
	}
	for key, cmd := range map[string]string{
		"j": "explorer.down", "down": "explorer.down", "k": "explorer.up", "up": "explorer.up",
		"g g": "explorer.top", "G": "explorer.bottom",
		"enter": "explorer.open", "l": "explorer.open", "right": "explorer.open", "h": "explorer.collapse", "left": "explorer.collapse",
		"a": "explorer.create", "r": "explorer.rename", "d": "explorer.delete", "R": "explorer.refresh",
		"tab": "focus.toggle", "ctrl+w": "focus.toggle",
		":": "mode.command", "ctrl+s": "file.save",
	} {
		bind([]string{modeExplorer}, key, cmd)
	}
	for key, cmd := range map[string]string{
		"enter": "edit.newline", "tab": "edit.tab", "backspace": "edit.backspace", "delete": "edit.delete_forward", "ctrl+s": "file.save",
	} {
		bind([]string{"insert"}, key, cmd)
	}
	return km
}

// ---- implementación de los comandos ----

// moveBy mueve el cursor con una motion.
func (m *Model) moveBy(d motionDef, a command.Args) {
	mot := d.fn(m, a.N(), a.HasCount, false)
	if mot.failed {
		return
	}
	if mot.vertical != 0 {
		m.cur.Vertical(m.buf, mot.vertical)
	} else {
		m.cur.Set(m.buf, mot.pos)
	}
}

// escape cancela lo pendiente y vuelve a Normal (desde Insert el cursor retrocede un carácter, como en Vim).
func (m *Model) escape() {
	m.resetPending()
	if m.mode == ModeInsert && m.cur.Col > 0 {
		m.cur.Set(m.buf, buffer.Pos{Line: m.cur.Line, Col: m.cur.Col - 1})
	}
	m.setMode(ModeNormal)
}

func (m *Model) toggleVisual(mode Mode) {
	if m.mode == mode {
		m.setMode(ModeNormal)
		return
	}
	if !m.mode.isVisual() {
		m.anchor = m.cur.Pos
	}
	m.setMode(mode)
}

// openLine abre una línea nueva debajo o encima de la actual, con su sangría, y entra en Insert.
func (m *Model) openLine(below bool) {
	line := m.cur.Line
	indent := ""
	for _, c := range buffer.Clusters(m.buf.Line(line)) {
		if c != " " && c != "\t" {
			break
		}
		indent += c
	}
	var at buffer.Pos
	var text string
	target := line
	if below {
		at, text, target = buffer.Pos{Line: line, Col: m.buf.LineLen(line)}, "\n"+indent, line+1
	} else {
		at, text = buffer.Pos{Line: line}, indent+"\n"
	}
	if _, _, err := m.buf.Insert(at, text); m.edit(err) {
		m.cur.Set(m.buf, buffer.Pos{Line: target, Col: buffer.Count(indent)})
		m.setMode(ModeInsert)
	}
}

func (m *Model) undo(redo bool) {
	var pos buffer.Pos
	var ok bool
	if redo {
		pos, ok = m.buf.Redo()
	} else {
		pos, ok = m.buf.Undo()
	}
	switch {
	case ok:
		m.cur.Set(m.buf, pos)
	case redo:
		m.msg = "Ya estás en el cambio más reciente"
	default:
		m.msg = "Ya estás en el cambio más antiguo"
	}
}

func (m *Model) label0() string {
	if m.label == "" {
		return "[Sin nombre]"
	}
	return m.label
}

// save guarda el buffer (en path si se indica) y deja el resultado en la línea de comandos. Devuelve si se guardó.
// Si el archivo cambió en disco desde que se cargó y hay cambios sin guardar, pide confirmación antes de sobrescribir.
func (m *Model) save(path string) bool {
	if path == "" && m.buf.Path() != "" && m.buf.Modified() && m.bufs[m.active].changedOnDisk() {
		if m.batchSave {
			m.msg, m.msgErr = fmt.Sprintf("E13: %q cambió en disco; guarda ese archivo con :w para confirmar", m.label0()), true
			return false
		}
		m.dialog = &dialog{
			title:   "El archivo cambió en disco",
			lines:   []string{fmt.Sprintf("%q se modificó fuera del editor.", m.label0()), "¿Sobrescribirlo con tus cambios?"},
			buttons: []dialogButton{btnOverride, btnCancel},
			choose: func(m *Model, i int) {
				if i == 0 {
					m.writeBuffer("")
				}
			},
		}
		return false
	}
	return m.writeBuffer(path)
}

// writeBuffer escribe el buffer en disco sin comprobar cambios externos.
func (m *Model) writeBuffer(path string) bool {
	var err error
	if path != "" {
		err = m.buf.SaveAs(path)
		if err == nil {
			m.label = path
		}
	} else {
		err = m.buf.Save()
	}
	if err != nil {
		m.reportSaveError(err)
		return false
	}
	m.stash()
	m.bufs[m.active].snapshot()
	if lang := syntax.Detect(m.buf.Path()); lang != m.bufs[m.active].hl.lang {
		m.bufs[m.active].detectLanguage() // «guardar como» con otra extensión cambia el lenguaje
	}
	m.refreshExplorer()
	m.msg, m.msgErr = fmt.Sprintf("%q %dL escritas", m.label, m.buf.LineCount()), false
	return true
}

// openFile implementa :e. Con ruta abre el archivo en un buffer (reutilizando el que ya esté abierto);
// sin ruta recarga el actual desde el disco (con «!» descarta los cambios).
func (m *Model) openFile(a command.Args) {
	if a.Text != "" {
		m.openPath(a.Text, a.Text)
		return
	}
	if m.buf.Path() == "" {
		m.msg, m.msgErr = "E32: sin nombre de archivo", true
		return
	}
	if m.buf.Modified() && !a.Bang {
		m.msg, m.msgErr = "E37: cambios sin guardar (usa :e! para descartarlos)", true
		return
	}
	m.reloadActive(fmt.Sprintf("%q recargado", m.label0()))
}
