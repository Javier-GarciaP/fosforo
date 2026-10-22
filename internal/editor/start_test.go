package editor_test

import (
	"path/filepath"
	"strings"
	"testing"
)

const startMarker = "Space f f" // un atajo de la lista: solo aparece si se dibuja la pantalla de inicio

func (e *ed) startVisible() bool { return strings.Contains(e.screen(), startMarker) }

// Un buffer nuevo (vacío, sin nombre y sin modificar) muestra la pantalla de inicio, con el logo y los atajos.
func TestStartScreenShownOnEmptyBuffer(t *testing.T) {
	e := newEd(t, "")
	if !e.startVisible() {
		t.Fatalf("debe verse la pantalla de inicio:\n%s", e.screen())
	}
	if !strings.Contains(e.screen(), "██╗") {
		t.Error("falta el logo")
	}
}

// Abrir un archivo la oculta; cerrar el último buffer la vuelve a mostrar; un archivo vacío con nombre no la muestra.
func TestStartScreenHiddenByFileAndBackAfterClose(t *testing.T) {
	root := project(t, "a.txt", "vacio.txt=")
	e := openProj(t, root)
	if !e.startVisible() {
		t.Fatalf("al abrir la carpeta debe verse:\n%s", e.screen())
	}
	e.ex("e " + filepath.Join(root, "a.txt"))
	if e.startVisible() {
		t.Error("con un archivo abierto no debe verse")
	}
	e.ex("bd")
	if !e.startVisible() {
		t.Errorf("al cerrar el último buffer debe volver:\n%s", e.screen())
	}
	e.ex("e " + filepath.Join(root, "vacio.txt"))
	if e.startVisible() {
		t.Error("un archivo existente pero vacío tiene nombre: no debe verse")
	}
}

// Pulsar «i» (modo Insert) o escribir texto la oculta; tras escribir sigue oculta aunque vuelvas a Normal.
func TestStartScreenHiddenWhenTyping(t *testing.T) {
	e := newEd(t, "")
	e.typ("i")
	if e.startVisible() {
		t.Error("en modo Insert no debe verse")
	}
	e.typ("x").send(keyEsc)
	if e.startVisible() {
		t.Error("con texto escrito (buffer modificado) no debe verse")
	}
}

// Con el explorador abierto el área del editor es más estrecha; sigue viéndose si cabe la lista.
func TestStartScreenNextToExplorer(t *testing.T) {
	e := openProj(t, project(t, "a.txt"))
	if !e.startVisible() {
		t.Errorf("100 columnas con explorador: debe verse la pantalla (al menos la lista):\n%s", e.screen())
	}
}

// El cursor no se muestra sobre la pantalla de inicio (ni siquiera se pinta el «~» de las filas vacías).
func TestStartScreenHasNoTildes(t *testing.T) {
	if strings.Contains(newEd(t, "").screen(), "~") {
		t.Error("con la pantalla de inicio no debe haber «~»")
	}
}

// Es solo informativa: un clic sobre una fila no cambia nada, y las teclas funcionan igual («:» abre la línea de comandos).
func TestStartScreenIsInformational(t *testing.T) {
	e := newEd(t, "")
	before := e.screen()
	for y := 0; y < 22; y++ { // un clic en cada fila del área
		e.click(40, y).release(40, y)
	}
	if e.screen() != before {
		t.Error("los clics no deben cambiar la pantalla")
	}
	e.typ(":")
	if !strings.Contains(e.screen(), ":") || e.m.Mode().Label() != "COMMAND" {
		t.Errorf("«:» debe abrir la línea de comandos, modo %v", e.m.Mode())
	}
}
