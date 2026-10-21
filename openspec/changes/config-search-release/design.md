# Design

## Context

Con las funciones de v1 completas, falta personalización, búsqueda en el proyecto y un proceso de publicación coherente con el flujo
`dev` → `main`.

## Goals / Non-Goals

**Goals:**
- Configuración tolerante a errores y fácil de depurar.
- Releases reproducibles y atados a `main`.

**Non-Goals:**
- Sistema de plugins o scripting.

## Decisions

- **Formato: TOML** (legible, comentarios, sin ambigüedades de YAML). Biblioteca `pelletier/go-toml/v2`; alternativa JSON descartada por no admitir comentarios.
- **Capas:** valores embebidos → archivo de usuario → opciones en tiempo de ejecución (`Space t`). La validación acumula errores por clave y nunca aborta.
- **Remapeo:** los atajos del usuario se aplican sobre el trie por defecto por contexto; un comando inexistente se rechaza con su nombre.
- **Tema automático:** consulta OSC 11 del fondo de la terminal con tiempo máximo corto; si no hay respuesta, tema oscuro.
- **Búsqueda:** `rg --json` por `os/exec` con lectura en streaming y cancelación por contexto; respaldo con `fs.Walk` y lectura por líneas con las mismas exclusiones.
- **Release:** GoReleaser (matriz Linux/macOS, amd64/arm64, checksums) lanzado por GitHub Actions al crear una etiqueta; una comprobación inicial exige que el commit
  etiquetado sea ancestro de `main`. Alternativa: script propio (más control, más mantenimiento).
- **Versión:** `-ldflags` inyecta versión y commit; `go install` sin etiqueta muestra `dev`.

## Risks / Trade-offs

- [Binarios de macOS descargados activan Gatekeeper] → documentar `xattr -d com.apple.quarantine` y valorar firma en el futuro.
- [`rg` y el respaldo devuelven resultados distintos] → documentar diferencias y probar ambos con el mismo proyecto de ejemplo.
- [Configuración rota tras una actualización] → claves desconocidas solo avisan; versionado semántico y notas de cambios.
- [OSC 11 no soportado o con retraso] → tiempo máximo corto y valor por defecto oscuro.

## Open Questions

- Alias corto del comando (p. ej. `fos`) y ruta del módulo: pendientes de crear el repositorio.
