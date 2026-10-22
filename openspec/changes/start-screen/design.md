# Design

## Context

- `ui.RenderEditor` (`internal/ui/editor_view.go`) dibuja el editor sobre una rejilla de celdas (`grid`) con el tema (`Theme`). Cuando el buffer es corto, las filas sobrantes llevan `~`.
- Con `fos carpeta/` o al cerrar el último buffer, el editor queda con un buffer vacío sin nombre (`[Sin nombre]`). Ese es el estado que activa la pantalla.
- Los atajos ya existen en el registro de comandos y en el mapa de teclas (`internal/editor/commands.go`). Ojo: `Space q` cierra el buffer y `Space Q` sale de la aplicación; la pantalla muestra `Space Q` para "Salir".
- `Areas(w, explorer, markdown)` da la zona del editor (`px`, `ew`). La pantalla se centra dentro de esa zona, no de toda la terminal.

## Goals / Non-Goals

**Goals:**
- Dibujar logo y lista centrados, con los colores del tema, en poco código y sin dependencias.
- Que la decisión de mostrarla viva en un solo sitio y sea fácil de probar.
- Que los atajos mostrados no puedan quedar desfasados sin que un test falle.

**Non-Goals:**
- Interacción (ratón, flechas), configuración, animación, logo desde archivo.

## Decisions

1. **Nuevo archivo `internal/ui/start_view.go`.** Tiene el logo como `[]string` (letras de bloque con `█`, 5 filas de alto, escritas a mano), la lista de `StartEntry{Label, Keys}` y `drawStart(g, x, y, w, h, th)`. Alternativa descartada: generar el logo con una librería de fuentes (dependencia nueva, prohibida sin permiso).
2. **`EditorView` gana un campo `Start bool`.** `RenderEditor` lo mira: si es `true` y el área del editor es válida, llama a `drawStart` en lugar de pintar los `~` y las líneas. Así `ui` solo dibuja y no decide nada.
3. **`Model` decide cuándo.** En `internal/editor/model.go`, al construir la vista: `Start = buf.Path()=="" && !buf.Modified() && buf vacío && no hay vista previa Markdown`. Se calcula en cada `View`, sin estado extra, así desaparece y vuelve sola.
4. **Tamaños con tres niveles.** El área necesita `logoRows + 1 + len(entries)` filas para el nivel completo; `len(entries)` para solo la lista; si no cabe la lista (o no cabe su anchura), no se dibuja nada y queda el editor vacío normal. El ancho se mide con el ancho visual de la celda, como el resto del código (`textWidth`).
5. **Atajos como texto fijo en `ui`, con un test que vigila el desfase.** Un test en `internal/editor` recorre `ui.StartEntries` y comprueba que cada atajo está asignado en el mapa de teclas y al comando esperado. Alternativa descartada: leerlos del registro en cada dibujo (más código, y `ui` no debe conocer `editor`).
6. **Sin cursor y sin ratón.** `CursorCell` devuelve `false` cuando `Start` está activo. El manejo del ratón (`mouse.go`) no se toca: las filas no son zona clicable.
7. **Colores:** logo en `th.Brand` en negrita; etiquetas en `th.Blue`; atajos en `th.Peach` (igual que la referencia de LazyVim); fondo `th.BG`.

## Risks / Trade-offs

- **Ancho de caracteres:** `█` es de ancho 1 en la mayoría de terminales; en algunas con ancho ambiguo puede salir doble. Mitigación: se usa solo `█` y espacios, y el logo se centra con margen de sobra.
- **Cursor y números de línea:** con la pantalla activa no hay números ni cursor; al pulsar `i` el buffer pasa a tener contenido y la pantalla desaparece. Hay un instante en que el buffer sigue vacío tras pulsar `i` (aún sin escribir): se acepta que la pantalla siga, y por eso `CursorCell` no debe mostrar el cursor solo si `Start` está activo; en modo Insert se oculta la pantalla (ver tareas 2.3).
- **Choque de teclas:** ninguno, no se añade ninguna tecla.
- **Rendimiento:** dibujar ~13 filas de texto fijo por fotograma es despreciable.
