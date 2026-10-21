package ui_test

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/termtest"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

type panicModel struct{}

func (panicModel) Init() tea.Cmd { return nil }
func (panicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "p" {
		panic("pánico de prueba")
	}
	return panicModel{}, nil
}
func (panicModel) View() tea.View {
	v := tea.NewView("modelo de prueba")
	v.AltScreen = true
	return v
}

func TestRunRestoresTerminalAfterPanic(t *testing.T) {
	if raceEnabled {
		// Bubble Tea (cancelreader) cierra el archivo de entrada mientras otra goroutine aún lo lee al recuperarse de un
		// pánico: el detector lo marca aunque el comportamiento es correcto. `make test` la ejecuta sin -race.
		t.Skip("carrera conocida en Bubble Tea al cerrar la entrada tras un pánico; se prueba sin -race")
	}
	pty := termtest.New(t)
	before := pty.State(t)

	errc := make(chan error, 1)
	go func() {
		errc <- ui.Run(panicModel{}, tea.WithInput(pty.Slave), tea.WithOutput(pty.Slave))
	}()

	pty.WaitFor(t, termtest.EnterAltScreen, 3*time.Second)
	if *pty.State(t) == *before {
		t.Fatal("durante la ejecución la terminal debía estar en modo raw")
	}
	pty.Write("p")

	select {
	case err := <-errc:
		if err == nil {
			t.Fatal("Run debía devolver un error tras el pánico")
		}
		if !errors.Is(err, tea.ErrProgramPanic) {
			t.Errorf("el error no es ErrProgramPanic: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run no terminó tras el pánico")
	}

	// la salida llega al lector del pseudo-terminal de forma asíncrona: se espera a que aparezca
	pty.WaitFor(t, termtest.ShowCursor, 3*time.Second)
	pty.WaitFor(t, termtest.ExitAltScreen, 3*time.Second)
	if after := pty.State(t); *after != *before {
		t.Errorf("el modo de la terminal no se restauró:\nantes  %+v\ndespués %+v", before, after)
	}
}
