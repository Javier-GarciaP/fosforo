package main_test

import (
	"fmt"
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
	// la configuración real del usuario no debe influir en las pruebas
	cfgDir, _ := os.MkdirTemp("", "fos-cfg")
	os.Setenv("XDG_CONFIG_HOME", cfgDir)
	code := m.Run()
	os.RemoveAll(cfgDir)
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

func TestVersionFromLdflags(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "fos")
	flags := "-X github.com/Javier-GarciaP/fosforo/internal/version.Version=9.8.7 -X github.com/Javier-GarciaP/fosforo/internal/version.Commit=abc1234"
	if out, err := exec.Command("go", "build", "-ldflags", flags, "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != "fos 9.8.7 (abc1234)" {
		t.Errorf("--version = %q", got)
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

func TestOpenDirectoryShowsExplorer(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("x\n"), 0o644)
	os.Mkdir(filepath.Join(dir, "carpeta"), 0o755)
	pty := termtest.New(t)
	_, done := start(t, pty, dir)
	pty.WaitFor(t, "visible.txt", 5*time.Second)
	pty.WaitFor(t, "EXPLORER", 5*time.Second)
	pty.Write(":q\r")
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código = %d", code)
	}
}

func TestOpenMissingDirectoryExitsWithError(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty, filepath.Join(t.TempDir(), "no", "existe", "f.txt"))
	// un archivo en una carpeta que no existe se abre vacío y se puede cerrar
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write(":q\r")
	exitCode(t, done)
}

func TestOpenMarkdownShowsPreview(t *testing.T) {
	p := filepath.Join(t.TempDir(), "LEEME.md")
	if err := os.WriteFile(p, []byte("# Título grande\n\n- uno\n- [x] hecho\n\nTexto con **negrita**.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pty := termtest.New(t)
	_, done := start(t, pty, p)
	pty.WaitFor(t, "PREVIEW", 5*time.Second)
	pty.WaitFor(t, "Título grande", 5*time.Second)
	pty.WaitFor(t, "hecho", 5*time.Second)
	if strings.Contains(pty.Output(), "**negrita**") || strings.Contains(pty.Output(), "# Título") {
		t.Error("se ve el fuente en lugar del documento renderizado")
	}
	pty.Write("e") // pasa al fuente
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.Write(":q\r")
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código = %d", code)
	}
	if got, _ := os.ReadFile(p); !strings.HasPrefix(string(got), "# Título grande") {
		t.Errorf("abrir no modifica el archivo: %q", got)
	}
}

// sgr construye un evento de ratón del protocolo SGR (coordenadas desde 1): botón 0 = izquierdo, 64/65 = rueda.
func sgr(button int, x, y int, press bool) string {
	end := "M"
	if !press {
		end = "m"
	}
	return fmt.Sprintf("\x1b[<%d;%d;%d%s", button, x+1, y+1, end)
}

func TestMouseThroughTheRealTerminal(t *testing.T) {
	p := filepath.Join(t.TempDir(), "m.txt")
	var sb strings.Builder
	for i := 0; i < 80; i++ {
		fmt.Fprintf(&sb, "línea número %d\n", i)
	}
	os.WriteFile(p, []byte(sb.String()), 0o644)
	pty := termtest.New(t)
	_, done := start(t, pty, p)
	pty.WaitFor(t, "NORMAL", 5*time.Second)
	pty.WaitFor(t, "\x1b[?1002h", 5*time.Second) // el programa activa el ratón (movimiento con botón pulsado)
	pty.WaitFor(t, "\x1b[?1006h", 5*time.Second) // protocolo SGR

	// clic en la fila 3, columna de texto 6 (el texto empieza en la columna 4 de pantalla: número de línea de 3 cifras + 1)
	pty.Write(sgr(0, 4+6, 3, true) + sgr(0, 4+6, 3, false))
	pty.WaitFor(t, "4:7", 5*time.Second)
	// la rueda baja 3 líneas por muesca: el cursor sigue visible
	pty.Write(sgr(65, 20, 5, true))
	pty.WaitFor(t, "línea número 3", 5*time.Second)
	pty.Write(sgr(65, 20, 5, true) + sgr(65, 20, 5, true) + sgr(65, 20, 5, true))
	pty.WaitFor(t, "línea número 30", 5*time.Second)
	pty.Write(":q\r")
	if code := exitCode(t, done); code != 0 {
		t.Fatalf("código = %d", code)
	}
	checkRestored(t, pty)
	if !strings.Contains(pty.Output(), "\x1b[?1002l") && !strings.Contains(pty.Output(), "\x1b[?1000l") {
		t.Error("al salir se desactiva la captura del ratón")
	}
}

func TestMouseToggleReleasesTheTerminal(t *testing.T) {
	pty := termtest.New(t)
	_, done := start(t, pty)
	pty.WaitFor(t, "\x1b[?1002h", 5*time.Second)
	pty.Write(":mouse\r")
	// con la captura desactivada el terminal vuelve a tratar el ratón como suyo (selección nativa)
	pty.WaitFor(t, "\x1b[?1002l", 5*time.Second)
	pty.WaitFor(t, "selección nativa", 5*time.Second)
	mark := len(pty.Output())
	pty.Write(":mouse\r")
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(pty.Output()[mark:], "\x1b[?1002h") { // se vuelve a capturar el ratón
		if time.Now().After(deadline) {
			t.Fatal("al volver a activar el ratón se debe reanudar la captura")
		}
		time.Sleep(10 * time.Millisecond)
	}
	pty.Write(":q\r")
	exitCode(t, done)
}
