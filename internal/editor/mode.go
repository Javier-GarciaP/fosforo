package editor

// Mode es el modo de edición.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeVisualLine
	ModeCommand // línea de comandos («:») o búsqueda («/») en edición
)

// Label es el texto que muestra la statusline.
func (m Mode) Label() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeVisualLine:
		return "V-LINE"
	case ModeCommand:
		return "COMMAND"
	}
	return "NORMAL"
}

func (m Mode) String() string { return m.Label() }

// keymapName es el nombre del árbol de teclas de este modo.
func (m Mode) keymapName() string {
	switch m {
	case ModeInsert:
		return "insert"
	case ModeVisual, ModeVisualLine:
		return "visual"
	}
	return "normal"
}

func (m Mode) isVisual() bool { return m == ModeVisual || m == ModeVisualLine }

// modeOperator es el árbol de teclas que se usa mientras hay un operador pendiente (d, c, y): solo movimientos.
const modeOperator = "operator"

// modeExplorer es el contexto de teclas del panel del explorador.
const modeExplorer = "explorer"
