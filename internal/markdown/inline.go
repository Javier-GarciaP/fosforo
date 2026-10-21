package markdown

import (
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// style es el estilo acumulado al recorrer los nodos en línea.
type style struct {
	role      Role
	bold      bool
	italic    bool
	strike    bool
	underline bool
}

func (s style) run(text string) Run {
	return Run{Text: text, Role: s.role, Bold: s.bold, Italic: s.italic, Strike: s.strike, Underline: s.underline}
}

// atom es una pieza de texto en línea: una palabra, un espacio (donde se puede cortar la línea) o un salto forzado.
type atom struct {
	text    string
	st      style
	space   bool
	newline bool
}

// inline convierte los hijos en línea de un bloque en átomos.
func (r *renderer) inline(parent ast.Node, base style) []atom {
	var out []atom
	var walk func(n ast.Node, st style)
	walk = func(n ast.Node, st style) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *ast.Text:
				out = append(out, words(string(c.Segment.Value(r.src)), st)...)
				if c.HardLineBreak() {
					out = append(out, atom{newline: true})
				} else if c.SoftLineBreak() {
					out = append(out, atom{space: true, text: " ", st: st})
				}
			case *ast.String:
				out = append(out, words(string(c.Value), st)...)
			case *ast.CodeSpan:
				var sb strings.Builder
				for cc := c.FirstChild(); cc != nil; cc = cc.NextSibling() {
					if t, ok := cc.(*ast.Text); ok {
						sb.Write(t.Segment.Value(r.src))
					}
				}
				cs := st
				cs.role = RoleInlineCode
				out = append(out, words(strings.ReplaceAll(sb.String(), "\n", " "), cs)...)
			case *ast.Emphasis:
				es := st
				if c.Level >= 2 {
					es.bold = true
				} else {
					es.italic = true
				}
				walk(c, es)
			case *east.Strikethrough:
				ss := st
				ss.strike = true
				walk(c, ss)
			case *ast.Link:
				ls := st
				ls.role, ls.underline = RoleLink, true
				walk(c, ls)
			case *ast.AutoLink:
				ls := st
				ls.role, ls.underline = RoleLink, true
				out = append(out, words(string(c.Label(r.src)), ls)...)
			case *ast.Image:
				alt := plainText(c, r.src)
				is := st
				is.role = RoleImage
				out = append(out, words("[imagen: "+alt+"]", is)...)
			case *ast.RawHTML:
				if isBreakTag(c, r.src) {
					out = append(out, atom{newline: true})
				} else {
					r.omitted = true
				}
			case *east.TaskCheckBox:
				// lo gestiona la lista
			default:
				walk(c, st)
			}
		}
	}
	walk(parent, base)
	return out
}

func isBreakTag(n *ast.RawHTML, src []byte) bool {
	var sb strings.Builder
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		sb.Write(seg.Value(src))
	}
	t := strings.ToLower(strings.ReplaceAll(sb.String(), " ", ""))
	return t == "<br>" || t == "<br/>"
}

// plainText devuelve el texto de un nodo sin formato (para el alt de las imágenes).
func plainText(n ast.Node, src []byte) string {
	var sb strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *ast.Text:
				sb.Write(c.Segment.Value(src))
			case *ast.String:
				sb.Write(c.Value)
			default:
				walk(c)
			}
		}
	}
	walk(n)
	return sb.String()
}

// words divide texto en palabras y espacios conservando el estilo.
func words(s string, st style) []atom {
	var out []atom
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, atom{text: cur.String(), st: st})
			cur.Reset()
		}
	}
	for _, rn := range s {
		if unicode.IsSpace(rn) {
			flush()
			if n := len(out); n == 0 || !out[n-1].space {
				out = append(out, atom{space: true, text: " ", st: st})
			}
			continue
		}
		if unicode.IsControl(rn) {
			continue
		}
		cur.WriteRune(rn)
	}
	flush()
	return out
}

// wrap coloca los átomos en líneas de como máximo width celdas, cortando por los espacios y partiendo las palabras
// que no caben en una línea entera.
func wrap(atoms []atom, width int) []Line {
	var lines []Line
	var cur []Run
	cw := 0
	endLine := func() {
		// se quita el espacio final
		for len(cur) > 0 && strings.TrimRight(cur[len(cur)-1].Text, " ") == "" {
			cur = cur[:len(cur)-1]
		}
		lines = append(lines, Line{Runs: mergeRuns(cur)})
		cur, cw = nil, 0
	}
	add := func(text string, st style) {
		cur = append(cur, st.run(text))
		cw += uniseg.StringWidth(text)
	}
	for _, a := range atoms {
		switch {
		case a.newline:
			endLine()
		case a.space:
			if cw == 0 {
				continue
			}
			if cw+1 > width {
				endLine()
				continue
			}
			add(" ", a.st)
		default:
			w := uniseg.StringWidth(a.text)
			if cw+w <= width {
				add(a.text, a.st)
				continue
			}
			if cw > 0 {
				endLine()
			}
			if w <= width {
				add(a.text, a.st)
				continue
			}
			// palabra más ancha que la línea: se parte por grafemas
			for _, g := range graphemes(a.text) {
				gw := uniseg.StringWidth(g)
				if cw+gw > width && cw > 0 {
					endLine()
				}
				add(g, a.st)
			}
		}
	}
	if len(cur) > 0 || len(lines) == 0 {
		endLine()
	}
	// una línea vacía final (p. ej. por un salto forzado al final) se descarta
	for len(lines) > 1 && len(lines[len(lines)-1].Runs) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// mergeRuns une fragmentos contiguos con el mismo estilo.
func mergeRuns(runs []Run) []Run {
	var out []Run
	for _, r := range runs {
		if n := len(out); n > 0 {
			p := &out[n-1]
			if p.Role == r.Role && p.Bold == r.Bold && p.Italic == r.Italic && p.Strike == r.Strike && p.Underline == r.Underline && p.Syntax == r.Syntax {
				p.Text += r.Text
				continue
			}
		}
		out = append(out, r)
	}
	return out
}
