// Package cli interpreta la línea de comandos y arranca la interfaz.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
	"github.com/Javier-GarciaP/fosforo/internal/version"
)

// Códigos de salida.
const (
	ExitOK          = 0
	ExitError       = 1
	ExitUsage       = 2
	ExitInterrupted = 130
)

// Options es el resultado de interpretar los argumentos.
type Options struct {
	ShowVersion bool
	ShowHelp    bool
	Path        string // ruta opcional para abrir; la usa core-text-editing
}

const usageText = `Uso: %s [opciones] [ruta]

Editor de código para la terminal, con teclado modal.

Opciones:
`

func newFlagSet(opts *Options, out io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(version.Name, flag.ContinueOnError)
	fs.SetOutput(out)
	fs.BoolVar(&opts.ShowVersion, "version", false, "muestra la versión y termina")
	fs.BoolVar(&opts.ShowHelp, "help", false, "muestra esta ayuda y termina")
	return fs
}

func writeUsage(w io.Writer) {
	fmt.Fprintf(w, usageText, version.Name)
	newFlagSet(&Options{}, w).PrintDefaults()
}

// Parse interpreta args (sin el nombre del programa). No escribe nada: quien llama muestra los errores.
func Parse(args []string) (Options, error) {
	var opts Options
	fs := newFlagSet(&opts, io.Discard)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			opts.ShowHelp = true
			return opts, nil
		}
		return opts, err
	}
	if fs.NArg() > 1 {
		return opts, fmt.Errorf("demasiados argumentos: se esperaba una sola ruta, no %d", fs.NArg())
	}
	opts.Path = fs.Arg(0)
	return opts, nil
}

// Run ejecuta el programa y devuelve el código de salida. runUI abre la interfaz con la ruta dada.
func Run(args []string, stdout, stderr io.Writer, runUI func(path string) error) int {
	opts, err := Parse(args)
	switch {
	case err != nil:
		fmt.Fprintf(stderr, "%s: %v\n\n", version.Name, err)
		writeUsage(stderr)
		return ExitUsage
	case opts.ShowVersion:
		fmt.Fprintln(stdout, version.String())
		return ExitOK
	case opts.ShowHelp:
		writeUsage(stdout)
		return ExitOK
	}
	if err := runUI(opts.Path); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return ExitInterrupted
		}
		fmt.Fprintf(stderr, "%s: %v\n", version.Name, err)
		return ExitError
	}
	return ExitOK
}

// RunEditor abre el editor en la terminal real.
func RunEditor(path string) error {
	return ui.Run(editor.New(path))
}
