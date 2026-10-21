// Package buffer guarda el texto de un archivo en memoria y lo modifica de forma atómica.
// No depende de la interfaz: se prueba con pruebas de tabla.
package buffer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// MaxEditableSize es el tamaño máximo que se puede editar; los archivos mayores se abren en solo lectura.
const MaxEditableSize = 10 << 20

// Errores devueltos por el buffer.
var (
	ErrReadOnly   = errors.New("el buffer es de solo lectura")
	ErrNoPath     = errors.New("el buffer no tiene nombre de archivo")
	ErrInvalidPos = errors.New("posición fuera del texto")
	ErrMismatch   = errors.New("el texto a eliminar no coincide con el del buffer")
)

// Edit es una modificación atómica: en Start se elimina Deleted y se inserta Inserted.
// Toda edición del texto pasa por aquí, así el historial de deshacer (modal-editing) solo necesita invertirlas.
type Edit struct {
	Start    Pos
	Deleted  string
	Inserted string
}

// Invert devuelve la edición que deshace e.
func (e Edit) Invert() Edit {
	return Edit{Start: e.Start, Deleted: e.Inserted, Inserted: e.Deleted}
}

// Buffer es el texto de un archivo en memoria. Las líneas se guardan sin su terminador.
type Buffer struct {
	path      string
	lines     []string
	crlf      bool // terminador CRLF en lugar de LF
	finalNL   bool // el archivo termina con salto de línea
	origEmpty bool // el archivo en disco estaba vacío (o no existía)
	readOnly  bool
	roReason  string

	// Historial de deshacer (ver history.go). La versión cuenta las ediciones aplicadas; el buffer está
	// modificado cuando su versión difiere de la que se guardó.
	hist         history
	rev          int        // contador de cambios del texto; solo crece (también al deshacer y rehacer)
	onChange     func(Edit) // se llama tras cada cambio del texto (p. ej. para mantener el resaltado)
	version      int
	savedVersion int // versión guardada en disco; -1 si ya no se puede volver a ella
}

// New crea un buffer vacío con nombre opcional. No toca el disco.
func New(path string) *Buffer {
	return &Buffer{path: path, lines: []string{""}, finalNL: true, origEmpty: true}
}

// FromString crea un buffer a partir de contenido ya leído, sin tocar el disco.
func FromString(path, content string) *Buffer {
	b := &Buffer{path: path}
	b.parse(content)
	return b
}

// Open carga un archivo. Si no existe devuelve un buffer vacío con ese nombre (se crea al guardar).
func Open(path string) (*Buffer, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return New(path), nil
	}
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s es un directorio", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := &Buffer{path: path}
	b.parse(string(data))
	switch {
	case len(data) > MaxEditableSize:
		b.setReadOnly("el archivo supera los 10 MiB")
	case bytes.IndexByte(data, 0) >= 0:
		b.setReadOnly("contiene bytes NUL (archivo binario)")
	case !utf8.Valid(data):
		b.setReadOnly("no es UTF-8 válido")
	}
	return b, nil
}

func (b *Buffer) setReadOnly(reason string) { b.readOnly, b.roReason = true, reason }

func (b *Buffer) parse(content string) {
	b.crlf = strings.Contains(content, "\r\n") && strings.Count(content, "\r\n") == strings.Count(content, "\n")
	eol := b.eol()
	b.origEmpty = content == ""
	if content == "" {
		b.lines, b.finalNL = []string{""}, true
		return
	}
	b.finalNL = strings.HasSuffix(content, eol)
	body := strings.TrimSuffix(content, eol)
	b.lines = strings.Split(body, eol)
}

func (b *Buffer) eol() string {
	if b.crlf {
		return "\r\n"
	}
	return "\n"
}

// Text devuelve el contenido tal como se guardaría en disco.
func (b *Buffer) Text() string {
	if len(b.lines) == 1 && b.lines[0] == "" && b.origEmpty {
		return ""
	}
	s := strings.Join(b.lines, b.eol())
	if b.finalNL {
		s += b.eol()
	}
	return s
}

// Path devuelve el nombre del archivo ("" si no tiene).
func (b *Buffer) Path() string { return b.path }

// Modified indica si hay cambios sin guardar. Deshacer hasta el punto guardado deja el buffer sin modificar.
func (b *Buffer) Modified() bool { return b.version != b.savedVersion }

// ReadOnly indica si el buffer no se puede editar y por qué.
func (b *Buffer) ReadOnly() (bool, string) { return b.readOnly, b.roReason }

// CRLF indica si el archivo usa terminadores CRLF.
func (b *Buffer) CRLF() bool { return b.crlf }

// LineCount devuelve el número de líneas (siempre al menos 1).
func (b *Buffer) LineCount() int { return len(b.lines) }

// Line devuelve la línea i sin terminador, o "" si no existe.
func (b *Buffer) Line(i int) string {
	if i < 0 || i >= len(b.lines) {
		return ""
	}
	return b.lines[i]
}

// LineLen devuelve el número de grafemas de la línea i.
func (b *Buffer) LineLen(i int) int { return Count(b.Line(i)) }

// Valid indica si p está dentro del texto.
func (b *Buffer) Valid(p Pos) bool {
	return p.Line >= 0 && p.Line < len(b.lines) && p.Col >= 0 && p.Col <= b.LineLen(p.Line)
}

