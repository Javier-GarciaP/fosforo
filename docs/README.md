<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="design/brand/fosforo-logo-white.png">
    <img src="design/brand/fosforo-logo-black.png" width="96" alt="fosforo">
  </picture>
</p>

# fosforo — editor de código en la terminal

> **fosforo**: la cerilla y el fósforo de los monitores CRT. Código que enciende ideas.
>
> Documentación técnica del proyecto. Para una presentación general, mira el [README](../README.md).
>
> **Estado del proyecto: beta (`0.9.0`).** Edición modal tipo Vim, explorador, colores, búsqueda, Markdown, ratón y configuración. Ver «Instalación» y «Qué funciona hoy».

## Qué queremos lograr

Un editor que **se ejecuta dentro de la terminal** y combina lo mejor de dos mundos:

- **La comodidad visual de VS Code:** explorador de archivos, paleta de comandos, resaltado de sintaxis,
  vista previa de Markdown, temas claro y oscuro.
- **La fluidez de Neovim:** edición modal con teclado (Normal / Insert / Visual / Command), secuencias de teclas
  combinables (`3dd`, `dw`, `y$`), y una tecla líder (`Space`) con menú de ayuda visible.
- **Mouse como apoyo, no como requisito:** se puede hacer clic, arrastrar y usar la rueda, pero todo está pensado
  para hacerse sin quitar las manos del teclado.

Quien lo use debe poder **abrir, editar y guardar archivos** con rapidez, y ver los `.md` **renderizados como en GitHub**
sin salir del editor.

### Principios de diseño

1. **Teclado primero.** El camino rápido es siempre el teclado; el mouse solo duplica acciones existentes.
2. **Todo es un comando con nombre.** Teclas, línea `:`, paleta, menú líder y mouse ejecutan el mismo registro de comandos.
   Así cada acción se programa y se prueba una sola vez.
3. **Nunca perder datos ni romper la terminal.** Guardado atómico, aviso de cambios sin guardar, y la terminal se
   restaura siempre al salir (también ante errores).
4. **Descubrible.** El menú líder y la paleta muestran los atajos; no hace falta memorizar nada para empezar.
5. **Fluido.** Dibujar solo lo visible, trabajo pesado (resaltado, Markdown, búsqueda) fuera del bucle de eventos.

## Qué funciona hoy

Incluido hasta la `0.9.0`:

