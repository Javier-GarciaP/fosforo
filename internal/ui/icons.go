package ui

import (
	"path/filepath"
	"strings"
)

// Icon es un icono de Nerd Font y la clave de color del tema con la que se dibuja.
type Icon struct {
	Glyph string
	Color string // «teal», «blue», «dim»... (ver Theme.color)
}

// IconMode controla si el explorador muestra iconos.
type IconMode int

const (
	// IconsNerd usa glifos de una Nerd Font (requiere la fuente en la terminal).
	IconsNerd IconMode = iota
	// IconsNone no imprime ningún icono: las carpetas terminan en «/» y se distinguen por el color.
	IconsNone
)

// ResolveIcon elige el icono de una entrada: carpeta (abierta o cerrada); para archivos, primero por nombre exacto,
// luego por extensión (sin distinguir mayúsculas) y, si no, el icono de archivo por defecto.
func ResolveIcon(name string, isDir, expanded bool) Icon {
	if isDir {
		if expanded {
			return iconDefs["folder_open"]
		}
		return iconDefs["folder"]
	}
	if key, ok := iconByName[name]; ok {
		return iconDefs[key]
	}
	if key, ok := iconByName[strings.ToLower(name)]; ok {
		return iconDefs[key]
	}
	if ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), "."); ext != "" {
		if key, ok := iconByExt[ext]; ok {
			return iconDefs[key]
		}
	}
	return iconDefs["default"]
}

// color traduce una clave de color del tema («teal», «dim»...) a su valor hexadecimal.
func (t Theme) color(key string) string {
	switch key {
	case "blue":
		return t.Blue
	case "green":
		return t.Green
	case "mauve":
		return t.Mauve
	case "peach":
		return t.Peach
	case "yellow":
		return t.Yellow
	case "red":
		return t.Red
	case "teal":
		return t.Teal
	case "dim":
		return t.Dim
	}
	return t.FG
}
