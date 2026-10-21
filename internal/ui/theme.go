package ui

import "github.com/Javier-GarciaP/fosforo/internal/syntax"

// Theme son los colores de la interfaz ("Tinta", ver docs/design/README.md), en hexadecimal.
type Theme struct {
	BG, Panel, Sel, Visual, Border string // fondos y bordes
	FG, Dim                        string // texto principal y atenuado
	Blue, Green, Mauve             string
	Peach, Yellow, Red             string
	Teal, Brand                    string
}

// Dark es el tema oscuro de la paleta Tinta (docs/design/src/palette.js).
func Dark() Theme {
	return Theme{
		BG: "#171b27", Panel: "#0f131f", Sel: "#232c42", Visual: "#323d5c", Border: "#424a5e",
		FG: "#cfd7eb", Dim: "#75809c",
		Blue: "#7fc1fe", Green: "#7fd096", Mauve: "#cca8f6",
		Peach: "#f5a574", Yellow: "#d5b85e", Red: "#fc9c9f",
		Teal: "#4ad1d3", Brand: "#69f19a",
	}
}

// Light es el tema claro de la paleta Tinta (docs/design/src/palette.js).
func Light() Theme {
	return Theme{
		BG: "#f4f7fc", Panel: "#e9edf5", Sel: "#d9e0ef", Visual: "#c1cae2", Border: "#a9b1c5",
		FG: "#252d42", Dim: "#5f6881",
		Blue: "#006cb4", Green: "#037e3f", Mauve: "#7d4dad",
		Peach: "#a24e04", Yellow: "#7f6601", Red: "#ae3b47",
		Teal: "#037879", Brand: "#067e42",
	}
}

// syntaxColor traduce una categoría de token a un color del tema; Plain y lo desconocido usan el texto normal.
// Es el único lugar donde se decide la paleta del código (ver la guía de estilo, maqueta 00).
func (t Theme) syntaxColor(k syntax.Kind) string {
	switch k {
	case syntax.Keyword:
		return t.Mauve
	case syntax.Type:
		return t.Yellow
	case syntax.Function:
		return t.Blue
	case syntax.Builtin, syntax.Escape:
		return t.Teal
	case syntax.String:
		return t.Green
	case syntax.Number, syntax.Constant:
		return t.Peach
	case syntax.Comment:
		return t.Dim
	case syntax.Tag:
		return t.Blue
	case syntax.Attribute:
		return t.Yellow
	case syntax.Error:
		return t.Red
	}
	return t.FG
}

// modeColor devuelve el color de la etiqueta de cada modo.
func (t Theme) modeColor(mode string) string {
	switch mode {
	case "INSERT":
		return t.Green
	case "VISUAL", "V-LINE":
		return t.Mauve
	case "COMMAND":
		return t.Peach
	case "PREVIEW", "EXPLORER":
		return t.Teal
	}
	return t.Blue
}