| Función | Detalle |
| --- | --- |
| Arranque | `fos [ruta]`, `--version`, `--help`; pantalla alternativa y terminal restaurada tras salida, pánico y señales |
| Abrir | Archivos UTF-8 conservando LF/CRLF y salto final; ruta inexistente = archivo nuevo; binarios, UTF-8 inválido y >10 MiB en solo lectura con aviso |
| Modos | Normal (por defecto), Insert, Visual (`v`/`V`) y línea de comandos, con su etiqueta y color en la statusline; `Esc` actúa al instante |
| Escribir (Insert) | `i a I A o O`; Enter con sangría, Backspace/Delete (unen líneas), Tab, pegado |
| Moverse | `h j k l w b e 0 ^ $ gg G {n}G`, `Ctrl+d`/`Ctrl+u`, flechas, Home/End, PageUp/PageDown; conteos (`5j`, `3w`); columna deseada en líneas de distinta longitud |
| Editar | Operadores `d c y` con cualquier movimiento (`dw`, `c$`, `dgg`), por línea (`dd cc yy`) y con conteo (`3dd`, `2d3w`); `x D C p P`; selección Visual con `d c y x o` |
| Deshacer | `u` / `Ctrl+r`; una sesión de Insert o una orden completa es un solo paso; deshacer hasta lo guardado deja el archivo sin `[+]` |
| Portapapeles | Lo copiado con `y` va al registro y al portapapeles del sistema (OSC 52 y, si existen, `wl-copy`/`xclip`/`xsel`/`pbcopy`) |
| Buscar | `/patrón`, `n`, `N` con expresiones regulares de Go (por línea); `E486` si no hay coincidencias |
| Línea de comandos | `:w [ruta]`, `:q`, `:q!`, `:wq`, `:x`, `:e [ruta]`, `:{n}`, con historial (`↑`/`↓`); `E492` si el comando no existe |
| Comandos | Registro único de comandos con nombre (`file.save`, `cursor.down`...): teclas y `:` ejecutan los mismos |
| Unicode | Cursor por caracteres percibidos (emoji compuestos, combinantes), ancho doble (CJK) y tabuladores de 4 |
| Guardar | `Ctrl+S` o `:w`: guardado atómico (archivo temporal + `rename`) que conserva permisos; fallos sin perder el original |
| Explorador | `Space e` lo muestra/oculta (24 columnas; se oculta solo bajo 80); `Tab`/`Ctrl+w` cambian el foco; `j k l h Enter gg G`; `a` crea (con `/` final, carpeta), `r` renombra, `d` borra con confirmación, `R` refresca; `fos carpeta/` lo abre con foco |
| Árbol | Sangría de 2 columnas sin flechas ni líneas guía; carpetas primero y por orden alfabético; oculta `.git` y `node_modules`; carga carpetas al expandirlas; no sigue enlaces simbólicos |
| Iconos | Un icono por tipo (nombre exacto → extensión → por defecto) y carpeta cerrada/abierta, con los colores de la paleta; vienen desactivados por defecto; `--icons=nerd` o `icons = "nerd"` los activan (sin ellos, carpetas con `/`, abiertas en azul y cerradas atenuadas) |
| Buffers | Varios archivos abiertos (`:e ruta`, el explorador); si ya está abierto se reutiliza con su cursor y scroll; `Space b n/p/d` o `:bn :bp :bd`; `n/m` en la statusline; cerrar o salir pregunta por cada buffer modificado |
| Cambios externos | Se detectan al activar un buffer, al volver el foco a la terminal y al guardar: sin modificar se recarga con aviso; modificado pide confirmación antes de sobrescribir |
| Sintaxis | Colores por lenguaje según el nombre o la extensión (Chroma): palabras clave, tipos, funciones, cadenas, números y comentarios con la paleta del tema; construcciones de varias líneas correctas tras editar |
| Resaltado asíncrono | Escribir nunca espera: los colores anteriores se desplazan con cada edición y se actualizan 50 ms después de la última; los textos de hasta 100 KiB se colorean al abrirlos |
| Límites | Archivos de más de 2 MiB y líneas de más de 2000 caracteres se muestran sin color |
| Temas | `Space t` alterna oscuro/claro y recolorea al instante sin volver a analizar el código |
| Pantalla de inicio | Shown when no file is open: FOSFORO logo plus the main leader shortcuts on the right; hides the logo (or everything) on small windows; informational only, keys and mouse are unchanged |
| Buscar archivos | `Space f f`: ventana difusa con vista previa coloreada, contador `coincidencias/total` y carga progresiva en proyectos grandes (no bloquea la entrada); `Space f r` recientes; `Space f n` nuevo |
| Buffers (picker) | `Space b b` lista los buffers abiertos marcando los modificados |
| Paleta de comandos | `Space Space` lista todos los comandos por su descripción con sus atajos a la derecha; `Enter` lo ejecuta igual que su atajo; `Space ?`/`Space h` muestran la lista de atajos |
| Sugerencias en `:` | Al escribir en `:` aparecen las mismas sugerencias; `Enter` ejecuta lo escrito, o la sugerencia si se eligió con `↑`/`↓` |
| Menú líder | Tras `Space` y 300 ms aparece el panel de ayuda generado desde las teclas reales (grupos `+`, `Esc` cancela) |
| Markdown | Los `.md` y `.markdown` se abren en **vista previa** (solo lectura, `PREVIEW`): encabezados con regla, listas, tareas `☑ ☐`, citas, tablas con bordes, código con colores, enlaces, imágenes como `[imagen: alt]`; se ajusta al ancho y se vuelve a ajustar al redimensionar |
| Editar Markdown | `e` (en la vista previa) o `Space m p` alterna con el fuente, dejando el cursor en el encabezado de la sección que veías; `Space m s` divide fuente \| vista previa con scroll sincronizado y actualización 150 ms después de editar |
| Límites de Markdown | Documentos de más de 2 MiB se abren como fuente con aviso; HTML crudo y diagramas se omiten con un aviso; las tablas anchas se recortan con `…` |
| Ratón | Clic: coloca el cursor (con números de línea, tabuladores y caracteres anchos); arrastrar: selección Visual; doble clic: la palabra; rueda: 3 líneas por muesca en el panel bajo el puntero (editor, explorador, vista previa, listas) sin cambiar el foco; clic en el explorador abre archivos y alterna carpetas; clic y doble clic en picker, paleta y botones de diálogos; clic fuera cierra el picker |
| Ratón y selección nativa | `:mouse` (o `mouse.toggle` en la paleta) suelta el ratón para usar la selección de texto del terminal; con el ratón activado, `Shift`+arrastrar sigue seleccionando de forma nativa en la mayoría de terminales y en `tmux` |
| Buscar en el proyecto | `Space s g` abre una búsqueda de texto: los resultados (`ruta:línea  texto`) aparecen mientras escribes (150 ms tras la última tecla), se cancelan al seguir escribiendo y `Enter` abre el archivo en esa línea. `Space s w` busca la palabra bajo el cursor. Distingue mayúsculas solo si escribes alguna; hasta 1000 resultados |
| Motor de búsqueda | Usa `rg` (ripgrep) si está instalado, respetando `.gitignore`; si no, un buscador propio que omite `.git`, `node_modules`, binarios y archivos de más de 5 MiB |
| Configuración | Archivo TOML opcional (`~/.config/fos/config.toml`): `theme` (`dark`/`light`/`auto`), `icons`, `tabwidth`, `indent`, `scrolloff`, `mouse` y atajos por contexto; un valor inválido avisa con el nombre de la clave y usa el valor por defecto, nunca impide arrancar. Ver «Configuración» |
| Recargar | `Space c r` (`config.reload`) vuelve a leer el archivo; guardar el propio archivo de configuración en el editor lo recarga solo. `Space c o` lo abre (creándolo con una plantilla comentada si no existe) |
| Salir | `:q`; con cambios sin guardar aparece `E37` y el diálogo Guardar / Descartar / Cancelar (`:q!` descarta). `Ctrl+C` solo cancela, como `Esc` |
| Vista | Números de línea, `~`, statusline (modo, ruta, `[+]`, solo lectura, `línea:columna`, porcentaje), scroll con margen de 3 líneas y horizontal, aviso por debajo de 40x10 |

