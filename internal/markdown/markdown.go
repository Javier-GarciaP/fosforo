// Package markdown convierte Markdown (CommonMark + GFM) en líneas con estilo listas para dibujar en una terminal, y
// calcula anclas para relacionar posiciones del fuente con las del resultado. No depende de la interfaz: los colores se
// expresan como «roles» que cada tema traduce.
package markdown

import (
	"sort"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

// MaxBytes es el tamaño máximo de un documento con vista previa; por encima se muestra el fuente.
const MaxBytes = 2 << 20

// Role es la función de un fragmento de texto; el tema decide su color.
type Role uint8

const (
	RoleText       Role = iota
	RoleHeading         // texto de encabezados (en negrita)
	RoleRule            // reglas bajo H1/H2 y líneas horizontales
	RoleBullet          // viñetas y números de listas
	RoleDone            // casilla marcada
	RoleTodo            // casilla pendiente
	RoleQuoteBar        // barra lateral de las citas
	RoleQuote           // texto de una cita
	RoleLink            // enlaces
	RoleInlineCode      // `código` en línea
	RoleCode            // texto de bloque de código sin color de sintaxis
	RoleTableLine       // bordes de tablas
	RoleDim             // avisos de contenido omitido
	RoleImage           // [imagen: alt]
)

// Run es un fragmento de línea con un mismo estilo.
type Run struct {
	Text      string
	Role      Role
	Bold      bool
	Italic    bool
	Strike    bool
	Underline bool
	Syntax    syntax.Kind // categoría de sintaxis en bloques de código (Plain: ninguna)
}

// Fill indica si la línea lleva fondo propio.
type Fill uint8

const (
	FillNone Fill = iota
	FillCode      // bloque de código: el fondo se extiende a todo el ancho del documento
)

// Line es una línea del documento renderizado.
type Line struct {
	Runs []Run
	Fill Fill
}

// Width devuelve el ancho de la línea en celdas.
func (l Line) Width() int {
	w := 0
	for _, r := range l.Runs {
		w += uniseg.StringWidth(r.Text)
	}
	return w
}

// Plain devuelve el texto de la línea sin estilos.
func (l Line) Plain() string {
	var sb strings.Builder
	for _, r := range l.Runs {
		sb.WriteString(r.Text)
	}
	return sb.String()
}

// Anchor relaciona una línea del fuente (Src, desde 0) con la primera línea de su bloque en el resultado (Out).
type Anchor struct {
	Src, Out int
	Heading  bool
}

// Doc es el resultado de renderizar un documento para un ancho dado.
type Doc struct {
	Lines    []Line
	Anchors  []Anchor // bloques de primer nivel, en orden
	SrcLines int      // número de líneas del fuente
	Width    int      // ancho para el que se renderizó
	Omitted  bool     // se omitió contenido que una terminal no puede mostrar (HTML, diagramas)
}

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Render renderiza src ajustando el texto a width columnas.
func Render(src string, width int) Doc {
	width = max(width, 10)
	b := []byte(src)
	root := md.Parser().Parse(text.NewReader(b))
	r := &renderer{src: b, lineStarts: lineStarts(b)}
	doc := Doc{Width: width, SrcLines: len(r.lineStarts)}
	r.top(root, width, &doc)
	doc.Omitted = r.omitted
	for i, l := range doc.Lines { // red de seguridad: ninguna línea se sale del ancho (p. ej. tablas con demasiadas columnas)
		if l.Width() > width {
			doc.Lines[i].Runs = truncateRuns(l.Runs, width)
		}
	}
	if r.omitted {
		if len(doc.Lines) > 0 {
			doc.Lines = append(doc.Lines, Line{})
		}
		note := []atom{{text: "(contenido que la terminal no puede mostrar: omitido)", st: style{role: RoleDim}}}
		doc.Lines = append(doc.Lines, wrap(splitWords(note), width)...)
	}
	return doc
}

type renderer struct {
	src        []byte
	lineStarts []int
	omitted    bool
}

func lineStarts(b []byte) []int {
	starts := []int{0}
	for i, c := range b {
		if c == '\n' && i+1 <= len(b) {
			starts = append(starts, i+1)
		}
	}
	if len(starts) > 1 && starts[len(starts)-1] == len(b) { // el último salto de línea no abre una línea nueva
		starts = starts[:len(starts)-1]
	}
	return starts
}

func (r *renderer) lineOf(offset int) int {
	return max(0, sort.Search(len(r.lineStarts), func(i int) bool { return r.lineStarts[i] > offset })-1)
}

// startOffset devuelve el desplazamiento del fuente donde empieza un nodo.
func (r *renderer) startOffset(n ast.Node) int {
	if t, ok := n.(*ast.Text); ok {
		return t.Segment.Start
	}
	if fc, ok := n.(*ast.FencedCodeBlock); ok {
		if fc.Info != nil {
			return fc.Info.Segment.Start
		}
		if fc.Lines().Len() > 0 {
			return max(0, fc.Lines().At(0).Start-1)
		}
	}
	if n.Type() == ast.TypeBlock && n.Lines().Len() > 0 {
		return n.Lines().At(0).Start
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if off := r.startOffset(c); off >= 0 {
			return off
		}
	}
	return -1
}

func (r *renderer) srcLine(n ast.Node) int {
	if off := r.startOffset(n); off >= 0 {
		return r.lineOf(off)
	}
	return -1
}

// top renderiza los bloques de primer nivel anotando un ancla por cada uno.
func (r *renderer) top(root ast.Node, width int, doc *Doc) {
	first := true
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		lines := r.block(n, width)
		if len(lines) == 0 {
			continue
		}
		if !first {
			doc.Lines = append(doc.Lines, Line{})
		}
		first = false
		if src := r.srcLine(n); src >= 0 {
			_, isHeading := n.(*ast.Heading)
			doc.Anchors = append(doc.Anchors, Anchor{Src: src, Out: len(doc.Lines), Heading: isHeading})
		}
		doc.Lines = append(doc.Lines, lines...)
	}
}

