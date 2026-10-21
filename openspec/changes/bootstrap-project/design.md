# Design

## Context

Proyecto nuevo, sin código. Stack acordado: Go + Bubble Tea (ver `docs/README.md`). Este cambio solo prepara el terreno;
las decisiones de arquitectura de fondo (buffer, comandos, keymap) viven en los cambios siguientes.

## Goals / Non-Goals

**Goals:**
- Que `go build ./...`, `go vet ./...` y `go test ./...` funcionen desde el primer commit y corran en CI.
- Probar el ciclo de vida de la terminal de forma automática.

**Non-Goals:**
- Diseñar la UI o el modelo de estado del editor.

## Decisions

- **Disposición del repo:** `cmd/fos/` para el binario e `internal/` para todo lo demás (nada se publica como librería).
  Alternativa descartada: `pkg/`, porque no hay consumidores externos.
- **Versiones de dependencias:** al empezar se verifica la versión vigente de Bubble Tea, Lip Gloss y Bubbles (la línea v2
  cambió las rutas de importación) y se fija en `go.mod`. No se mezclan líneas v1 y v2.
- **Pantalla alternativa y restauración:** Bubble Tea con pantalla alternativa; `main` envuelve la ejecución con
  `recover` y manejo de señales para restaurar la terminal antes de salir con error.
- **Pruebas:** `teatest` para el modelo y pruebas de tabla para el parseo de flags; el ciclo de vida se prueba lanzando
  el binario compilado en un pseudo-terminal.
- **CI:** GitHub Actions con `go vet`, `go test ./...` y `go build ./...` en cada PR a `dev` y a `main`.
- **Nombre:** el producto se llama *fosforo* (nombre comercial y de marca) y el comando es `fos` (binario `fos`, directorio `cmd/fos`). El nombre del comando vive en una sola constante para no esparcirlo por el código.

## Risks / Trade-offs

- [Restauración incompleta en casos raros (SIGKILL no se puede interceptar)] → documentarlo; `reset` queda como salida de emergencia.
- [Prueba en pseudo-terminal frágil entre sistemas] → aislarla en un test marcado y mantener la lógica probable con `teatest`.
- [Versión de Go o de Bubble Tea distinta entre máquinas] → fijar `go` y `toolchain` en `go.mod`.

## Open Questions

- Ruta del módulo: `github.com/Javier-GarciaP/fosforo` (el repositorio conserva el nombre de marca).
