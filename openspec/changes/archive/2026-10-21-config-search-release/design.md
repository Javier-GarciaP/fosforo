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
- **Tema automático:** Bubble Tea pide el fondo (OSC 11) al arrancar y entrega `BackgroundColorMsg`; como la petición es asíncrona nunca bloquea el arranque y, si la terminal no responde, se queda el tema oscuro.
- **Búsqueda:** `rg --json` por `os/exec` con lectura en streaming y cancelación por contexto; respaldo con `fs.Walk` y lectura por líneas con las mismas exclusiones.
- **Release:** GoReleaser (matriz Linux/macOS, amd64/arm64, checksums) lanzado por GitHub Actions al crear una etiqueta; una comprobación inicial exige que el commit
  etiquetado sea ancestro de `main`. Alternativa: script propio (más control, más mantenimiento).
- **Versión:** `-ldflags` inyecta versión y commit; con `go install ...@vX.Y.Z` la versión sale de los metadatos de compilación; un binario de desarrollo muestra `dev`.
- **Recarga:** `config.reload` (y guardar el propio archivo de configuración) vuelve a aplicar todos los ajustes; los atajos se reconstruyen sobre una copia del trie de fábrica, así que un atajo rechazado nunca deja el trie a medias.
- **Búsqueda en la interfaz:** reutiliza la ventana de selección con un tipo de ventana de altura fija; lo escrito es la consulta (no un filtro difuso), se lanza 150 ms después de la última tecla y cada consulta nueva cancela la anterior. La consulta es literal con *smart-case*.

## Risks / Trade-offs

- [Binarios de macOS descargados activan Gatekeeper] → documentar `xattr -d com.apple.quarantine` y valorar firma en el futuro.
- [`rg` y el respaldo devuelven resultados distintos] → documentar diferencias y probar ambos con el mismo proyecto de ejemplo.
- [Configuración rota tras una actualización] → claves desconocidas solo avisan; versionado semántico y notas de cambios.
- [OSC 11 no soportado o con retraso] → tiempo máximo corto y valor por defecto oscuro.

## Open Questions

- Ninguna: el comando es `fos` y el módulo es `github.com/Javier-GarciaP/fosforo`.
