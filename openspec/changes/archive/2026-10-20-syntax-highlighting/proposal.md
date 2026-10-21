# Proposal

## Why

El color por sintaxis es lo que hace legible el código y lo que más separa un editor "serio" de un editor de texto plano.
Además lo reutilizan el picker (vista previa) y el Markdown (bloques de código), por lo que conviene resolverlo antes de ellos.

## What Changes

- Detección del lenguaje por extensión o nombre de archivo y resaltado con una paleta de colores por tema (ver `docs/design/exports/00-guia-de-estilo.png`).
- Tokenización asíncrona con caché por versión del buffer, para no bloquear la escritura.
- Composición correcta del resaltado con la línea actual, la selección Visual y el cursor.
- Límites de seguridad para archivos o líneas enormes.

## Capabilities

### New Capabilities
- `syntax-highlighting`: asignación de colores a los tokens del código según el lenguaje y el tema activo.

### Modified Capabilities

Ninguna.

## Non-goals

- Análisis semántico, plegado de código o sangría por árbol sintáctico (tree-sitter queda como mejora futura).
- Temas definidos por el usuario (`config-search-release`); aquí solo oscuro y claro predefinidos.

## Impact

- Código nuevo: `internal/syntax`; cambios en `internal/ui` para pintar tokens.
- Dependencias: Chroma.
- Depende de: `core-text-editing` (vista); aprovecha buffers de `file-explorer-buffers` si ya existe.
