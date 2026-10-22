# Proposal

## Why

Hoy, al abrir un proyecto con `fos .` (o al cerrar el último archivo), la zona del editor queda como un buffer vacío lleno de `~`. No dice qué hacer primero ni enseña los atajos. Una pantalla de inicio con la marca FOSFORO y los atajos principales hace la primera impresión más clara y ayuda a aprender la tecla líder `Space`.

## What Changes

- Nueva pantalla de inicio en el área del editor: logo "FOSFORO" en letras grandes de bloque (color `Brand` del tema Tinta) y, debajo, una lista centrada de acciones con su atajo a la derecha.
- Solo se muestra cuando el buffer activo está vacío, sin nombre y sin modificar. Desaparece en cuanto se abre un archivo o se escribe texto.
- Es solo informativa: no se puede pulsar con el ratón.
- Si la ventana es pequeña, el logo se oculta primero y después la lista; el editor sigue funcionando igual.
- No cambia ningún atajo ni comando existente.

## Capabilities

### New Capabilities
- `start-screen`: pantalla de inicio con logo y atajos cuando no hay ningún archivo abierto.

### Modified Capabilities

(ninguna: no cambia el comportamiento de `editor-view`; la pantalla de inicio es una capacidad nueva)

## Non-goals

- Filas con ratón o selección con las flechas (se puede hacer en otro cambio).
- Teclas sueltas tipo LazyVim (`f`, `n`, `r`): chocan con los comandos Normal.
- Mostrar plugins, tiempo de carga o lista de archivos recientes dentro de la pantalla.
- Opción de configuración para desactivarla (se decidirá si hace falta más adelante).
- Logo animado o cargado desde un archivo externo.

## Impact

- Código: `internal/ui` (dibujo del logo y la lista, nuevo campo en `EditorView`) e `internal/editor/model.go` (decidir cuándo mostrarla).
- Docs: `docs/README.md` (tabla de funciones) y `CHANGELOG.md`.
- Sin dependencias nuevas. Diseño visual de referencia: la marca de `docs/design/exports/12-marca-fosforo.png` y la captura del dashboard de LazyVim.
