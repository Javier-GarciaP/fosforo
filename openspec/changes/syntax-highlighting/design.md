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

## Risks / Trade-offs

- [Retokenizar todo en cada edición es caro en archivos de ~2 MiB] → debounce, cancelación de trabajos antiguos y límite de tamaño; medir antes de optimizar.
- [Los lexers de Chroma difieren de lo esperable en algunos lenguajes] → aceptar aproximación; mapa de categorías ajustable.
- [Desfase visual momentáneo tras editar] → desplazar tramos antiguos y refrescar al llegar el resultado.

## Open Questions

- Lista inicial de lenguajes verificados con pruebas golden (mínimo: Go, Markdown, JSON, YAML, TOML, JS/TS, Python, Rust, shell).
