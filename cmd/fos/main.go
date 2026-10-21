// Command fos es el editor de código para la terminal.
package main

import (
	"os"

	"github.com/Javier-GarciaP/fosforo/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.RunEditor))
}
