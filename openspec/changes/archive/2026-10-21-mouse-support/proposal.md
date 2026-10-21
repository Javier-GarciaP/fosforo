# Proposal

## Why

El teclado es la vía principal, pero el mouse debe funcionar de forma natural para quien lo prefiera en ciertas tareas
(posicionar el cursor, desplazarse, abrir archivos). Es un requisito explícito del producto y no debe duplicar lógica.

## What Changes

- Clic para mover el cursor, arrastrar para seleccionar (entra en Visual), doble clic para seleccionar una palabra.
- Rueda para desplazar el panel bajo el puntero (editor, explorador, vista previa, listas de picker/paleta).
- Clic en el explorador para abrir archivos y expandir directorios; clic en elementos de picker, paleta y diálogos.
- Todas las acciones del mouse se traducen a comandos del registro (sin lógica paralela).

## Capabilities

### New Capabilities
- `mouse-interaction`: reglas de cómo responde cada zona de la interfaz al mouse.

### Modified Capabilities

Ninguna.

## Non-goals

- Arrastrar y soltar archivos, menús contextuales, redimensionar paneles arrastrando.
- Clic derecho y botón central.
- Soporte de mouse en terminales sin protocolo SGR.

## Impact

- Código nuevo: geometría compartida de paneles y búsqueda de zona bajo el puntero en `internal/ui`; nuevos comandos (`cursor.move_to`, `view.scroll`, `visual.select_range`).
- Depende de: `modal-editing`, `file-explorer-buffers`, `picker-and-palette` y, para la rueda en preview, `markdown-preview`.
