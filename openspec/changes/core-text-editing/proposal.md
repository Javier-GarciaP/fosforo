# Proposal

## Why

Sin un buffer de texto fiable y una vista que lo dibuje no hay editor. Los errores de cursor con Unicode, tabs y
guardado son la fuente típica de pérdida de datos, así que se cimientan primero, con pruebas, antes del teclado modal.

## What Changes

- Buffer de texto en memoria: abrir, editar y guardar archivos UTF-8 con guardado atómico.
- Vista del editor: números de línea, scroll, relleno `~`, statusline y línea de comandos (ver `docs/design/exports/01-layout-principal-normal.png`).
- Escritura básica tipo modo Insert: insertar, borrar, saltos de línea, flechas, `Ctrl+S` guardar y `Ctrl+Q` salir (atajos transitorios que `modal-editing` sustituye, conservando el comportamiento).
- Abrir el archivo pasado como argumento (`fos <ruta>`).

## Capabilities

### New Capabilities
- `text-buffer`: carga, modificación, seguimiento de cambios y guardado seguro de texto.
- `editor-view`: representación en pantalla del buffer, scroll, statusline y redimensionado.
- `insert-typing`: edición directa de texto con teclado (comportamiento del futuro modo Insert).

### Modified Capabilities
- `app-startup`: la salida normal pasa de `q` a `Ctrl+Q`/`Ctrl+C`, porque `q` ahora escribe una letra (ver `specs/app-startup/spec.md`).

## Non-goals

- Modos Normal/Visual, motions, operadores y deshacer para el usuario (`modal-editing`).
- Resaltado de sintaxis, explorador, múltiples buffers (cambios posteriores).
- Ajuste de línea (word wrap): en v1 las líneas largas se desplazan horizontalmente.

## Impact

- Código nuevo: `internal/buffer`, `internal/editor`, `internal/ui`, `cmd/fos`.
- Dependencias: `rivo/uniseg` (y/o `go-runewidth`) para grafemas y ancho de celdas.
- Depende de: `bootstrap-project`.
