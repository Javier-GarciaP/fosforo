# Proposal

## Why

El repositorio está vacío: no hay toolchain de Go, ni esqueleto compilable, ni convenciones de pruebas. Un editor de
terminal tiene un riesgo temprano muy concreto: si el programa sale mal (error, pánico, Ctrl+C) deja la terminal rota
(sin eco, sin cursor, en pantalla alternativa). Hay que resolverlo y probarlo antes de construir funcionalidades encima.

## What Changes

- Instalar Go y fijar la versión en `go.mod`.
- Crear el esqueleto del módulo: `cmd/ted/` y los paquetes vacíos previstos en `internal/` (ver `docs/README.md`).
- Un binario Bubble Tea mínimo que abre la pantalla alternativa, muestra un marcador y sale con `q` o `Ctrl+C`.
- Flags `--version` y `--help`; el argumento posicional `<ruta>` se acepta y se guarda (lo usan cambios posteriores).
- Restaurar siempre la terminal, también ante pánico o señal.
- `Makefile` (build, test, vet, run) y CI de GitHub Actions que valida cada PR a `dev`.

## Capabilities

### New Capabilities
- `app-startup`: ciclo de vida de la aplicación en la terminal: arranque, salida limpia, flags y tamaño mínimo.

### Modified Capabilities

Ninguna: no existen specs previas.

## Non-goals

- Cualquier edición de texto, apertura de archivos o UI real (cambio `core-text-editing` en adelante).
- Empaquetado y releases (cambio `config-search-release`).
- Decidir el nombre definitivo del producto: `ted` es provisional.

## Impact

- Código nuevo: `go.mod`, `cmd/ted/main.go`, `internal/*/doc.go`, `Makefile`, `.github/workflows/ci.yml`.
- Dependencias: Bubble Tea y, para pruebas, `teatest`.
- Sin impacto en datos ni APIs externas.
