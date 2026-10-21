package ui

// Theme son los colores de la interfaz ("Tinta", ver docs/design/README.md), en hexadecimal.
type Theme struct {
	BG, Panel, Sel, Border string // fondos y bordes
	FG, Dim                string // texto principal y atenuado
	Blue, Green, Mauve     string
	Peach, Yellow, Red     string
	Teal, Brand            string
}

// Dark es el tema oscuro de la paleta Tinta (docs/design/src/palette.js).
func Dark() Theme {
	return Theme{
		BG: "#171b27", Panel: "#0f131f", Sel: "#232c42", Border: "#424a5e",
		FG: "#cfd7eb", Dim: "#75809c",
		Blue: "#7fc1fe", Green: "#7fd096", Mauve: "#cca8f6",
		Peach: "#f5a574", Yellow: "#d5b85e", Red: "#fc9c9f",
		Teal: "#4ad1d3", Brand: "#69f19a",
	}
}

// modeColor devuelve el color de la etiqueta de cada modo.
func (t Theme) modeColor(mode string) string {
	switch mode {
	case "INSERT":
		return t.Green
	case "VISUAL":
		return t.Mauve
	case "COMMAND":
		return t.Peach
	case "PREVIEW":
		return t.Teal
	}
	return t.Blue
}