Sin implementar a propósito: repetir con `.`, objetos de texto (`iw`), macros y marcas.

## Alcance de la versión 1

| Incluido en v1 | Fuera de v1 (por ahora) |
| --- | --- |
| Abrir, editar y guardar archivos de texto | Pestañas visuales |
| Edición modal tipo Vim (subconjunto documentado) | Terminal integrada |
| Explorador de archivos y varios buffers | LSP (autocompletado, diagnósticos) |
| Resaltado de sintaxis | Integración con git |
| Buscar archivos (fuzzy), paleta de comandos, menú líder | Plugins y scripting |
| **Vista previa de Markdown estilo GitHub** (completa y dividida) | Imágenes, Mermaid y HTML en la vista previa |
| Mouse (clic, arrastrar, rueda) | División de ventanas (splits) |
| Configuración, temas claro/oscuro, búsqueda en proyecto | Windows como plataforma oficial |
| Binarios para Linux y macOS | Repetir con `.`, objetos de texto, macros |

## Cómo se verá

El diseño visual vive en [`design/`](design/README.md) y se hizo con **OpenPencil**. Son 12 pantallas sobre una
cuadrícula de terminal de 120×36 celdas.

| Pantalla | Imagen |
| --- | --- |
| Guía de estilo (paleta, modos, glifos) | [`00`](design/exports/00-guia-de-estilo.png) |
| Layout principal, modo Normal | [`01`](design/exports/01-layout-principal-normal.png) |
| Modo Insert | [`02`](design/exports/02-modo-insert.png) |
| Modo Visual | [`03`](design/exports/03-modo-visual.png) |
| Menú líder `Space` | [`04`](design/exports/04-menu-lider-space.png) |
| Buscar archivos (picker) | [`05`](design/exports/05-picker-de-archivos.png) |
| Paleta de comandos | [`06`](design/exports/06-paleta-de-comandos.png) |
| Markdown: vista previa | [`07`](design/exports/07-markdown-vista-previa.png) |
| Markdown: dividido fuente \| vista previa | [`08`](design/exports/08-markdown-dividido.png) |
| Tema claro | [`09`](design/exports/09-tema-claro-normal.png) |
| Diálogo de cambios sin guardar | [`10`](design/exports/10-dialogo-cambios-sin-guardar.png) |
| Iconos por tipo de archivo y modo sin Nerd Font | [`11`](design/exports/11-iconos-y-explorador.png) |
| Marca: logo unicolor, versiones y tamaños | [`12`](design/exports/12-marca-fosforo.png) |

