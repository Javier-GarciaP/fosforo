package buffer

// group es una unidad de deshacer: las ediciones hechas entre BeginGroup y EndGroup (o una sola edición suelta).
type group struct {
	edits  []Edit
	cursor Pos // dónde estaba el cursor al empezar el grupo; se restaura al deshacer
}

type history struct {
	undo, redo []*group
	open       *group // grupo que está recibiendo ediciones
	grouping   bool
	cursor     Pos // posición del cursor al abrir el grupo
}

// BeginGroup agrupa todas las ediciones siguientes en un solo paso de deshacer (p. ej. una sesión de Insert).
// cursor es la posición que se restaurará al deshacerlo. Se cierra con EndGroup. Si no llegan ediciones, no crea ningún paso.
func (b *Buffer) BeginGroup(cursor Pos) {
	b.hist.grouping, b.hist.open, b.hist.cursor = true, nil, cursor
}

// EndGroup cierra el grupo abierto: la siguiente edición empezará un paso de deshacer nuevo.
func (b *Buffer) EndGroup() {
	b.hist.grouping, b.hist.open = false, nil
}

// record anota una edición ya aplicada.
func (b *Buffer) record(e Edit) {
	h := &b.hist
	if len(h.redo) > 0 {
		// Una edición nueva descarta el futuro; si el punto guardado estaba en él, ya no se puede volver a él.
		if b.savedVersion > b.version {
			b.savedVersion = -1
		}
		h.redo = nil
	}
	if h.open == nil {
		cur := e.Start
		if h.grouping {
			cur = h.cursor
		}
		h.open = &group{cursor: cur}
		h.undo = append(h.undo, h.open)
	}
	h.open.edits = append(h.open.edits, e)
	b.version++
	if !h.grouping {
		h.open = nil
	}
}

// CanUndo y CanRedo indican si hay pasos que deshacer o rehacer.
func (b *Buffer) CanUndo() bool { return len(b.hist.undo) > 0 }
func (b *Buffer) CanRedo() bool { return len(b.hist.redo) > 0 }

// Undo deshace el último paso y devuelve la posición donde estaba el cursor antes de él. false si no hay nada que deshacer.
func (b *Buffer) Undo() (Pos, bool) {
	h := &b.hist
	if len(h.undo) == 0 {
		return Pos{}, false
	}
	b.EndGroup()
	g := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	for i := len(g.edits) - 1; i >= 0; i-- {
		if _, err := b.apply(g.edits[i].Invert()); err != nil {
			// no debería ocurrir: el historial es coherente con el texto. Se descarta para no corromper nada.
			h.undo, h.redo = nil, nil
			return b.Clamp(g.cursor), false
		}
	}
	h.redo = append(h.redo, g)
	b.version -= len(g.edits)
	return b.Clamp(g.cursor), true
}

// Redo rehace el último paso deshecho y devuelve la posición del inicio de su primera edición.
func (b *Buffer) Redo() (Pos, bool) {
	h := &b.hist
	if len(h.redo) == 0 {
		return Pos{}, false
	}
	b.EndGroup()
	g := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	for _, e := range g.edits {
		if _, err := b.apply(e); err != nil {
			h.undo, h.redo = nil, nil
			return Pos{}, false
		}
	}
	h.undo = append(h.undo, g)
	b.version += len(g.edits)
	return b.Clamp(g.edits[0].Start), true
}
