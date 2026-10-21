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
	if Version == "dev" { // sin -ldflags: con «go install ...@vX.Y.Z» la versión viene de los metadatos de compilación
		if v := moduleVersion(); v != "" {
			return Name + " " + v
		}
	}
	if commit == "" {
		commit = vcsRevision()
	}
	if commit == "" {
		return Name + " " + Version
	}
	return Name + " " + Version + " (" + commit + ")"
}

// moduleVersion devuelve la versión del módulo si el binario se instaló con «go install» desde una etiqueta (sin la «v»).
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" || strings.Contains(info.Main.Version, "+dirty") {
		return ""
	}
	return strings.TrimPrefix(info.Main.Version, "v")
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