![Layout principal](design/exports/01-layout-principal-normal.png)

## Atajos previstos

Los atajos quedan fijados por las specs (`openspec/specs/`). **Ya funcionan** todos los de «Modos», «Movimiento y edición» y «Línea de comandos»,
(incluidos `:bn` `:bp` `:bd`). La tecla líder `Space` ya funciona entera salvo el grupo `s` (buscar), que llega con su cambio y aparecerá solo en el panel.
En Insert: `Esc`/`Ctrl+C` vuelven a Normal, `Ctrl+S` guarda y las flechas cruzan de línea.

### Modos

| Tecla | Acción |
| --- | --- |
| `Esc` | Volver a Normal |
| `i` `a` `I` `A` `o` `O` | Entrar en Insert (en distintos puntos) |
| `v` / `V` | Selección Visual por carácter / por línea |
| `:` | Línea de comandos |

### Movimiento y edición (Normal)

| Tecla | Acción |
| --- | --- |
| `h` `j` `k` `l` (o flechas) | Izquierda, abajo, arriba, derecha |
| `w` `b` `e` | Siguiente palabra, palabra anterior, fin de palabra |
| `0` `^` `$` | Inicio de línea, primer carácter, fin de línea |
| `gg` `G` `12G` | Primera línea, última línea, ir a la línea 12 |
| `d` `c` `y` + movimiento | Borrar, cambiar, copiar (`dw`, `y$`); duplicado = línea (`dd` `cc` `yy`) |
| `x` `D` `C` `p` `P` | Borrar carácter, borrar/cambiar hasta fin de línea, pegar |
| `u` / `Ctrl+r` | Deshacer / rehacer |
| `/texto` `n` `N` | Buscar en el archivo, siguiente, anterior |
| `3dd`, `5j` | Un número antes repite la acción |
| `Ctrl+d` / `Ctrl+u` | Media pantalla abajo / arriba |
| `Ctrl+s` | Guardar (en cualquier modo) |

### Línea de comandos

| Comando | Acción |
| --- | --- |
| `:w` `:w ruta` | Guardar / guardar como |
| `:q` `:q!` `:wq` `:x` | Salir / salir sin guardar / guardar y salir |
| `:e ruta` | Abrir archivo en un buffer (`:e` recarga el actual) |
| `:42` | Ir a la línea 42 |
| `:bn` `:bp` `:bd` | Siguiente buffer, anterior, cerrar |

### Tecla líder `Space`

