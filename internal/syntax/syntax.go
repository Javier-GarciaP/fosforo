// Package syntax tokeniza código con Chroma y lo traduce a categorías abstractas (palabra clave, tipo, cadena...)
// que cada tema convierte en colores. No depende de la interfaz.
package syntax

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Límites de seguridad: por encima de ellos se muestra el texto sin color.
const (
	MaxFileBytes = 2 << 20 // 2 MiB
	MaxLineBytes = 2000    // caracteres por línea
)

// Kind es una categoría abstracta de token; el tema decide su color, así cambiar de tema no obliga a retokenizar.
type Kind uint8

const (
	Plain Kind = iota
	Keyword
	Type
	Function
	Builtin
	String
	Escape
	Number
	Constant
	Comment
	Tag
	Attribute
	Error
)

var kindNames = [...]string{"plain", "keyword", "type", "function", "builtin", "string", "escape", "number", "constant", "comment", "tag", "attribute", "error"}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "?"
}

// Span es un tramo de una línea: bytes [Start, End) de esa línea con la categoría Kind.
// Los tramos de texto sin categoría (Plain) no se guardan.
type Span struct {
	Start, End int
	Kind       Kind
}

// byName fija el lenguaje de nombres de archivo que Chroma reconocería mal por su extensión.
var byName = map[string]string{"go.mod": "Go", "go.work": "Go", "go.sum": ""}

// Detect elige el lenguaje por el nombre o la extensión del archivo. Devuelve "" para texto plano o tipos desconocidos.
func Detect(name string) string {
	base := filepath.Base(name)
	if base == "" || base == "." {
		return ""
	}
	if lang, ok := byName[base]; ok {
		return lang
	}
	lx := lexers.Match(base)
	if lx == nil {
		lx = lexers.Match(strings.ToLower(base)) // las extensiones no distinguen mayúsculas: TRIE.GO es Go
	}
	if lx == nil {
		return ""
	}
	n := lx.Config().Name
	if strings.EqualFold(n, "plaintext") || strings.EqualFold(n, "text") {
		return ""
	}
	return n
}

// Tokenize colorea text con el lenguaje lang y devuelve los tramos de cada línea (una entrada por línea de text, sin
// el terminador). Las construcciones de varias líneas se resuelven con el estado del analizador. Un lenguaje desconocido
// o un texto demasiado grande devuelve nil. Las líneas de más de MaxLineBytes se dejan sin color.
func Tokenize(lang, text string) ([][]Span, error) {
	if lang == "" || len(text) > MaxFileBytes {
		return nil, nil
	}
	lx := lexers.Get(lang)
	if lx == nil {
		return nil, nil
	}
	lx = chroma.Coalesce(lx)

	lines := strings.Split(text, "\n")
	long := make([]bool, len(lines))
	for i, l := range lines {
		if len(l) > MaxLineBytes {
			long[i] = true
			lines[i] = "" // se tokeniza vacía para no frenar al analizador, pero se conserva el número de líneas
		}
	}
	it, err := lx.Tokenise(&chroma.TokeniseOptions{State: "root", EnsureLF: true}, strings.Join(lines, "\n"))
	if err != nil {
		return nil, err
	}

	out := make([][]Span, len(lines))
	line, off := 0, 0
	for tok := it(); tok != chroma.EOF; tok = it() {
		kind := kindOf(tok.Type)
		parts := strings.Split(tok.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				line, off = line+1, 0
			}
			if line >= len(out) {
				break
			}
			if part != "" && kind != Plain && !long[line] {
				if n := len(out[line]); n > 0 && out[line][n-1].Kind == kind && out[line][n-1].End == off {
					out[line][n-1].End = off + len(part) // se funden tramos contiguos de la misma categoría
				} else {
					out[line] = append(out[line], Span{Start: off, End: off + len(part), Kind: kind})
				}
			}
			off += len(part)
		}
	}
	return out, nil
}

// kindOf traduce un tipo de token de Chroma a una categoría abstracta.
func kindOf(t chroma.TokenType) Kind {
	switch {
	case t == chroma.KeywordType, t == chroma.NameClass, t == chroma.NameException, t == chroma.NameNamespace:
		return Type
	case t == chroma.KeywordConstant, t == chroma.NameConstant:
		return Constant
	case t.InCategory(chroma.Keyword):
		return Keyword
	case t == chroma.NameBuiltin, t == chroma.NameBuiltinPseudo:
		return Builtin
	case t.InSubCategory(chroma.NameFunction), t == chroma.NameDecorator:
		return Function
	case t == chroma.NameTag:
		return Tag
	case t == chroma.NameAttribute:
		return Attribute
	case t == chroma.LiteralStringEscape:
		return Escape
	case t.InSubCategory(chroma.LiteralString):
		return String
	case t.InSubCategory(chroma.LiteralNumber):
		return Number
	case t == chroma.CommentPreproc, t == chroma.CommentPreprocFile:
		return Keyword
	case t.InCategory(chroma.Comment):
		return Comment
	case t == chroma.GenericHeading, t == chroma.GenericSubheading:
		return Function
	case t == chroma.GenericStrong:
		return Keyword
	case t == chroma.GenericEmph:
		return Type
	case t == chroma.GenericInserted:
		return String
	case t == chroma.GenericDeleted, t == chroma.Error, t == chroma.GenericError:
		return Error
	}
	return Plain
}
