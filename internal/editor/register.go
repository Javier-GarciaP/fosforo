package editor

import (
	"context"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// register es el registro sin nombre: el último texto borrado o copiado, que pegan p y P.
type register struct {
	text     string
	linewise bool // se copió/borró por líneas: se pega en líneas nuevas
}

func (r register) empty() bool { return r.text == "" && !r.linewise }

// Clipboard copia texto al portapapeles del sistema y devuelve un comando de Bubble Tea (o nil).
type Clipboard func(text string) tea.Cmd

// defaultClipboard usa la secuencia OSC 52 (funciona por SSH y tmux) y, además, una utilidad del sistema si existe.
func defaultClipboard(text string) tea.Cmd {
	return tea.Batch(tea.SetClipboard(text), systemClipboard(text))
}

func systemClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		for _, c := range [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "--clipboard", "--input"}, {"pbcopy"}} {
			path, err := exec.LookPath(c[0])
			if err != nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			cmd := exec.CommandContext(ctx, path, c[1:]...)
			cmd.Stdin = strings.NewReader(text)
			_ = cmd.Run()
			cancel()
			return nil
		}
		return nil
	}
}
