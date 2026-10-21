// Package command es el registro central de comandos con nombre: el único punto por el que cualquier acción del editor
// se invoca, sea desde una tecla, la línea de comandos «:», la paleta o el ratón.
package command

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	tea "charm.land/bubbletea/v2"
)

// Args son los argumentos con los que se ejecuta un comando.
type Args struct {
	Count    int    // repeticiones o número de línea; 0 si no se indicó
	HasCount bool   // se escribió un conteo explícito
	Bang     bool   // se usó «!» (p. ej. :q!)
	Text     string // argumento de texto (p. ej. la ruta de :e)
}

// N devuelve el conteo, o 1 si no se indicó ninguno.
func (a Args) N() int {
	if a.Count < 1 {
		return 1
	}
	return a.Count
}

// Command es una acción con nombre. S es el estado sobre el que actúa (el modelo del editor).
// Run puede devolver un tea.Cmd (salir, copiar al portapapeles...) o nil.
type Command[S any] struct {
	Name string   // «área.verbo», único
	Desc string   // descripción en español
	Keys []string // atajos por defecto, solo informativos (p. ej. «gg», «Ctrl+S»)
	Run  func(s S, a Args) tea.Cmd
}

// ErrUnknown se devuelve al invocar un comando que no existe.
var ErrUnknown = errors.New("comando desconocido")

// UnknownError indica qué nombre no se encontró.
type UnknownError struct{ Name string }

func (e UnknownError) Error() string { return "no es un comando del editor: " + e.Name }
func (e UnknownError) Unwrap() error { return ErrUnknown }

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// Registry guarda los comandos por nombre.
type Registry[S any] struct {
	cmds map[string]Command[S]
}

// NewRegistry crea un registro vacío.
func NewRegistry[S any]() *Registry[S] { return &Registry[S]{cmds: map[string]Command[S]{}} }

// Register añade un comando. El nombre debe ser único y tener la forma «área.verbo»; la descripción es obligatoria.
func (r *Registry[S]) Register(c Command[S]) error {
	switch {
	case !namePattern.MatchString(c.Name):
		return fmt.Errorf("nombre de comando inválido %q: debe ser «área.verbo»", c.Name)
	case c.Desc == "":
		return fmt.Errorf("el comando %s necesita una descripción", c.Name)
	case c.Run == nil:
		return fmt.Errorf("el comando %s no tiene función", c.Name)
	}
	if _, dup := r.cmds[c.Name]; dup {
		return fmt.Errorf("el comando %s ya está registrado", c.Name)
	}
	r.cmds[c.Name] = c
	return nil
}

// MustRegister es Register y entra en pánico si falla (para tablas fijas de comandos).
func (r *Registry[S]) MustRegister(c Command[S]) {
	if err := r.Register(c); err != nil {
		panic(err)
	}
}

// Get busca un comando por nombre.
func (r *Registry[S]) Get(name string) (Command[S], bool) {
	c, ok := r.cmds[name]
	return c, ok
}

// List devuelve todos los comandos ordenados por nombre, con descripción y atajos.
func (r *Registry[S]) List() []Command[S] {
	out := make([]Command[S], 0, len(r.cmds))
	for _, c := range r.cmds {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Exec ejecuta el comando name. Si no existe devuelve UnknownError y no toca el estado.
func (r *Registry[S]) Exec(s S, name string, a Args) (tea.Cmd, error) {
	c, ok := r.cmds[name]
	if !ok {
		return nil, UnknownError{Name: name}
	}
	return c.Run(s, a), nil
}