| Secuencia | Acción |
| --- | --- |
| `Space f f` | Buscar archivo |
| `Space b b` | Buscar entre buffers abiertos |
| `Space e` | Mostrar/ocultar explorador |
| `Space Space` | Paleta de comandos |
| `Space f r` / `Space f n` | Archivos recientes / nuevo archivo |
| `Space w` / `Space q` / `Space Q` | Guardar / cerrar buffer / salir |
| `Space t` | Alternar tema claro/oscuro |
| `Space /` / `Space :` | Buscar en el archivo / línea de comandos |
| `Space ?` / `Space h` | Lista de atajos |
| `Space b n` `Space b p` `Space b d` | Buffer siguiente, anterior, cerrar |
| `Space s g` / `Space s w` | Buscar texto en el proyecto / la palabra bajo el cursor |
| `Space c o` / `Space c r` | Abrir / recargar la configuración |
| `Space m p` / `Space m s` | Markdown: alternar vista previa y fuente (en la vista previa también `e`) / dividir |

Con `Space` pulsado y 300 ms de espera aparece el panel de ayuda (maqueta `04`), generado desde las teclas reales.

## Instalación

`fos` es un único programa sin dependencias. Hay binarios para Linux y macOS (amd64 y arm64).

```bash
# 1) Script de instalación (verifica la suma SHA256; instala en ~/.local/bin)
curl -fsSL https://raw.githubusercontent.com/Javier-GarciaP/fosforo/main/scripts/install.sh | sh
#    FOS_VERSION=0.9.0 fija la versión; INSTALL_DIR=/usr/local/bin cambia la carpeta

# 2) Con Go 1.27 o superior
go install github.com/Javier-GarciaP/fosforo/cmd/fos@latest

# 3) A mano: descarga fos_<versión>_<sistema>_<arquitectura>.tar.gz de la página de releases,
#    comprueba checksums.txt y copia `fos` a una carpeta de tu PATH
```

