package ui

// StartEntry es una acción de la pantalla de inicio con su atajo del líder.
type StartEntry struct {
	Label string
	Keys  string
}

// StartEntries son las ocho acciones que se muestran en la pantalla de inicio.
// Los atajos son texto fijo: un test en internal/editor comprueba que cada uno
// sigue asignado en el mapa de teclas al comando que representa.
var StartEntries = []StartEntry{
	{"Buscar archivo", "Space f f"},
	{"Nuevo archivo", "Space f n"},
	{"Archivos recientes", "Space f r"},
	{"Buscar texto", "Space s g"},
	{"Explorador", "Space e"},
	{"Configuración", "Space c o"},
	{"Lista de atajos", "Space ?"},
	{"Salir", "Space Q"},
}

// startLogo es la palabra FOSFORO en letras de bloque con sombra (█ y trazos de caja): seis filas de alto, 59 de ancho.
var startLogo = []string{
	"███████╗ ██████╗ ███████╗███████╗ ██████╗ ██████╗  ██████╗ ",
	"██╔════╝██╔═══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔═══██╗",
	"█████╗  ██║   ██║███████╗█████╗  ██║   ██║██████╔╝██║   ██║",
	"██╔══╝  ██║   ██║╚════██║██╔══╝  ██║   ██║██╔══██╗██║   ██║",
	"██║     ╚██████╔╝███████║██║     ╚██████╔╝██║  ██║╚██████╔╝",
	"╚═╝      ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ",
}

// startGap es el hueco mínimo entre una etiqueta y su atajo.
const startGap = 4

// startLogoWidth es el ancho del logo en celdas.
func startLogoWidth() int {
	w := 0
	for _, line := range startLogo {
		w = max(w, textWidth(line))
	}
	return w
}

// startListWidth es el ancho de la lista: etiqueta más ancha, hueco y atajo más ancho.
func startListWidth() int {
	labels, keys := 0, 0
	for _, e := range StartEntries {
		labels = max(labels, textWidth(e.Label))
		keys = max(keys, textWidth(e.Keys))
	}
	return labels + startGap + keys
}

// startFits indica si la lista de acciones cabe en un área de w×h celdas (el mínimo para mostrar la pantalla de inicio).
func startFits(w, h int) bool { return w >= startListWidth() && h >= len(StartEntries) }

// drawStart pinta la pantalla de inicio centrada en el rectángulo (x, y, w, h) del editor.
// Tres niveles (decisión 4): si caben logo, hueco y lista se dibuja todo; si no, solo la
// lista; y si la lista no cabe (en filas o en columnas) no se dibuja nada y devuelve false,
// de modo que el editor quede como un editor vacío normal.
func drawStart(g *grid, x, y, w, h int, th Theme) bool {
	n := len(StartEntries)
	listW := startListWidth()
	if !startFits(w, h) {
		return false
	}
	full := h >= len(startLogo)+1+n && w >= startLogoWidth()
	rows := n
	logoSt := style{fg: th.Brand, bg: th.BG, bold: true}
	shadowSt := style{fg: th.Dim, bg: th.BG}
	labelSt := style{fg: th.Blue, bg: th.BG}
	keysSt := style{fg: th.Peach, bg: th.BG}
	top := y + (h-rows)/2
	if full {
		rows = len(startLogo) + 1 + n
		top = y + (h-rows)/2
		lx := x + (w-startLogoWidth())/2
		for i, line := range startLogo {
			drawLogoLine(g, lx, top+i, line, logoSt, shadowSt)
		}
	}
	bx := x + (w-listW)/2
	ly := top
	if full {
		ly = top + len(startLogo) + 1
	}
	for i, e := range StartEntries {
		g.put(bx, ly+i, e.Label, labelSt)
		g.putRight(bx+listW, ly+i, e.Keys, keysSt)
	}
	return true
}

// drawLogoLine pinta una fila del logo: los bloques (█) con st y los trazos de la sombra con shadow.
func drawLogoLine(g *grid, x, y int, line string, st, shadow style) {
	for _, r := range line {
		switch r {
		case ' ':
		case '█':
			g.put(x, y, string(r), st)
		default:
			g.put(x, y, string(r), shadow)
		}
		x++
	}
}
