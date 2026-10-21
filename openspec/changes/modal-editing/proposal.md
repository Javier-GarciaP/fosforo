# Proposal

## Why

La propuesta de valor del editor es la fluidez de Neovim: edición modal, secuencias de teclas y comandos combinables.
Es lo que más condiciona la arquitectura (todo debe ser un comando con nombre) y lo más caro de rehacer, así que se
construye justo después del núcleo de edición y antes de explorador, paleta o mouse, que dependen de él.

## What Changes

- Registro único de comandos con nombre (`file.save`, `cursor.down`, ...) usado por teclas, `:`, paleta y mouse.
- Modos Normal (por defecto), Insert, Visual (carácter y línea) y Command; etiqueta y color por modo (ver `docs/design/exports/00-guia-de-estilo.png`, `02`, `03`).
- Resolución de secuencias de teclas (`dd`, `gg`, `Space f f`) con árbol de prefijos, tiempo de espera y conteos (`3dd`).
- Motions, operadores (`d`, `c`, `y`), pegar, deshacer/rehacer y búsqueda en el archivo (`/`, `n`, `N`).
- Línea de comandos `:` con `:w`, `:q`, `:q!`, `:wq`, `:e <ruta>` y `:<n>`.
- Reemplazo de los atajos transitorios `Ctrl+S` / `Ctrl+Q` de `core-text-editing`.

## Capabilities

### New Capabilities
- `command-registry`: catálogo de comandos con nombre, descripción y atajos por defecto.
- `modal-input`: modos, resolución de secuencias, conteos y tiempos de espera.
- `editing-commands`: motions, operadores, registros, deshacer/rehacer, modo Visual y búsqueda en archivo.
- `ex-commands`: línea de comandos `:` y sus comandos.

### Modified Capabilities

Ninguna en `openspec/specs/`. Este cambio sustituye el requisito transitorio "Transitional save and quit" de `core-text-editing`
(`Ctrl+Q` pasa a ser `:q`/`:wq`; `Ctrl+S` se conserva, ver `modal-input`). Esa spec aún no está archivada, por lo que
`core-text-editing` debe archivarse antes de implementar este cambio.

## Non-goals

- Repetir con `.`, objetos de texto (`iw`, `i"`), macros, registros con nombre y marcas.
- Configuración de teclas por el usuario (`config-search-release`).
- Menú líder visual y paleta (`picker-and-palette`), aunque el árbol de prefijos ya admite `Space`.

## Impact

- Código nuevo: `internal/command`, `internal/keymap`; cambios grandes en `internal/editor` y `internal/buffer` (historial de deshacer).
- Depende de: `core-text-editing`.