Después: `fos --version`, `fos archivo.go`, `fos carpeta/`. Para ver los iconos del explorador hace falta una
[Nerd Font](https://www.nerdfonts.com/) y `--icons=nerd` (o `icons = "nerd"`); por defecto no se muestran. **Opcional:** instala [`rg`](https://github.com/BurntSushi/ripgrep)
(ripgrep) para que `Space s g` sea más rápido y respete `.gitignore`; sin él funciona igual con el buscador propio.

Las versiones se publican solo desde `main` con una etiqueta `vX.Y.Z` (el workflow de release falla si la etiqueta no está en `main`).

## Configuración

Todo es opcional. El archivo es `$XDG_CONFIG_HOME/fos/config.toml` (por defecto `~/.config/fos/config.toml`); `Space c o` lo crea
con una plantilla comentada.

```toml
theme = "auto"      # dark | light | auto (auto lee el fondo de la terminal; si no responde, oscuro)
icons = "none"      # none | nerd  (--icons en la línea de comandos manda sobre esto)
tabwidth = 4        # 1..16
indent = "tabs"     # tabs | spaces: qué inserta la tecla Tab
scrolloff = 3       # 0..20 líneas de margen
mouse = true

[keys.normal]       # contextos: normal, insert, visual, explorer, preview
"space g" = "search.project"   # secuencia = comando (los nombres salen de la paleta)
"space t" = ""                 # un comando vacío quita el atajo
```

Un comando inexistente o una secuencia que choca con otra se rechazan con un aviso (el atajo de fábrica se conserva).

## Lenguajes con resaltado

Se elige por nombre o extensión (sin distinguir mayúsculas) y funciona con los más de 200 lenguajes de Chroma. Verificados con pruebas:
**Go** (también `go.mod`/`go.work`), **Markdown**, **JSON**, **YAML**, **TOML**, **JavaScript**, **TypeScript**, **Python**, **Rust**, **shell/Bash**,
**Makefile**, **Dockerfile**, **CSS** y **HTML**. Los tipos desconocidos (`.txt`, `.xyz`) se ven como texto plano sin errores.

## Markdown como en GitHub

Al abrir un `.md` se ve **renderizado**: encabezados con regla, listas, tareas (`☑ ☐`), citas, tablas con bordes,
código con colores y enlaces. Con `e` se pasa a editar el fuente; con `Space m s` se ven fuente y vista previa lado a lado,
con desplazamiento sincronizado.

**Límites honestos** (una terminal no es un navegador): no hay imágenes reales (se muestra `[imagen: texto]`), ni Mermaid,
ni HTML embebido. La fidelidad con GitHub es aproximada: misma estructura y jerarquía visual.

## Iconos y fuentes

El explorador no usa flechas: la estructura se lee por la **sangría** y el estado de cada carpeta por su **icono** (cerrada o
abierta). Cada archivo lleva el icono de su tipo (Go, Markdown, JSON, Makefile, LICENSE, `.gitignore`…; ver la pantalla
[`11`](design/exports/11-iconos-y-explorador.png)).

- **Los iconos vienen desactivados por defecto** (`icons = "none"`): las carpetas se muestran con `/` al final y todo funciona igual. Actívalos con `fos --icons=nerd` o `icons = "nerd"`; son caracteres de una Nerd Font, así que la terminal necesita una. Basta con instalar una Nerd Font completa o
  solo la fuente de símbolos («Symbols Nerd Font»), que las terminales usan como respaldo sin cambiar tu fuente actual
  (el paquete oficial incluye un archivo de `fontconfig` para ello en Linux). Descargas: <https://www.nerdfonts.com>.
- No se puede detectar automáticamente qué fuente usa la terminal, por eso es un ajuste y no una detección.

## Stack y arquitectura

- **Lenguaje:** Go. **Interfaz:** Bubble Tea (bucle Model/Update/View), Lip Gloss (estilos), Glamour (Markdown),
  Chroma (sintaxis). Se verifica la versión vigente de cada librería al empezar el código.
- **Un solo binario**, sin dependencias en tiempo de ejecución (`rg` es opcional para buscar más rápido).

```
cmd/fos/            arranque y flags
internal/buffer/    texto, cursor, undo/redo, cargar y guardar (sin UI)
internal/editor/    modos, estado global (buffers, foco), despacho de comandos
internal/command/   registro de comandos con nombre
internal/keymap/    teclas → comandos por contexto, secuencias, conteos
internal/ui/        vistas: editor, explorador, statusline, picker, overlays, preview
internal/syntax/    tokenización con Chroma y caché
internal/markdown/  render con Glamour (estilos github-dark/light)
internal/fs/        árbol de directorios, creación/renombrado/borrado
internal/config/    configuración TOML, temas y remapeo de teclas
internal/search/    búsqueda en el proyecto
```

## Hoja de ruta

Cada hito es un **cambio de OpenSpec** en [`openspec/changes/`](../openspec/changes/) con propuesta, specs, diseño y tareas
(125 tareas en total). Se implementan en este orden; "depende de" indica qué debe estar archivado antes.

| # | Cambio | Qué entregó | Estado |
| --- | --- | --- | :-: |
| 1 | `bootstrap-project` | Go, esqueleto, arranque/salida limpia, CI | ✅ |
| 2 | `core-text-editing` | Buffer, vista, abrir/guardar, escritura básica | ✅ |
| 3 | `modal-editing` | Comandos con nombre, modos, teclas, motions, operadores, `:` | ✅ |
| 4 | `file-explorer-buffers` | Explorador, varios buffers, cambios externos | ✅ |
| 5 | `syntax-highlighting` | Resaltado de sintaxis asíncrono | ✅ |
| 6 | `picker-and-palette` | Picker fuzzy, paleta, menú líder | ✅ |
| 7 | `markdown-preview` | Vista previa y dividido estilo GitHub | ✅ |
| 8 | `mouse-support` | Clic, arrastrar, rueda | ✅ |
| 9 | `config-search-release` | Configuración, búsqueda en proyecto, releases | ✅ |
| + | `icons-off-by-default` | Iconos desactivados por defecto (sin Nerd Font se veían cuadros) | ✅ |

Los diez cambios se publican juntos en la **`0.8.0`** (los hitos `0.1.0`–`0.7.0` del plan original quedaron incluidos en ella), con binarios para Linux y macOS.

Después de `0.9.0` vienen `0.9.x` (pulido y errores) y `1.0.0` (primera versión estable). La numeración sigue el
[Versionado Semántico](https://semver.org/lang/es/): ver [`CONTRIBUTING.md`](../CONTRIBUTING.md) y el [`CHANGELOG`](../CHANGELOG.md).

## Cómo se construyó

Todo el programa se hizo siguiendo el flujo de [`CONTRIBUTING.md`](../CONTRIBUTING.md): una **especificación (OpenSpec) por cambio**, una rama `feat/<cambio>`
desde `dev`, un pull request a `dev` con CI en verde (`vet · test · build`), el cambio archivado en [`openspec/changes/archive/`](../openspec/changes/archive/)
y, al final, un único PR `dev` → `main` con la etiqueta `v0.8.0`. Quedan 23 especificaciones vigentes en [`openspec/specs/`](../openspec/specs/).
Decisiones que conviene conocer: Go + Bubble Tea con una cuadrícula de celdas propia (los colores y el ratón comparten una sola geometría), un registro único de
comandos (teclas, `:`, paleta y ratón ejecutan lo mismo), goldmark propio para Markdown en lugar de Glamour, Chroma asíncrono para la sintaxis, y pruebas con
`teatest` y con un pseudo-terminal real. El detalle de cada cambio está en su PR y en su `design.md` archivado.

## Cómo trabajamos

- **Todo cambio pasa por una spec** (OpenSpec): se propone, se valida y se archiva. Ver [`../CONTRIBUTING.md`](../CONTRIBUTING.md).
- **Ramas:** `main` solo guarda versiones estables; `dev` integra todo; cada cambio vive en su rama y entra a `dev` por PR.
- **Diseño visual:** se edita con OpenPencil. Guía de uso en [`design/README.md`](design/README.md).

## Requisitos para desarrollar

| Herramienta | Para qué | Estado en esta máquina |
| --- | --- | --- |
| Go 1.27.3 | Compilar y probar | Instalado en `~/.local/go` (enlazado en `~/.local/bin`) |
| OpenSpec CLI | Specs, validación, archivado | Instalado |
| OpenPencil (`op` + app de escritorio) | Ver y modificar el diseño | Instalado |
| `gh` | PRs desde la terminal | Instalado y autenticado |
| `rg` (ripgrep) | Búsqueda más rápida (opcional) | Instalado (14.1.1) |
| Nerd Font o «Symbols Nerd Font» | Ver los iconos del explorador | **Sin instalar** (la fuente del sistema, Adwaita Mono, no los incluye) |

## Comandos de desarrollo

Con Go instalado, desde un clon limpio (no hace falta nada más para compilar):

```bash
git clone https://github.com/Javier-GarciaP/fosforo.git && cd fosforo
make build            # compila bin/fos
make run              # compila y abre el editor (make run ARGS="README.md")
make test             # todas las pruebas, con detector de carreras
make vet              # go vet + comprobación de formato (gofmt)
make fmt              # da formato al código
./bin/fos --version   # fos dev (abc1234)
```

Las pruebas de ciclo de vida ejecutan el binario real dentro de un pseudo-terminal (Linux/macOS) y comprueban que la terminal
queda restaurada tras salida normal, pánico y señales (`SIGINT`, `SIGTERM`). Si falla la terminal de tu entorno,
`reset` la deja como nueva.

El CI (`.github/workflows/ci.yml`) ejecuta `make vet`, `make test` y `make build` en cada PR a `dev` y a `main`, y rechaza los PR a
`main` que no salgan de `dev`.
