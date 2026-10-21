# Design

## Context

Parte de `bootstrap-project` (esqueleto y arranque). El modelo de teclas modal y el registro de comandos llegan en
`modal-editing`, por lo que aquí los atajos son provisionales pero la API de edición ya debe ser la definitiva.

## Goals / Non-Goals

**Goals:**
- API de `buffer` independiente de la UI y de Bubble Tea, probable con pruebas de tabla.
- Render puro: estado → cadena, verificable con pruebas "golden".

**Non-Goals:**
- Estructuras para archivos enormes (rope/piece table), ajuste de línea, múltiples cursores.

## Decisions

- **Almacenamiento:** `[]string` de líneas. Es simple y suficiente hasta ~10 MiB (límite de solo lectura). Si el rendimiento lo
  exige, se cambia por rope detrás de la misma interfaz de `buffer`. Alternativa: piece table desde el inicio (más compleja sin necesidad probada).
- **Operaciones atómicas:** toda edición pasa por `Apply(Edit)` donde `Edit = {inicio, texto eliminado, texto insertado}`.
  Esto da undo/redo gratis en `modal-editing` y mantiene un único punto donde se marca "modificado".
- **Posiciones:** el cursor guarda (línea, índice de grafema); la columna visual se calcula con `uniseg` + tabs a 4. Se cachea el
  "columna deseada" para el movimiento vertical.
- **Render:** `ui.RenderEditor(state, size) string` sin efectos; Bubble Tea solo orquesta (`editor.Model.Update/View`).
- **Guardado:** archivo temporal en el mismo directorio + `rename`; `fsync` antes del rename. Se conservan permisos y, si es posible, propietario.
- **Solo lectura:** se decide al cargar (NUL, UTF-8 inválido, >10 MiB) y se expone como propiedad del buffer.

- **Posición y vista (decisiones tomadas al implementar):** el cursor vive en `buffer.Cursor` (con la columna visual deseada); la vista
  (`top`, `left`) vive en `editor.Model` y se ajusta tras cada mensaje. Al volver atrás horizontalmente el cursor se centra en
  lugar de quedar pegado al borde. Las filas se dibujan sobre una cuadrícula de celdas (`ui/grid.go`), lo que permite superponer
  el diálogo y recortar caracteres anchos sin romper la alineación.
- **Seguridad de dibujo:** los caracteres de control del archivo (ESC, BEL, NUL, C1) se muestran como `·` y los bytes UTF-8 inválidos como `�`,
  de modo que el contenido de un archivo nunca envía secuencias al terminal.
- **Salida transitoria:** `Ctrl+Q` y `Ctrl+C` salen (con diálogo si hay cambios); `q` pasa a escribir una letra y se retira de `app-startup`.
- **Statusline estrecha:** la ruta se recorta por la izquierda con «…» para que `[+]` y `[solo lectura]` siempre se vean; por debajo de 60 columnas se omite `utf-8`/`LF`.
- **Pruebas del pánico sin `-race`:** Bubble Tea tiene una carrera interna al cerrar la entrada tras un pánico; esa prueba se ejecuta aparte sin `-race`.

## Risks / Trade-offs

- [Errores de cursor con Unicode] → pruebas con CJK, emoji compuestos, combinantes y tabs desde el primer día.
- [Redibujar todo en cada tecla] → renderizar solo filas visibles; medir antes de optimizar.
- [Perder datos al guardar] → guardado atómico y pruebas de fallo (directorio sin permisos, disco lleno simulado).
- [Atajos transitorios `Ctrl+S` / `Ctrl+Q` chocan con el terminal (XOFF/XON)] → desactivar flow control en el modo raw; se reemplazan en `modal-editing`.

## Open Questions

- Tabulación: ¿insertar `\t` real o espacios por defecto? Se asume `\t` real con ancho 4 hasta que `config-search-release` lo haga configurable.