// blocks renderiza los hijos de un contenedor separados por una línea en blanco.
func (r *renderer) blocks(parent ast.Node, width int, tight bool) []Line {
	var out []Line
	for n := parent.FirstChild(); n != nil; n = n.NextSibling() {
		lines := r.block(n, width)
		if len(lines) == 0 {
			continue
		}
		if len(out) > 0 && !tight {
			out = append(out, Line{})
		}
		out = append(out, lines...)
	}
	return out
}

func (r *renderer) block(n ast.Node, width int) []Line {
	switch n := n.(type) {
	case *ast.Heading:
		return r.heading(n, width)
	case *ast.Paragraph, *ast.TextBlock:
		return wrap(r.inline(n, style{}), width)
	case *ast.List:
		return r.list(n, width)
	case *ast.Blockquote:
		return r.quote(n, width)
	case *ast.FencedCodeBlock:
		return r.code(n, width, string(n.Language(r.src)))
	case *ast.CodeBlock:
		return r.code(n, width, "")
	case *ast.ThematicBreak:
		return []Line{rule(width)}
	case *east.Table:
		return r.table(n, width)
	case *ast.HTMLBlock:
		r.omitted = true
		return nil
	}
	// contenedor desconocido: se renderizan sus hijos
	return r.blocks(n, width, false)
}

func rule(width int) Line {
	return Line{Runs: []Run{{Text: strings.Repeat("─", width), Role: RoleRule}}}
}

func (r *renderer) heading(h *ast.Heading, width int) []Line {
	lines := wrap(r.inline(h, style{bold: true, role: RoleHeading}), width)
	if h.Level <= 2 {
		lines = append(lines, rule(width))
	}
	return lines
}

func (r *renderer) quote(q *ast.Blockquote, width int) []Line {
	inner := r.blocks(q, max(width-2, 4), false)
	for i := range inner {
		for j := range inner[i].Runs {
			if inner[i].Runs[j].Role == RoleText {
				inner[i].Runs[j].Role = RoleQuote
			}
		}
	}
	return prefixLines(inner, []Run{{Text: "┃ ", Role: RoleQuoteBar}}, []Run{{Text: "┃ ", Role: RoleQuoteBar}})
}

// prefixLines antepone first a la primera línea y rest a las demás.
func prefixLines(lines []Line, first, rest []Run) []Line {
	out := make([]Line, len(lines))
	for i, l := range lines {
		p := rest
		if i == 0 {
			p = first
		}
		runs := append(append([]Run(nil), p...), l.Runs...)
		out[i] = Line{Runs: runs, Fill: l.Fill}
	}
	return out
}

