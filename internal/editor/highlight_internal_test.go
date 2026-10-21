package editor

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/syntax"
)

func sp(start, end int, k syntax.Kind) syntax.Span {
	return syntax.Span{Start: start, End: end, Kind: k}
}

func TestShiftSpans(t *testing.T) {
	kw, str := syntax.Keyword, syntax.String
	for _, tc := range []struct {
		name            string
		in              []syntax.Span
		off, del, delta int
		want            []syntax.Span
	}{
		{"insertar antes desplaza", []syntax.Span{sp(4, 8, kw)}, 0, 0, 2, []syntax.Span{sp(6, 10, kw)}},
		{"insertar después no toca", []syntax.Span{sp(0, 4, kw)}, 6, 0, 3, []syntax.Span{sp(0, 4, kw)}},
		{"insertar justo al final del tramo no lo extiende", []syntax.Span{sp(0, 4, kw)}, 4, 0, 1, []syntax.Span{sp(0, 4, kw)}},
		{"insertar justo al inicio lo desplaza", []syntax.Span{sp(4, 8, kw)}, 4, 0, 1, []syntax.Span{sp(5, 9, kw)}},
		{"insertar dentro lo extiende", []syntax.Span{sp(0, 10, str)}, 5, 0, 1, []syntax.Span{sp(0, 11, str)}},
		{"borrar antes desplaza hacia atrás", []syntax.Span{sp(6, 10, kw)}, 0, 2, -2, []syntax.Span{sp(4, 8, kw)}},
		{"borrar dentro lo acorta", []syntax.Span{sp(0, 10, str)}, 3, 2, -2, []syntax.Span{sp(0, 8, str)}},
		{"borrar todo el tramo lo elimina", []syntax.Span{sp(2, 5, kw)}, 2, 3, -3, nil},
		{"borrar solapando el inicio", []syntax.Span{sp(4, 8, kw)}, 2, 4, -4, []syntax.Span{sp(2, 4, kw)}},
		{"dos tramos", []syntax.Span{sp(0, 4, kw), sp(6, 9, str)}, 5, 0, 2, []syntax.Span{sp(0, 4, kw), sp(8, 11, str)}},
		{"sin tramos", nil, 3, 0, 1, nil},
		{"hasta el final recorta", []syntax.Span{sp(0, 4, kw), sp(6, 9, str), sp(10, 12, kw)}, 7, truncateAll, 0, []syntax.Span{sp(0, 4, kw), sp(6, 7, str)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := shiftSpans(tc.in, tc.off, tc.del, tc.delta)
			if !reflect.DeepEqual(got, tc.want) && !(len(got) == 0 && len(tc.want) == 0) {
				t.Errorf("obtenido %+v, quería %+v", got, tc.want)
			}
		})
	}
}

// El desplazamiento de tramos debe dejar los colores donde estaban respecto del texto: se compara con retokenizar.
func TestShiftMatchesRetokenizeForSimpleEdits(t *testing.T) {
	src := "package a\n\nfunc f() {\n\tx := \"hola\" + 12\n\treturn x\n}\n"
	for _, tc := range []struct {
		name string
		edit func(b *buffer.Buffer)
	}{
		{"insertar línea vacía", func(b *buffer.Buffer) { b.Insert(buffer.Pos{Line: 1}, "\n") }},
		{"insertar línea al final", func(b *buffer.Buffer) { b.Insert(buffer.Pos{Line: 5, Col: 1}, "\n\n") }},
		{"borrar una línea entera", func(b *buffer.Buffer) { b.Delete(buffer.Pos{Line: 1}, buffer.Pos{Line: 2}) }},
		{"escribir al inicio de línea", func(b *buffer.Buffer) { b.Insert(buffer.Pos{Line: 2}, "  ") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := buffer.FromString("a.go", src)
			bs := &bufState{buf: b}
			bs.initHighlight()
			old := bs.hl.lines
			if old == nil {
				t.Fatal("no se tokenizó")
			}
			tc.edit(b)
			if len(bs.hl.lines) < b.LineCount() {
				t.Fatalf("hay %d tramos para %d líneas", len(bs.hl.lines), b.LineCount())
			}
			fresh, _ := syntax.Tokenize("Go", lfText(b))
			// las líneas que no son la editada y que no están al final de la edición deben coincidir con el resultado real
			var mismatches int
			for i := 0; i < b.LineCount(); i++ {
				got, want := bs.hl.lines[i], fresh[i]
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !reflect.DeepEqual(got, want) {
					mismatches++
				}
			}
			if mismatches > 2 { // como mucho las 1-2 líneas tocadas por la edición quedan sin color hasta retokenizar
				t.Errorf("%d líneas con colores desplazados incorrectamente\ndesplazado: %v\nreal: %v", mismatches, bs.hl.lines, fresh)
			}
		})
	}
}

func TestShiftGivesUpOnInconsistentState(t *testing.T) {
	b := buffer.FromString("a.go", "x\n")
	h := &hlState{lines: [][]syntax.Span{nil}}
	h.shift(b, buffer.Edit{Start: buffer.Pos{Line: 0}, Deleted: strings.Repeat("\n", 5)})
	if h.lines != nil {
		t.Error("si la edición no cuadra con los tramos, se descartan en lugar de dar colores falsos")
	}
	h.shift(b, buffer.Edit{Start: buffer.Pos{}, Inserted: "x"}) // sin tramos: no hace nada ni falla
}

func TestLfTextIgnoresCRLF(t *testing.T) {
	b := buffer.FromString("", "a\r\nb\r\n")
	if got := lfText(b); got != "a\nb" {
		t.Errorf("lfText = %q", got)
	}
}
