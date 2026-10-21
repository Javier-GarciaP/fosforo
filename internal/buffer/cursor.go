package buffer

// Cursor es una posición en el buffer más la columna visual deseada para el movimiento vertical.
// Al subir o bajar por líneas de distinta longitud el cursor recuerda la columna original y la recupera
// cuando la línea lo permite.
type Cursor struct {
	Pos
	want int // columna visual deseada; -1 si no hay ninguna fijada
}

// NewCursor devuelve un cursor en p.
func NewCursor(p Pos) Cursor { return Cursor{Pos: p, want: -1} }

// Set coloca el cursor en p (acercándolo a una posición válida) y olvida la columna deseada.
func (c *Cursor) Set(b *Buffer, p Pos) {
	c.Pos, c.want = b.Clamp(p), -1
}

// VisualCol devuelve la columna de pantalla del cursor.
func (c Cursor) VisualCol(b *Buffer) int { return VisualCol(b.Line(c.Line), c.Col) }

// Left mueve un carácter a la izquierda, pasando al final de la línea anterior desde el inicio de línea.
func (c *Cursor) Left(b *Buffer) {
	if p, ok := b.Prev(c.Pos); ok {
		c.Set(b, p)
	}
}

// Right mueve un carácter a la derecha, pasando a la línea siguiente desde el final de línea.
func (c *Cursor) Right(b *Buffer) {
	if p, ok := b.Next(c.Pos); ok {
		c.Set(b, p)
	}
}

// Home mueve al inicio de la línea.
func (c *Cursor) Home(b *Buffer) { c.Set(b, Pos{c.Line, 0}) }

// End mueve al final de la línea.
func (c *Cursor) End(b *Buffer) { c.Set(b, Pos{c.Line, b.LineLen(c.Line)}) }

// Vertical mueve n líneas (negativo sube) conservando la columna visual deseada. Se detiene en la primera y última línea.
func (c *Cursor) Vertical(b *Buffer, n int) {
	if c.want < 0 {
		c.want = c.VisualCol(b)
	}
	line := max(0, min(c.Line+n, b.LineCount()-1))
	c.Pos = Pos{line, ColAt(b.Line(line), c.want)}
}

// Up sube una línea.
func (c *Cursor) Up(b *Buffer) { c.Vertical(b, -1) }

// Down baja una línea.
func (c *Cursor) Down(b *Buffer) { c.Vertical(b, 1) }

// PageUp sube n líneas.
func (c *Cursor) PageUp(b *Buffer, n int) { c.Vertical(b, -n) }

// PageDown baja n líneas.
func (c *Cursor) PageDown(b *Buffer, n int) { c.Vertical(b, n) }