func (r *renderer) list(l *ast.List, width int) []Line {
	var out []Line
	num := l.Start
	if num == 0 && l.IsOrdered() {
		num = 0
	}
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		marker, role := "• ", RoleBullet
		if l.IsOrdered() {
			marker = itoa(num) + string(l.Marker) + " "
			num++
		}
		if cb := firstTaskCheckBox(item); cb != nil {
			if cb.IsChecked {
				marker, role = "☑ ", RoleDone
			} else {
				marker, role = "☐ ", RoleTodo
			}
			cb.Parent().RemoveChild(cb.Parent(), cb)
		}
		indent := strings.Repeat(" ", uniseg.StringWidth(marker))
		inner := r.blocks(item, max(width-uniseg.StringWidth(marker), 4), true)
		if len(inner) == 0 {
			inner = []Line{{}}
		}
		if !l.IsTight && len(out) > 0 {
			out = append(out, Line{})
		}
		out = append(out, prefixLines(inner, []Run{{Text: marker, Role: role}}, []Run{{Text: indent}})...)
	}
	return out
}

func firstTaskCheckBox(item ast.Node) *east.TaskCheckBox {
	c := item.FirstChild()
	if c == nil {
		return nil
	}
	if first := c.FirstChild(); first != nil {
		if cb, ok := first.(*east.TaskCheckBox); ok {
			return cb
		}
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func (r *renderer) code(n ast.Node, width int, lang string) []Line {
	var sb strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		sb.Write(seg.Value(r.src))
	}
	src := strings.TrimRight(sb.String(), "\n")
	if strings.EqualFold(lang, "mermaid") {
		r.omitted = true
		return []Line{{Runs: []Run{{Text: "[diagrama mermaid: no se puede mostrar]", Role: RoleDim}}}}
	}
	lines := strings.Split(strings.ReplaceAll(src, "\r", ""), "\n")
	for i := range lines {
		lines[i] = strings.ReplaceAll(lines[i], "\t", "    ") // se expanden antes de tokenizar para que los tramos cuadren
	}
	var spans [][]syntax.Span
	if lang != "" {
		spans, _ = syntax.Tokenize(lang, strings.Join(lines, "\n"))
	}
	inner := max(width-2, 4)
	out := make([]Line, 0, len(lines))
	for i, ln := range lines {
		var sp []syntax.Span
		if i < len(spans) {
			sp = spans[i]
		}
		out = append(out, codeLine(ln, sp, inner))
	}
	if len(out) == 0 {
		out = []Line{codeLine("", nil, inner)}
	}
	return out
}

// codeLine construye una línea de código con relleno de un carácter a cada lado y recorte con «…» si no cabe.
func codeLine(text string, spans []syntax.Span, inner int) Line {
	if uniseg.StringWidth(text) > inner {
		text, spans = truncate(text, inner), nil
	}
	runs := []Run{{Text: " ", Role: RoleCode}}
	bo, si := 0, 0
	var cur Run
	flush := func() {
		if cur.Text != "" {
			runs = append(runs, cur)
			cur = Run{}
		}
	}
	for _, g := range graphemes(text) {
		kind := syntax.Plain
		for si < len(spans) && spans[si].End <= bo {
			si++
		}
		if si < len(spans) && spans[si].Start <= bo {
			kind = spans[si].Kind
		}
		bo += len(g)
		if cur.Text != "" && cur.Syntax != kind {
			flush()
		}
		cur.Role, cur.Syntax = RoleCode, kind
		cur.Text += g
	}
	flush()
	return Line{Runs: runs, Fill: FillCode}
}

func graphemes(s string) []string {
	var out []string
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		out = append(out, g.Str())
	}
	return out
}

// truncate recorta s a w celdas terminando en «…».
func truncate(s string, w int) string {
	if uniseg.StringWidth(s) <= w {
		return s
	}
	var sb strings.Builder
	acc := 0
	for _, g := range graphemes(s) {
		gw := uniseg.StringWidth(g)
		if acc+gw > w-1 {
			break
		}
		sb.WriteString(g)
		acc += gw
	}
	return sb.String() + "…"
}

// splitWords divide átomos de frase completa en palabras y espacios para poder ajustarlos.
func splitWords(in []atom) []atom {
	var out []atom
	for _, a := range in {
		out = append(out, words(a.text, a.st)...)
	}
	return out
}
