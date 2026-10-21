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
> **Estado del proyecto: beta en desarrollo (última versión publicada `0.0.1`).** Ya hay código: edición modal tipo Vim, Unicode correcto y guardado seguro. Ver «Qué funciona hoy».

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

Implementado en la rama `dev` (aún sin publicar):

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
| Iconos | Un icono por tipo (nombre exacto → extensión → por defecto) y carpeta cerrada/abierta, con los colores de la paleta; `--icons=none` los quita (carpetas con `/`, abiertas en azul y cerradas atenuadas) |
| Buffers | Varios archivos abiertos (`:e ruta`, el explorador); si ya está abierto se reutiliza con su cursor y scroll; `Space b n/p/d` o `:bn :bp :bd`; `n/m` en la statusline; cerrar o salir pregunta por cada buffer modificado |
| Cambios externos | Se detectan al activar un buffer, al volver el foco a la terminal y al guardar: sin modificar se recarga con aviso; modificado pide confirmación antes de sobrescribir |
| Sintaxis | Colores por lenguaje según el nombre o la extensión (Chroma): palabras clave, tipos, funciones, cadenas, números y comentarios con la paleta del tema; construcciones de varias líneas correctas tras editar |
| Resaltado asíncrono | Escribir nunca espera: los colores anteriores se desplazan con cada edición y se actualizan 50 ms después de la última; los textos de hasta 100 KiB se colorean al abrirlos |
| Límites | Archivos de más de 2 MiB y líneas de más de 2000 caracteres se muestran sin color |
| Temas | `Space t` alterna oscuro/claro y recolorea al instante sin volver a analizar el código |
| Salir | `:q`; con cambios sin guardar aparece `E37` y el diálogo Guardar / Descartar / Cancelar (`:q!` descarta). `Ctrl+C` solo cancela, como `Esc` |
| Vista | Números de línea, `~`, statusline (modo, ruta, `[+]`, solo lectura, `línea:columna`, porcentaje), scroll con margen de 3 líneas y horizontal, aviso por debajo de 40x10 |

Todavía **no** hay: buscador de archivos y paleta, menú líder `Space`, Markdown, ratón ni configuración (ver la hoja de ruta). Sin implementar a propósito: repetir con `.`, objetos de texto (`iw`), macros y marcas.

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
(incluidos `:bn` `:bp` `:bd`). De la tecla líder `Space` ya funcionan `Space e` y `Space b n/p/d`; el panel de ayuda y el resto llegan con `picker-and-palette`.
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
| `Space s g` / `Space s w` | Buscar en el proyecto / la palabra bajo el cursor |
| `Space m p` / `Space m s` | Markdown: alternar vista previa / dividir |
| `Space w` | Guardar |
| `Space t` | Alternar tema claro/oscuro (**ya funciona**) |
| `Space b n` `Space b p` `Space b d` | Buffer siguiente, anterior, cerrar |

Con `Space` pulsado y 300 ms de espera aparece el panel de ayuda (maqueta `04`).

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

- **Los iconos son caracteres de una Nerd Font**, así que la terminal necesita una. Basta con instalar una Nerd Font completa o
  solo la fuente de símbolos («Symbols Nerd Font»), que las terminales usan como respaldo sin cambiar tu fuente actual
  (el paquete oficial incluye un archivo de `fontconfig` para ello en Linux). Descargas: <https://www.nerdfonts.com>.
- **Sin Nerd Font**, ejecuta `fos --icons=none` (más adelante será el ajuste `icons = "none"` de la configuración): las carpetas se muestran con `/` al final y todo sigue funcionando.
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

| # | Versión | Cambio | Qué entrega | Depende de |
| --- | :-: | --- | --- | --- |
| 1 | 0.1.0 | `bootstrap-project` | Go, esqueleto, arranque/salida limpia, CI | — |
| 2 | 0.1.0 | `core-text-editing` | Buffer, vista, abrir/guardar, escritura básica | 1 |
| 3 | 0.2.0 | `modal-editing` | Comandos con nombre, modos, teclas, motions, operadores, `:` | 2 |
| 4 | 0.3.0 | `file-explorer-buffers` | Explorador, varios buffers, cambios externos | 3 |
| 5 | 0.4.0 | `syntax-highlighting` | Resaltado de sintaxis asíncrono | 2 |
| 6 | 0.5.0 | `picker-and-palette` | Picker fuzzy, paleta, menú líder | 3, 4, 5 |
| 7 | 0.6.0 | `markdown-preview` | Vista previa y dividido estilo GitHub | 2, 4, 5 |
| 8 | 0.7.0 | `mouse-support` | Clic, arrastrar, rueda | 3, 4, 6 |
| 9 | 0.8.0 | `config-search-release` | Configuración, búsqueda en proyecto, releases | todos |

Después de `0.8.0` vienen `0.9.x` (pulido y errores) y `1.0.0` (primera versión estable). La numeración sigue el
[Versionado Semántico](https://semver.org/lang/es/): ver [`CONTRIBUTING.md`](../CONTRIBUTING.md) y el [`CHANGELOG`](../CHANGELOG.md).

Si se quiere ver Markdown cuanto antes, el cambio 7 puede adelantarse justo después del 5 (no depende del modo Vim); las versiones se renumeran en ese caso.

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
