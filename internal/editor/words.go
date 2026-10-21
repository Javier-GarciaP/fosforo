package editor

import (
	"unicode"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

// Clases de carácter para los movimientos por palabra, como en Vim: blancos, palabra (letras, dígitos, «_») y puntuación.
const (
	classBlank = iota
	classWord
	classPunct
)

func charClass(cluster string) int {
	for _, r := range cluster {
		switch {
		case r == ' ' || r == '\t' || unicode.IsSpace(r):
			return classBlank
		case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r):
			return classWord
		}
		return classPunct
	}
	return classBlank
}

// nextWordStart devuelve el inicio de la palabra siguiente. Una línea vacía cuenta como una palabra.
// En la última palabra del archivo devuelve la posición justo después del último carácter.
func nextWordStart(b *buffer.Buffer, p buffer.Pos) buffer.Pos {
	line, col := p.Line, p.Col
	cl := buffer.Clusters(b.Line(line))
	if col < len(cl) {
		if c := charClass(cl[col]); c != classBlank {
			for col < len(cl) && charClass(cl[col]) == c {
				col++
			}
		}
	}
	for {
		for col < len(cl) && charClass(cl[col]) == classBlank {
			col++
		}
		if col < len(cl) {
			return buffer.Pos{Line: line, Col: col}
		}
		if line == b.LineCount()-1 {
			return buffer.Pos{Line: line, Col: len(cl)}
		}
		line++
		cl = buffer.Clusters(b.Line(line))
		col = 0
		if len(cl) == 0 {
			return buffer.Pos{Line: line}
		}
	}
}

// prevWordStart devuelve el inicio de la palabra anterior (o el inicio del archivo).
func prevWordStart(b *buffer.Buffer, p buffer.Pos) buffer.Pos {
	line, col := p.Line, p.Col
	cl := buffer.Clusters(b.Line(line))
	for {
		if col == 0 {
			if line == 0 {
				return buffer.Pos{}
			}
			line--
			cl = buffer.Clusters(b.Line(line))
			col = len(cl)
			if len(cl) == 0 {
				return buffer.Pos{Line: line}
			}
			continue
		}
		col--
		if charClass(cl[col]) != classBlank {
			break
		}
	}
	c := charClass(cl[col])
	for col > 0 && charClass(cl[col-1]) == c {
		col--
	}
	return buffer.Pos{Line: line, Col: col}
}

// wordEnd devuelve el último carácter de la palabra actual o, si ya está en él, de la siguiente.
func wordEnd(b *buffer.Buffer, p buffer.Pos) buffer.Pos {
	line, col := p.Line, p.Col+1
	cl := buffer.Clusters(b.Line(line))
	for {
		for col < len(cl) && charClass(cl[col]) == classBlank {
			col++
		}
		if col < len(cl) {
			break
		}
		if line == b.LineCount()-1 {
			return buffer.Pos{Line: line, Col: max(0, len(cl)-1)}
		}
		line++
		cl = buffer.Clusters(b.Line(line))
		col = 0
	}
	c := charClass(cl[col])
	for col+1 < len(cl) && charClass(cl[col+1]) == c {
		col++
	}
	return buffer.Pos{Line: line, Col: col}
}

// endOfCurrentWord devuelve el último carácter de la palabra bajo el cursor (o p si está en un blanco).
func endOfCurrentWord(b *buffer.Buffer, p buffer.Pos) buffer.Pos {
	cl := buffer.Clusters(b.Line(p.Line))
	if p.Col >= len(cl) || charClass(cl[p.Col]) == classBlank {
		return p
	}
	c, col := charClass(cl[p.Col]), p.Col
	for col+1 < len(cl) && charClass(cl[col+1]) == c {
		col++
	}
	return buffer.Pos{Line: p.Line, Col: col}
}

// firstNonBlank devuelve el índice del primer carácter que no es espacio ni tabulador (o el último si toda la línea es blanca).
func firstNonBlank(line string) int {
	cl := buffer.Clusters(line)
	for i, c := range cl {
		if c != " " && c != "\t" {
			return i
		}
	}
	return max(0, len(cl)-1)
}
