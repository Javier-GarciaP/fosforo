// Package keymap resuelve secuencias de teclas (dd, gg, Space f f) en comandos con un árbol de prefijos por modo.
package keymap

import (
	"fmt"
	"sort"
	"strings"
)

// Status es el resultado de resolver una tecla.
type Status int

const (
	// Unmapped: la secuencia no lleva a ningún comando.
	Unmapped Status = iota
	// Pending: la secuencia es prefijo de otras más largas; hay que esperar otra tecla.
	Pending
	// Complete: la secuencia es un comando.
	Complete
)

func (s Status) String() string { return [...]string{"Unmapped", "Pending", "Complete"}[s] }

// Result es la respuesta de Resolve. Command solo tiene valor si Status es Complete.
type Result struct {
	Status  Status
	Command string
}

type node struct {
	children map[string]*node
	command  string
}

// Keymap guarda un árbol de prefijos por cada modo.
type Keymap struct {
	modes map[string]*node
}

// New crea un keymap vacío.
func New() *Keymap { return &Keymap{modes: map[string]*node{}} }

// Keys divide una secuencia escrita con espacios («g g», «space f f», «ctrl+d») en teclas.
func Keys(seq string) []string { return strings.Fields(seq) }

// Bind asocia la secuencia seq (teclas separadas por espacios) al comando en un modo.
// Falla si la secuencia está vacía o choca con otra: no puede ser prefijo de un comando ni extender un comando ya asignado.
func (k *Keymap) Bind(mode, seq, command string) error {
	keys := Keys(seq)
	if len(keys) == 0 {
		return fmt.Errorf("secuencia vacía para %s", command)
	}
	root := k.modes[mode]
	if root == nil {
		root = &node{}
		k.modes[mode] = root
	}
	n := root
	for i, key := range keys {
		if n.command != "" {
			return fmt.Errorf("%q en modo %s: %q ya es el comando %s", seq, mode, strings.Join(keys[:i], " "), n.command)
		}
		if n.children == nil {
			n.children = map[string]*node{}
		}
		child := n.children[key]
		if child == nil {
			child = &node{}
			n.children[key] = child
		}
		n = child
	}
	switch {
	case n.command != "":
		return fmt.Errorf("%q en modo %s ya está asignada a %s", seq, mode, n.command)
	case len(n.children) > 0:
		return fmt.Errorf("%q en modo %s es prefijo de otras secuencias", seq, mode)
	}
	n.command = command
	return nil
}

// MustBind es Bind y entra en pánico si falla (para las tablas por defecto).
func (k *Keymap) MustBind(mode, seq, command string) {
	if err := k.Bind(mode, seq, command); err != nil {
		panic(err)
	}
}

// Resolve resuelve key después de las teclas pendientes en un modo.
func (k *Keymap) Resolve(mode string, pending []string, key string) Result {
	n := k.modes[mode]
	if n == nil {
		return Result{}
	}
	for _, p := range pending {
		if n = n.children[p]; n == nil {
			return Result{}
		}
	}
	if n = n.children[key]; n == nil {
		return Result{}
	}
	if n.command != "" {
		return Result{Status: Complete, Command: n.command}
	}
	return Result{Status: Pending}
}

// HasPrefix indica si alguna secuencia del modo empieza por las teclas dadas.
func (k *Keymap) HasPrefix(mode string, keys []string) bool {
	n := k.modes[mode]
	for _, key := range keys {
		if n == nil {
			return false
		}
		n = n.children[key]
	}
	return n != nil
}

// Walk recorre todas las secuencias asignadas, en orden estable (por modo y secuencia), para listarlas en la ayuda o la paleta.
func (k *Keymap) Walk(fn func(mode, seq, command string)) {
	modes := make([]string, 0, len(k.modes))
	for m := range k.modes {
		modes = append(modes, m)
	}
	sort.Strings(modes)
	for _, mode := range modes {
		var walk func(n *node, prefix []string)
		walk = func(n *node, prefix []string) {
			if n.command != "" {
				fn(mode, strings.Join(prefix, " "), n.command)
				return
			}
			keys := make([]string, 0, len(n.children))
			for key := range n.children {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				walk(n.children[key], append(prefix[:len(prefix):len(prefix)], key))
			}
		}
		walk(k.modes[mode], nil)
	}
}
