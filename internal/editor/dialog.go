package editor

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

type dialogButton struct{ label, key string }

// dialog es un cuadro modal con botones. El último botón siempre es «Cancelar» y lo activa Esc.
// choose recibe el índice del botón elegido; el diálogo ya está cerrado cuando se llama.
type dialog struct {
	title    string
	lines    []string
	buttons  []dialogButton
	selected int
	choose   func(m *Model, idx int)
}

var (
	btnSave     = dialogButton{"Guardar (y)", "y"}
	btnDiscard  = dialogButton{"Descartar (n)", "n"}
	btnSaveAll  = dialogButton{"Guardar todo (y)", "y"}
	btnDiscAll  = dialogButton{"Descartar todo (n)", "n"}
	btnCancel   = dialogButton{"Cancelar (Esc)", "esc"}
	btnDelete   = dialogButton{"Borrar (y)", "y"}
	btnOverride = dialogButton{"Sobrescribir (y)", "y"}
)

func (m *Model) dialogKey(k tea.KeyPressMsg) tea.Cmd {
	d := m.dialog
	n := len(d.buttons)
	choice := -1
	switch key := k.String(); key {
	case "left", "shift+tab":
		d.selected = (d.selected + n - 1) % n
	case "right", "tab":
		d.selected = (d.selected + 1) % n
	case "enter":
		choice = d.selected
	case "esc", "ctrl+c":
		choice = n - 1
	default:
		for i, b := range d.buttons {
			if b.key == key {
				choice = i
			}
		}
	}
	if choice >= 0 {
		m.chooseDialog(choice)
	}
	return nil
}

// chooseDialog cierra el diálogo y ejecuta el botón elegido (con el teclado o con el ratón).
func (m *Model) chooseDialog(choice int) {
	d := m.dialog
	if d == nil || choice < 0 || choice >= len(d.buttons) {
		return
	}
	m.dialog = nil
	m.msg, m.msgErr = "", false
	d.choose(m, choice)
}

// view convierte el diálogo en lo que dibuja la interfaz.
func (d *dialog) view() *ui.Dialog {
	labels := make([]string, len(d.buttons))
	for i, b := range d.buttons {
		labels[i] = b.label
	}
	return &ui.Dialog{Title: d.title, Lines: d.lines, Buttons: labels, Selected: d.selected, Hint: "←→ elegir   ⏎ confirmar"}
}

// unsavedDialog pregunta qué hacer con los buffers modificados. onSave guarda (y sigue), onDiscard descarta.
func unsavedDialog(names []string, onSave, onDiscard func(m *Model)) *dialog {
	d := &dialog{title: "Cambios sin guardar"}
	if len(names) == 1 {
		d.lines = []string{fmt.Sprintf("%q tiene cambios sin guardar.", names[0]), "¿Qué quieres hacer?"}
		d.buttons = []dialogButton{btnSave, btnDiscard, btnCancel}
	} else {
		d.lines = []string{fmt.Sprintf("%d archivos tienen cambios sin guardar: %s.", len(names), joinNames(names, 3)), "¿Qué quieres hacer?"}
		d.buttons = []dialogButton{btnSaveAll, btnDiscAll, btnCancel}
	}
	d.choose = func(m *Model, i int) {
		switch i {
		case 0:
			onSave(m)
		case 1:
			onDiscard(m)
		}
	}
	return d
}

// joinNames une nombres con comas mostrando como mucho max y «y N más».
func joinNames(names []string, max int) string {
	if len(names) <= max {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s y %d más", strings.Join(names[:max], ", "), len(names)-max)
}
