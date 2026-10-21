// Package termtest ayuda a probar programas de terminal con un pseudo-terminal real.
package termtest

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/creack/pty"
)

// PTY es un pseudo-terminal: Slave es lo que ve el programa y Master lo que ve la prueba.
type PTY struct {
	Master, Slave *os.File
	mu            sync.Mutex
	out           bytes.Buffer
	done          chan struct{}
}

// New abre un pseudo-terminal de 80x24 y empieza a recoger lo que el programa escribe.
func New(t *testing.T) *PTY {
	t.Helper()
	master, slave, err := pty.Open()
	if err != nil {
		t.Skipf("no hay pseudo-terminal disponible: %v", err)
	}
	if err := pty.Setsize(master, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	p := &PTY{Master: master, Slave: slave, done: make(chan struct{})}
	go func() {
		defer close(p.done)
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			p.mu.Lock()
			p.out.Write(buf[:n])
			p.mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() { master.Close(); slave.Close() })
	return p
}

// Output devuelve todo lo escrito hasta ahora.
func (p *PTY) Output() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.out.String()
}

// WaitFor espera hasta que la salida contenga s.
func (p *PTY) WaitFor(t *testing.T, s string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if bytes.Contains([]byte(p.Output()), []byte(s)) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no apareció %q en la salida: %q", s, p.Output())
}

// Write envía teclas al programa.
func (p *PTY) Write(s string) { _, _ = io.WriteString(p.Master, s) }

// State devuelve el modo actual de la terminal (eco, modo de línea...).
func (p *PTY) State(t *testing.T) *term.State {
	t.Helper()
	st, err := term.GetState(p.Slave.Fd())
	if err != nil {
		t.Fatal(err)
	}
	return st
}

// Escapes de terminal que Bubble Tea emite al volver al estado normal.
const (
	ShowCursor     = "\x1b[?25h"
	ExitAltScreen  = "\x1b[?1049l"
	EnterAltScreen = "\x1b[?1049h"
)
