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
	for _, want := range []string{termtest.EnterAltScreen, termtest.ShowCursor, termtest.ExitAltScreen} {
		pty.WaitFor(t, want, 3*time.Second) // la salida llega de forma asíncrona
	}
	out := pty.Output()
	if strings.LastIndex(out, termtest.ExitAltScreen) < strings.LastIndex(out, termtest.EnterAltScreen) {
		t.Error("la pantalla alternativa no se cerró después de abrirse")
	}
}

func TestExitWithQuitCommand(t *testing.T) {
	pty := termtest.New(t)
	before := pty.State(t)
	_, done := start(t, pty)
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write(":q\r")
	if code := exitCode(t, done); code != 0 {
		t.Errorf("código de salida = %d, quería 0", code)
	}
	checkRestored(t, pty)
	if after := pty.State(t); *after != *before {
		t.Error("el modo de la terminal no se restauró")
	}
}

func TestCtrlQAndCtrlCDoNotExit(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty)
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write("\x11\x03") // Ctrl+Q ya no está asignada y Ctrl+C solo cancela
	time.Sleep(300 * time.Millisecond)
	select {
	case err := <-done:
		t.Fatalf("el programa no debía terminar: %v", err)
	default:
	}
	pty.Write(":q\r")
	if code := exitCode(t, done); code != 0 {
		t.Errorf("código de salida = %d", code)
	}
}

func TestSignalsRestoreTerminal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		t.Run(sig.String(), func(t *testing.T) {
			pty := termtest.New(t)
			before := pty.State(t)
			cmd, done := start(t, pty)
			pty.WaitFor(t, "NORMAL", 5*time.Second)
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

func TestOpenEditAndSaveRealFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "notas.txt")
	if err := os.WriteFile(p, []byte("hola\nmundo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pty := termtest.New(t)
	before := pty.State(t)
	_, done := start(t, pty, p)

	pty.WaitFor(t, "mundo", 5*time.Second) // el contenido del archivo aparece en pantalla
	pty.Write("iX")                        // Insert y escribe al inicio de «hola»
	pty.WaitFor(t, "[+]", 5*time.Second)
	pty.Write("\x13") // Ctrl+S
	pty.WaitFor(t, "escritas", 5*time.Second)
	pty.Write("\x03") // Ctrl+C: vuelve a Normal
	pty.Write(":q\r") // sin cambios pendientes sale directamente
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código de salida = %d", code)
	}
	if got, _ := os.ReadFile(p); string(got) != "Xhola\nmundo\n" {
		t.Errorf("archivo = %q", got)
	}
	checkRestored(t, pty)
	if after := pty.State(t); *after != *before {
		t.Error("el modo de la terminal no se restauró")
	}
}

func TestQuitAsksBeforeDiscarding(t *testing.T) {
	p := filepath.Join(t.TempDir(), "n.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	pty := termtest.New(t)
	_, done := start(t, pty, p)
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write("iz")
	pty.Write("\x03")
	pty.Write(":q\r") // :q con cambios: aparece el diálogo
	pty.WaitFor(t, "Cambios sin guardar", 5*time.Second)
	pty.Write("n") // descartar
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código de salida = %d", code)
	}
	if got, _ := os.ReadFile(p); string(got) != "a\n" {
		t.Errorf("descartar no debe tocar el archivo: %q", got)
	}
}

func TestOpenMissingPathCreatesOnSave(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nuevo.txt")
	pty := termtest.New(t)
	_, done := start(t, pty, p)
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write("ihola")
	pty.Write("\x13")
	pty.WaitFor(t, "escritas", 5*time.Second)
	pty.Write("\x03")
	pty.Write(":q\r")
	exitCode(t, done)
	if got, _ := os.ReadFile(p); string(got) != "hola\n" {
		t.Errorf("archivo nuevo = %q", got)
	}
}

func TestOpenDirectoryExitsWithError(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty, t.TempDir())
	if code := exitCode(t, done); code != 1 {
		t.Fatalf("código = %d, quería 1", code)
	}
	pty.WaitFor(t, "directorio", 3*time.Second)
	if strings.Contains(pty.Output(), termtest.EnterAltScreen) {
		t.Error("no debe abrirse la pantalla alternativa si el archivo no se puede abrir")
	}
}
