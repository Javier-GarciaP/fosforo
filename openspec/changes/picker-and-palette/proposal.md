# Proposal

## Why

La fluidez prometida depende de llegar a cualquier archivo o comando con pocas teclas y sin memorizar todo. El picker difuso, la
paleta de comandos y el menú líder con ayuda visible son los tres mecanismos que lo logran, y comparten el mismo componente de lista.

## What Changes

- Componente de selección reutilizable (entrada + lista filtrada difusa + vista previa opcional), ver `docs/design/exports/05-picker-de-archivos.png`.
- Buscar archivos (`Space f f`) y buffers (`Space b b`).
- Paleta de comandos (`Space Space`) y sugerencias al escribir en `:` (ver `06-paleta-de-comandos.png`).
- Menú líder `Space` con ayuda visible (ver `04-menu-lider-space.png`).

## Capabilities

### New Capabilities
- `fuzzy-picker`: selección difusa de archivos y buffers con vista previa.
- `command-palette`: búsqueda y ejecución de comandos del registro con sus atajos.
- `leader-menu`: ayuda visible de las secuencias que empiezan por `Space`.

### Modified Capabilities

Ninguna.

## Non-goals

- Búsqueda de texto en el proyecto (`config-search-release`).
- Divisiones de ventana: el picker abre en el panel actual (sin `Ctrl+v` dividir).
- Historial persistente entre sesiones.

## Impact

- Código nuevo: componente `Picker` y capa de superposición en `internal/ui`, fuente de archivos en `internal/fs`, listado en `internal/command`.
- Dependencias: algoritmo de coincidencia difusa (p. ej. `sahilm/fuzzy` o implementación propia).
- Depende de: `modal-editing` (registro y keymap), `file-explorer-buffers` (lista de archivos y buffers), `syntax-highlighting` (vista previa coloreada).
