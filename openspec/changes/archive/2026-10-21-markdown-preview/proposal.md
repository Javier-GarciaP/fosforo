# Proposal

## Why

Es un requisito explícito del producto: los archivos Markdown deben abrirse en una vista renderizada "como en GitHub", no como
texto con símbolos. Quien documenta con README y notas necesita leer cómodamente y alternar con la edición sin salir del editor.

## What Changes

- Los archivos `.md` y `.markdown` se abren por defecto en **vista previa** renderizada (ver `docs/design/exports/07-markdown-vista-previa.png`).
- Alternar entre vista previa y fuente, y modo **dividido** fuente | vista previa con scroll sincronizado (ver `08-markdown-dividido.png`).
- Renderizado con estilo tipo GitHub en tema oscuro y claro: encabezados, listas, tareas, citas, código con resaltado, tablas, enlaces y reglas.
- Ajuste automático al ancho del panel y recálculo al editar o redimensionar.

## Capabilities

### New Capabilities
- `markdown-preview`: renderizado de Markdown en el editor, navegación de la vista previa y su relación con la edición del fuente.

### Modified Capabilities

Ninguna.

## Non-goals

- Imágenes reales, Mermaid, HTML embebido y matemáticas: se muestran como texto alternativo o se omiten (limitación de la terminal).
- Seguir enlaces, abrir URL en el navegador o navegar entre documentos.
- Exportar a HTML/PDF.
- Fidelidad pixel a pixel con GitHub: el objetivo es la misma estructura y jerarquía visual.

## Impact

- Código nuevo: `internal/markdown`; vista de preview en `internal/ui`; nuevos comandos `markdown.*`.
- Dependencias: Glamour (sobre goldmark) con estilos propios, reutiliza `syntax-highlighting` para bloques de código.
- Depende de: `core-text-editing`, `file-explorer-buffers` (apertura y buffers) y `syntax-highlighting`. No depende del modo Vim más allá de los comandos.
