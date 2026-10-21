// Package version guarda el nombre del comando y la versión del programa.
package version

import (
	"runtime/debug"
	"strings"
)

// Name es el nombre del comando ejecutable. Es la única mención al nombre en el código.
const Name = "fos"

// Version y Commit se rellenan al compilar una release con -ldflags "-X ...".
var (
	Version = "dev"
	Commit  = ""
)

// String devuelve la versión en una línea, p. ej. "fos 0.8.0 (abc1234)".
func String() string {
	commit := Commit
	if commit == "" {
		commit = vcsRevision()
	}
	if commit == "" {
		return Name + " " + Version
	}
	return Name + " " + Version + " (" + commit + ")"
}

func vcsRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			return strings.TrimSpace(s.Value[:min(7, len(s.Value))])
		}
	}
	return ""
}
