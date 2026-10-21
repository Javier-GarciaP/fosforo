# Proposal

## Why

Una terminal no puede saber si tiene una Nerd Font. Con los iconos activados por defecto, quien no la tiene ve cuadros vacíos que además pisan la primera
letra de cada nombre (se comprobó probando la app). El explorador ya es plenamente usable sin iconos, así que el valor por defecto más seguro es no mostrarlos.

## What Changes

- El valor por defecto de `icons` pasa de `nerd` a `none` (configuración y `--icons`); `nerd` sigue disponible para quien tenga la fuente.
- Se actualizan la plantilla de configuración y la documentación.

## Capabilities

### New Capabilities

Ninguna.

### Modified Capabilities
- `user-config`: el valor por defecto de `icons` es `none`.

## Non-goals

- Detectar automáticamente la Nerd Font (no es fiable) o quitar el modo `nerd`.

## Impact

`internal/config`, `internal/cli`, plantilla de configuración y documentación.
