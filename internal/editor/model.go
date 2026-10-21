// Package editor contiene el estado y la lógica del editor.
package editor

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/version"
)

// Tamaño mínimo de terminal en el que se dibuja la interfaz.
const (
	MinWidth  = 40
	MinHeight = 10
)

// Model es el modelo de Bubble Tea del editor. Por ahora solo muestra un marcador.
type Model struct {
	Path          string // ruta pedida en la línea de comandos (la usa core-text-editing)
	width, height int
}

// New crea el modelo inicial.
func New(path string) Model { return Model{Path: path} }

// Init no lanza comandos iniciales.
func (m Model) Init() tea.Cmd { return nil }

// Update procesa mensajes: redimensionado y salida con q o Ctrl+C.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// TooSmall indica si la terminal es menor que el mínimo. Mientras no se conoce el tamaño (0x0) no se avisa.
func (m Model) TooSmall() bool {
	if m.width == 0 && m.height == 0 {
		return false
	}
	return m.width < MinWidth || m.height < MinHeight
}

// View dibuja el marcador o el aviso de tamaño mínimo.
func (m Model) View() tea.View {
	var s string
	if m.TooSmall() {
		s = fmt.Sprintf("Terminal demasiado pequeña (%dx%d).\nMínimo: %dx%d.", m.width, m.height, MinWidth, MinHeight)
	} else {
		s = fmt.Sprintf("%s\n\nPulsa q para salir.", version.String())
		if m.Path != "" {
			s += "\nRuta: " + m.Path
		}
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
