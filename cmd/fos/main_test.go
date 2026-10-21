package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Javier-GarciaP/fosforo/internal/termtest"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "fos-e2e")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(dir, "fos")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// start lanza el binario con el pseudo-terminal como terminal de control.
func start(t *testing.T, pty *termtest.PTY, args ...string) (*exec.Cmd, chan error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = pty.Slave, pty.Slave, pty.Slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	return cmd, done
}

func exitCode(t *testing.T, done chan error) int {
	t.Helper()
	select {
	case err := <-done:
		if err == nil {
			return 0
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("el programa no terminó a tiempo")
	}
	return -1
}

func checkRestored(t *testing.T, pty *termtest.PTY) {
	t.Helper()
	out := pty.Output()
	for _, want := range []string{termtest.EnterAltScreen, termtest.ShowCursor, termtest.ExitAltScreen} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en la salida", want)
		}
	}
	if strings.LastIndex(out, termtest.ExitAltScreen) < strings.LastIndex(out, termtest.EnterAltScreen) {
		t.Error("la pantalla alternativa no se cerró después de abrirse")
	}
}

func TestExitKeys(t *testing.T) {
	for _, key := range []string{"q", "\x03"} {
		t.Run(strings.ReplaceAll(key, "\x03", "ctrl+c"), func(t *testing.T) {
			pty := termtest.New(t)
			before := pty.State(t)
			_, done := start(t, pty)
			pty.WaitFor(t, "Pulsa q para salir", 5*time.Second)
			pty.Write(key)
			if code := exitCode(t, done); code != 0 {
				t.Errorf("código de salida = %d, quería 0", code)
			}
			checkRestored(t, pty)
			if after := pty.State(t); *after != *before {
				t.Error("el modo de la terminal no se restauró")
			}
		})
	}
}

func TestSignalsRestoreTerminal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		t.Run(sig.String(), func(t *testing.T) {
			pty := termtest.New(t)
			before := pty.State(t)
			cmd, done := start(t, pty)
			pty.WaitFor(t, "Pulsa q para salir", 5*time.Second)
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			exitCode(t, done)
			checkRestored(t, pty)
			if after := pty.State(t); *after != *before {
				t.Error("el modo de la terminal no se restauró tras la señal")
			}
		})
	}
}

func TestVersionDoesNotUseAltScreen(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty, "--version")
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código = %d", code)
	}
	time.Sleep(50 * time.Millisecond)
	out := pty.Output()
	if strings.Contains(out, termtest.EnterAltScreen) {
		t.Error("--version no debe abrir la pantalla alternativa")
	}
	if !strings.HasPrefix(out, "fos ") {
		t.Errorf("salida inesperada: %q", out)
	}
}

func TestUnknownFlagExitsWithTwo(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty, "--no-existe")
	if code := exitCode(t, done); code != 2 {
		t.Fatalf("código = %d, quería 2", code)
	}
}
