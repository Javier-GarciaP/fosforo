# Design

## Context

Hasta ahora `editor.Model` tiene un único buffer y un único foco. `modal-editing` ya aporta registro de comandos y keymap por modo.
Este cambio añade el concepto de **foco/contexto** (editor o explorador) y una **lista de buffers**.

## Goals / Non-Goals

**Goals:**
- Que el explorador reutilice el registro de comandos y el keymap (sus teclas son comandos como `explorer.open`).
- Mantener `internal/fs` sin dependencias de UI para probarlo con directorios temporales.

**Non-Goals:**
- Watcher del sistema de archivos con inotify; árbol virtualizado para cientos de miles de nodos.

## Decisions

- **Estado:** `editor.State` pasa a tener `Buffers []*BufferState`, `Active int` y `Focus` (`FocusEditor` | `FocusExplorer`).
  `BufferState` agrupa `*buffer.Buffer`, cursor, scroll y metadatos de disco (mtime, tamaño).
- **Contextos del keymap:** el trie se indexa por contexto (`editor/normal`, `editor/insert`, `editor/visual`, `explorer`) en lugar de solo por modo.
  Alternativa descartada: reutilizar el modo Normal para el árbol (mezcla comandos incompatibles como `d`).
- **Árbol (`internal/fs`):** `Node{Path, IsDir, Expanded, Children, Loaded}`; `Expand()` lee el directorio bajo demanda, ordena (directorios primero)
  y aplica la lista de ignorados. No sigue enlaces simbólicos a directorios para evitar ciclos.
- **Operaciones de archivo:** `fs.Create/Rename/Remove` devuelven errores tipados que la UI muestra en línea. El borrado es definitivo en v1, con confirmación obligatoria.
- **Cambios externos:** comparar `mtime`+tamaño al activar buffer, al recuperar el foco (evento de foco de Bubble Tea) y antes de guardar.
- **Iconos:** tabla única en `internal/ui` (clave → glifo Nerd Font, color de la paleta, nombres exactos y extensiones). La referencia es `docs/design/tools/make-icons.py`
  (diccionario `SET`). Resolución: nombre exacto → extensión en minúsculas → icono por defecto. Cada icono ocupa 2 columnas (glifo + espacio) y se
  mide como ancho 1 + 1 (los glifos de Nerd Fonts están en el área de uso privado, sin ancho Unicode fiable).
- **Sin detección de fuente:** una aplicación de terminal no puede saber qué fuente usa la terminal, por lo que `icons` es un ajuste (`nerd` | `none`) y no una detección.
  Alternativa descartada: sondear con códigos de escape (un glifo ausente ocupa el mismo ancho que uno presente).
- **Layout:** el panel del explorador mide 24 columnas fijas (maqueta 01); el resto se asigna al editor. Con terminales estrechas (<80 columnas) el explorador se oculta automáticamente.

- **Decisiones tomadas al implementar:**
  - `editor.Model` conserva los campos de trabajo `buf/cur/top/left/label` del buffer activo y una lista `bufs []*bufState`; `stash()` y `load()` los
    intercambian al cambiar de buffer. Así el código de edición de `modal-editing` no cambió y las 220 pruebas anteriores siguen valiendo.
  - Contextos del keymap: se añade el contexto `explorer` junto a `normal/insert/visual/operator`; con el foco en el explorador se usa ese árbol de teclas.
    Una prueba comprueba que ninguna orden del explorador es alcanzable desde el editor (salvo `explorer.toggle`).
  - Un solo diálogo genérico (`dialog` con botones y una función `choose`) sirve para cambios sin guardar (uno o varios buffers), cerrar buffer,
    borrar archivo y sobrescribir tras un cambio externo. El último botón es siempre «Cancelar» y lo activa `Esc`.
  - Raíz del explorador: la carpeta de trabajo, o la carpeta pasada a `fos`. Los nombres de buffer se muestran relativos a la raíz.
  - El árbol trata los enlaces simbólicos como archivos (nunca los sigue) y la raíz no se puede colapsar, renombrar ni borrar.
  - `Enter`/`l` sobre una carpeta ya abierta bajan a su primer hijo; `h` colapsa o sube a la carpeta padre.
  - Cambios externos: se guardan fecha y tamaño al cargar/guardar. Se revisan al activar un buffer, al recuperar el foco la terminal
    (`tea.FocusMsg`, con `ReportFocus`) y antes de guardar. Un archivo que desaparece del disco no cuenta como cambiado. «Guardar todo y salir» no
    pregunta por sobrescrituras: aborta en el primer conflicto (`E13`) para no pisar nada sin confirmación explícita.
  - `:wq`/`:x` guardan el buffer actual y, si otros buffers tienen cambios, muestran el diálogo de salida con ellos en lugar de salir.
  - La tabla de iconos (`internal/ui/icons_table.go`) se **genera** desde la misma tabla del diseño con `docs/design/tools/make-icons-go.py`,
    de modo que editor y maquetas comparten una sola fuente de verdad. Sin configuración todavía, el modo se elige con `--icons=nerd|none`.
  - `Space e`, `Space b n/p/d` ya funcionan como secuencias; el panel de ayuda del líder llega con `picker-and-palette`.

## Risks / Trade-offs

- [Borrado irreversible] → confirmación obligatoria; valorar enviar a la papelera como mejora.
- [Directorios enormes] → carga perezosa, lista de ignorados, errores de permiso mostrados en línea sin abortar.
- [Renombrar un archivo abierto rompe la ruta del buffer] → actualizar la ruta del `BufferState` en la misma operación.
- [Evento de foco no disponible en todas las terminales] → además comprobar en `Space e` y al guardar; `R` manual siempre disponible.

- [Usuario sin Nerd Font ve cuadros vacíos en lugar de iconos] → documentarlo en el README de instalación, ofrecer `icons = "none"` y una vía mínima (instalar solo la fuente de símbolos, que las terminales usan como respaldo).
- [Un glifo más ancho que una celda se superpone al nombre] → el espacio posterior lo absorbe; probar con la variante normal y la `Mono` de las Nerd Fonts.

## Open Questions

- ¿Enviar a la papelera (`gio trash`) en vez de borrar? Se decide al implementar `d`, sin cambiar la spec (el diálogo ya exige confirmación).
