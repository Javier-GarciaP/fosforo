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
	Path        string // archivo o carpeta opcional para abrir
	Icons       string // «nerd» (por defecto) o «none»: iconos del explorador
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
	fs.StringVar(&opts.Icons, "icons", "nerd", "iconos del explorador: nerd (necesita una Nerd Font) o none")
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
	if opts.Icons != "nerd" && opts.Icons != "none" {
		return opts, fmt.Errorf("valor no válido para -icons: %q (usa nerd o none)", opts.Icons)
	}
	opts.Path = fs.Arg(0)
	return opts, nil
}

// Run ejecuta el programa y devuelve el código de salida. runUI abre la interfaz con la ruta dada.
func Run(args []string, stdout, stderr io.Writer, runUI func(opts Options) error) int {
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
	if err := runUI(opts); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return ExitInterrupted
		}
		fmt.Fprintf(stderr, "%s: %v\n", version.Name, err)
		return ExitError
	}
	return ExitOK
}

// RunEditor abre el editor en la terminal real.
func RunEditor(opts Options) error {
	icons := ui.IconsNerd
	if opts.Icons == "none" {
		icons = ui.IconsNone
	}
	m, err := editor.Open(opts.Path, editor.WithIcons(icons))
	if err != nil {
		return err
	}
	return ui.Run(m)
}
