# Proposal

## Why

Antes de empezar a implementar hay que fijar cómo se ejecuta el programa y bajo qué licencia se publica. El nombre de marca
*fosforo* es largo para escribirlo decenas de veces al día en una terminal, y sin licencia nadie puede usar ni contribuir al
código legalmente.

## What Changes

- El comando que se escribe en la terminal pasa a ser **`fos`**; *fosforo* se mantiene como nombre comercial, de marca y del repositorio.
- Las specs y la documentación que nombraban el comando `fosforo` se actualizan a `fos` (binario, directorio `cmd/fos`, `--version`,
  `go install`, ruta de configuración `~/.config/fos/`).
- Se publica el proyecto bajo licencia **MIT**: archivo `LICENSE`, insignia y sección del README.
- Verificación previa de que `fos` no choca con herramientas de uso común del ecosistema de desarrollo.

## Capabilities

### New Capabilities
- `command-name`: nombre del comando ejecutable y su relación con el nombre de marca.
- `licensing`: licencia del proyecto y su declaración en el repositorio.

### Modified Capabilities

Ninguna publicada todavía. Los cambios pendientes `bootstrap-project`, `core-text-editing`, `markdown-preview` y
`config-search-release` se actualizan en su texto para usar `fos`.

## Non-goals

- Crear un alias `fosforo` en el binario instalado (se puede añadir más adelante si hace falta).
- Renombrar el repositorio, el logo o la marca.
- Escribir código.

## Impact

- Documentación y specs únicamente: `openspec/changes/*`, `docs/README.md`, `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `LICENSE`.
