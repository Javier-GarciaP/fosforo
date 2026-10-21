package buffer_test

import (
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

func TestVisualColAndColAt(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		col  int
		want int
	}{
		{"ascii", "abc", 2, 2},
		{"tab al inicio", "\tx", 1, 4},
		{"tab tras una letra", "a\tx", 2, 4},
		{"tab tras tres letras", "abc\tx", 4, 4},
		{"tab tras cuatro letras", "abcd\tx", 5, 8},
		{"CJK ocupa dos", "日本x", 2, 4},
		{"antes del CJK", "日本x", 1, 2},
		{"emoji compuesto ocupa dos", "👨‍👩‍👧x", 1, 2},
		{"combinante no suma", "éx", 1, 1},
		{"control ocupa una", "a\x1bb", 2, 2},
		{"más allá del final", "ab", 5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := buffer.VisualCol(tc.line, tc.col); got != tc.want {
				t.Errorf("VisualCol(%q,%d) = %d, quería %d", tc.line, tc.col, got, tc.want)
			}
		})
	}

	// ColAt es la inversa: una columna visual dentro de un carácter ancho devuelve ese carácter
	for _, tc := range []struct {
		line string
		v    int
		want int
	}{
		{"日本x", 0, 0}, {"日本x", 1, 0}, {"日本x", 2, 1}, {"日本x", 3, 1}, {"日本x", 4, 2}, {"日本x", 9, 3},
		{"a\tb", 1, 1}, {"a\tb", 3, 1}, {"a\tb", 4, 2},
	} {
		if got := buffer.ColAt(tc.line, tc.v); got != tc.want {
			t.Errorf("ColAt(%q,%d) = %d, quería %d", tc.line, tc.v, got, tc.want)
		}
	}
}

func TestCursorHorizontalCrossesLines(t *testing.T) {
	b := buffer.FromString("", "ab\n日\n")
	c := buffer.NewCursor(buffer.Pos{0, 2})
	c.Right(b)
	if c.Pos != (buffer.Pos{1, 0}) {
		t.Fatalf("Right desde el final de línea: %+v", c.Pos)
	}
	c.Right(b)
	if c.Pos != (buffer.Pos{1, 1}) || c.VisualCol(b) != 2 {
		t.Fatalf("avanzar sobre un carácter ancho: lógica %+v, visual %d (debía avanzar 1 lógica y 2 celdas)", c.Pos, c.VisualCol(b))
	}
	c.Left(b)
	c.Left(b)
	if c.Pos != (buffer.Pos{0, 2}) {
		t.Fatalf("Left desde el inicio de línea: %+v", c.Pos)
	}
	c.Set(b, buffer.Pos{0, 0})
	c.Left(b)
	if c.Pos != (buffer.Pos{0, 0}) {
		t.Fatalf("Left en el inicio del texto: %+v", c.Pos)
	}
	c.Set(b, buffer.Pos{1, 1})
	c.Right(b)
	if c.Pos != (buffer.Pos{1, 1}) {
		t.Fatalf("Right en el final del texto: %+v", c.Pos)
	}
}

func TestCursorDesiredColumn(t *testing.T) {
	long := "0123456789012345678901234"
	b := buffer.FromString("", long+"\nabcde\n"+long+"\n")
	c := buffer.NewCursor(buffer.Pos{0, 20})
	c.Down(b)
	if c.Pos != (buffer.Pos{1, 5}) {
		t.Fatalf("a la línea corta: %+v, quería {1 5}", c.Pos)
	}
	c.Down(b)
	if c.Pos != (buffer.Pos{2, 20}) {
		t.Fatalf("de vuelta a la línea larga: %+v, quería {2 20}", c.Pos)
	}
	c.Up(b)
	c.Up(b)
	if c.Pos != (buffer.Pos{0, 20}) {
		t.Fatalf("subiendo: %+v", c.Pos)
	}
	// un movimiento horizontal olvida la columna deseada
	c.Down(b)
	c.Left(b)
	c.Down(b)
	if c.Pos.Col != 4 {
		t.Fatalf("tras Left la columna deseada debe reiniciarse: %+v", c.Pos)
	}
}

func TestCursorVerticalWithWideChars(t *testing.T) {
	b := buffer.FromString("", "日本語x\nabcdefgh\n")
	c := buffer.NewCursor(buffer.Pos{0, 2}) // columna visual 4
	c.Down(b)
	if c.Pos != (buffer.Pos{1, 4}) {
		t.Errorf("Down mantiene la columna visual, no la lógica: %+v", c.Pos)
	}
	c.Up(b)
	if c.Pos != (buffer.Pos{0, 2}) {
		t.Errorf("Up: %+v", c.Pos)
	}
}

func TestCursorLimits(t *testing.T) {
	b := buffer.FromString("", "a\nb\nc\nd\ne\n")
	c := buffer.NewCursor(buffer.Pos{})
	c.Up(b)
	if c.Line != 0 {
		t.Errorf("Up en la primera línea: %+v", c.Pos)
	}
	c.PageDown(b, 100)
	if c.Line != 4 {
		t.Errorf("PageDown se detiene en la última línea: %+v", c.Pos)
	}
	c.Down(b)
	if c.Line != 4 {
		t.Errorf("Down en la última línea: %+v", c.Pos)
	}
	c.PageUp(b, 3)
	if c.Line != 1 {
		t.Errorf("PageUp(3): %+v", c.Pos)
	}
	c.PageUp(b, 100)
	if c.Line != 0 {
		t.Errorf("PageUp se detiene en la primera línea: %+v", c.Pos)
	}
	c.End(b)
	if c.Col != 1 {
		t.Errorf("End: %+v", c.Pos)
	}
	c.Home(b)
	if c.Col != 0 {
		t.Errorf("Home: %+v", c.Pos)
	}
}
