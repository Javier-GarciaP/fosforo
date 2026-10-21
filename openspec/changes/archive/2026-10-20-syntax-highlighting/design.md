# Design

## Context

La vista (`internal/ui`) pinta texto plano por líneas visibles. Hace falta que reciba, por línea, una lista de tramos (texto + estilo).
Hay que decidir el motor de tokenización y cómo no penalizar la escritura.

## Goals / Non-Goals

**Goals:**
- Tokenización correcta de construcciones multilínea con un motor probado.
- Cero bloqueo del bucle de Bubble Tea.

**Non-Goals:**
- Incrementalidad fina por nodo (tree-sitter).

## Decisions

- **Motor: Chroma.** Es Go puro (sin cgo), tiene decenas de lenguajes y se distribuye como binario único. Alternativa: tree-sitter
  (más preciso e incremental, pero cgo complica compilación cruzada y empaquetado). Queda como mejora futura detrás de la misma interfaz.
- **Interfaz:** `syntax.Highlighter.Tokenize(lang, text) [][]Span` donde `Span{Start, End, Kind}`; `Kind` es una categoría abstracta
  (keyword, type, function, string, number, comment, ...) que el tema traduce a color. Así cambiar de tema no retokeniza.
- **Asincronía:** al cambiar el buffer se envía un `tea.Cmd` con (versión, texto) con *debounce* de ~50 ms; el resultado vuelve como mensaje
  y se descarta si la versión ya es antigua. Mientras tanto se usan los tramos anteriores, desplazados por línea cuando se insertan/borran líneas.
- **Tokenizar todo el texto, pintar lo visible:** Chroma mantiene estado entre líneas, por lo que se tokeniza el buffer completo (≤ 2 MiB)
  y la vista solo consulta las líneas visibles.
- **Mapa de categorías → colores:** definido por tema (ver guía de estilo); un único lugar para ajustar la paleta.

- **Decisiones tomadas al implementar:**
  - `Span` usa **desplazamientos en bytes** dentro de la línea y no se guardan los tramos sin categoría (`Plain`); los tramos contiguos de la misma
    categoría se funden. La vista recorre los grafemas llevando la cuenta de bytes, así los caracteres de varios bytes y el scroll horizontal no desalinean colores.
  - `buffer` ofrece `Rev()` (contador de cambios que también crece al deshacer/rehacer) y `OnChange(func(Edit))`. Con ello el editor sabe si un resultado
    asíncrono sigue siendo válido y desplaza los tramos con cada edición (`hlState.shift`): una edición de una línea reajusta los desplazamientos; una de varias
    líneas conserva el principio de la primera y el final de la última, y deja sin color las líneas nuevas del medio hasta retokenizar.
  - Los textos de hasta 100 KiB se resaltan **al cargarlos** (síncrono, rápido); los demás y todas las ediciones pasan por `tea.Tick` (50 ms de debounce) y un
    `tea.Cmd` que tokeniza una copia del texto. Solo corre una tokenización a la vez y el resultado se descarta si la revisión ya cambió.
  - Rendimiento medido: Chroma tokeniza ~0,6 MB/s (≈3 s para 2 MiB). Por eso el límite de 2 MiB sigue siendo asíncrono y no bloquea; si hiciera falta, se podría
    bajar el límite o sustituir Chroma por tree-sitter detrás de la misma interfaz.
  - Las líneas de más de 2000 bytes se tokenizan como vacías (para no frenar el analizador y conservar el número de líneas) y se muestran sin color.
  - Detección: `go.mod`/`go.work` se resaltan como Go y `go.sum` queda plano (Chroma asociaba `.mod` a otro lenguaje); las extensiones no distinguen mayúsculas.
  - Aproximación conocida: Chroma no distingue algunos identificadores (p. ej. el nombre de un tipo declarado por el usuario se ve como texto normal).
  - El tema claro se añadió aquí (`ui.Light()`, paleta de `docs/design/src/palette.js`) junto al comando `theme.toggle` (`Space t`); alternar solo cambia los
    colores con que se pintan las categorías.

## Risks / Trade-offs

- [Retokenizar todo en cada edición es caro en archivos de ~2 MiB] → debounce, cancelación de trabajos antiguos y límite de tamaño; medir antes de optimizar.
- [Los lexers de Chroma difieren de lo esperable en algunos lenguajes] → aceptar aproximación; mapa de categorías ajustable.
- [Desfase visual momentáneo tras editar] → desplazar tramos antiguos y refrescar al llegar el resultado.

## Open Questions

- Lista inicial de lenguajes verificados con pruebas golden (mínimo: Go, Markdown, JSON, YAML, TOML, JS/TS, Python, Rust, shell).
