# Proposal

## Why

Un editor de proyectos necesita navegar archivos y tener varios abiertos a la vez. Con un solo buffer, abrir otro archivo
descarta el trabajo; sin explorador hay que escribir rutas. Ambos se diseñan juntos porque el explorador es la principal
fuente de buffers nuevos.

## What Changes

- Panel lateral con el árbol de carpetas (ver `docs/design/exports/01-layout-principal-normal.png` y `07`), alternable con `Space e`.
- Navegación y operaciones de archivo desde el árbol: abrir, expandir/colapsar, crear, renombrar y borrar.
- Foco explícito entre explorador y editor.
- Lista de buffers abiertos con cambio, cierre y protección ante cambios sin guardar.
- Detección de archivos modificados fuera del editor.

## Capabilities

### New Capabilities
- `file-explorer`: árbol de archivos, navegación y operaciones de archivo.
- `buffer-management`: varios buffers abiertos, cambio entre ellos, cierre y detección de cambios externos.

### Modified Capabilities

Ninguna en `openspec/specs/`. Amplía `editor-view` (anchura del panel y barra de estado con nº de buffers); ese cambio se documenta en el diseño
y en los requisitos nuevos, sin alterar los requisitos de `editor-view`.

## Non-goals

- Pestañas visuales (solo lista de buffers por comandos y, luego, por el picker).
- Monitoreo del sistema de archivos en tiempo real: v1 refresca al guardar, al volver el foco y con `R`.
- Operaciones por lotes, arrastrar y soltar, vista de git.

## Impact

- Código nuevo: `internal/fs`, vista del explorador en `internal/ui`; `internal/editor` gestiona lista de buffers y foco.
- Depende de: `modal-editing` (comandos y keymap por contexto).