// Clamp acerca p a la posición válida más cercana.
func (b *Buffer) Clamp(p Pos) Pos {
	p.Line = max(0, min(p.Line, len(b.lines)-1))
	p.Col = max(0, min(p.Col, b.LineLen(p.Line)))
	return p
}

// Range devuelve el texto entre start (incluida) y end (excluida), con saltos de línea "\n".
func (b *Buffer) Range(start, end Pos) string {
	if end.Less(start) {
		start, end = end, start
	}
	start, end = b.Clamp(start), b.Clamp(end)
	if start.Line == end.Line {
		l := b.lines[start.Line]
		return l[byteOffset(l, start.Col):byteOffset(l, end.Col)]
	}
	var sb strings.Builder
	first := b.lines[start.Line]
	sb.WriteString(first[byteOffset(first, start.Col):])
	for i := start.Line + 1; i < end.Line; i++ {
		sb.WriteByte('\n')
		sb.WriteString(b.lines[i])
	}
	sb.WriteByte('\n')
	last := b.lines[end.Line]
	sb.WriteString(last[:byteOffset(last, end.Col)])
	return sb.String()
}

// Apply aplica una edición, la anota en el historial de deshacer y devuelve la posición justo después del texto insertado.
// Falla sin modificar nada si el buffer es de solo lectura, la posición no existe o el texto a eliminar no coincide.
func (b *Buffer) Apply(e Edit) (Pos, error) {
	end, err := b.apply(e)
	if err == nil {
		b.record(e)
	}
	return end, err
}

// apply aplica una edición sin tocar el historial.
func (b *Buffer) apply(e Edit) (Pos, error) {
	if b.readOnly {
		return e.Start, ErrReadOnly
	}
	if !b.Valid(e.Start) {
		return e.Start, ErrInvalidPos
	}
	end := endOf(e.Start, e.Deleted)
	if !b.Valid(end) {
		return e.Start, ErrInvalidPos
	}
	if b.Range(e.Start, end) != e.Deleted {
		return e.Start, ErrMismatch
	}
	first, last := b.lines[e.Start.Line], b.lines[end.Line]
	merged := first[:byteOffset(first, e.Start.Col)] + strings.ReplaceAll(e.Inserted, "\r\n", "\n") + last[byteOffset(last, end.Col):]
	repl := strings.Split(merged, "\n")
	lines := make([]string, 0, len(b.lines)-(end.Line-e.Start.Line)+len(repl)-1)
	lines = append(lines, b.lines[:e.Start.Line]...)
	lines = append(lines, repl...)
	lines = append(lines, b.lines[end.Line+1:]...)
	b.lines = lines
	b.rev++
	if b.onChange != nil {
		b.onChange(Edit{Start: e.Start, Deleted: e.Deleted, Inserted: strings.ReplaceAll(e.Inserted, "\r\n", "\n")})
	}
	return endOf(e.Start, strings.ReplaceAll(e.Inserted, "\r\n", "\n")), nil
}

// Insert inserta text en p y devuelve la edición aplicada y la posición tras el texto.
func (b *Buffer) Insert(p Pos, text string) (Edit, Pos, error) {
	e := Edit{Start: p, Inserted: strings.ReplaceAll(text, "\r\n", "\n")}
	end, err := b.Apply(e)
	return e, end, err
}

// Delete elimina el texto entre start y end y devuelve la edición aplicada.
func (b *Buffer) Delete(start, end Pos) (Edit, error) {
	if end.Less(start) {
		start, end = end, start
	}
	e := Edit{Start: start, Deleted: b.Range(start, end)}
	_, err := b.Apply(e)
	return e, err
}

// Prev devuelve la posición anterior a p (cruzando al final de la línea anterior) y false si p es el inicio del texto.
func (b *Buffer) Prev(p Pos) (Pos, bool) {
	switch {
	case p.Col > 0:
		return Pos{p.Line, p.Col - 1}, true
	case p.Line > 0:
		return Pos{p.Line - 1, b.LineLen(p.Line - 1)}, true
	}
	return p, false
}

// Next devuelve la posición siguiente a p (cruzando al inicio de la línea siguiente) y false si p es el final del texto.
func (b *Buffer) Next(p Pos) (Pos, bool) {
	switch {
	case p.Col < b.LineLen(p.Line):
		return Pos{p.Line, p.Col + 1}, true
	case p.Line < len(b.lines)-1:
		return Pos{p.Line + 1, 0}, true
	}
	return p, false
}

// SetPath cambia el nombre del archivo sin tocar el texto ni el estado de modificado (p. ej. tras renombrarlo en disco).
func (b *Buffer) SetPath(path string) { b.path = path }

// Rev devuelve el contador de cambios del texto. Distinto de Modified: crece con cada edición, deshacer y rehacer,
// de modo que dos lecturas con el mismo valor garantizan que el texto no cambió entre ellas.
func (b *Buffer) Rev() int { return b.rev }

// OnChange registra una función que se ejecuta tras cada cambio del texto con la edición aplicada.
// Solo hay una; registrar otra sustituye a la anterior (nil la quita).
func (b *Buffer) OnChange(fn func(Edit)) { b.onChange = fn }
