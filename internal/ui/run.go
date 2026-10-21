// Package ui contiene lo que dibuja y arranca la interfaz de terminal.
package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Run ejecuta el modelo en la pantalla alternativa y devuelve error si la terminal no terminó limpia.
// Bubble Tea restaura la terminal (cursor, pantalla normal, modo de línea) al terminar por salida normal,
// pánico o señal; aquí solo se traduce el pánico a un error con contexto.
func Run(model tea.Model, opts ...tea.ProgramOption) error {
	_, err := tea.NewProgram(model, opts...).Run()
	if err != nil {
		return fmt.Errorf("la interfaz terminó con error: %w", err)
	}
	return nil
}
