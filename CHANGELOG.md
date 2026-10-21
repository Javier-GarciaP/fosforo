# Changelog

Todos los cambios importantes de fosforo se anotan aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto usa [Versionado Semántico](https://semver.org/lang/es/)
(`0.MINOR.PATCH` mientras esté en beta).

## [Sin publicar]

## [0.8.0] - 2026-10-21

**La primera versión con programa.** Todo lo anterior era diseño y planificación.

### Añadido

- **Editor modal** (`fos [ruta]`): modos Normal, Insert, Visual y línea de comandos; movimientos, operadores (`d c y`) con conteos, deshacer/rehacer, registros y portapapeles (OSC 52), búsqueda `/` y comandos `:w :q :wq :x :e :bn :bp :bd`. Unicode correcto (grafemas, CJK, emoji) y guardado atómico.
- **Explorador** minimalista con iconos por tipo (`--icons=none` para quitarlos), varios buffers y detección de cambios externos.
- **Resaltado de sintaxis** asíncrono con Chroma (temas oscuro y claro, `Space t`).
- **Buscar archivos** difuso con vista previa, **paleta de comandos** (`Space Space`) y menú de la tecla líder `Space` con ayuda visible.
- **Markdown como en GitHub:** los `.md` se abren en vista previa; alternar con el fuente (`e` / `Space m p`) y modo dividido con scroll sincronizado (`Space m s`).
- **Ratón:** clic, arrastrar, doble clic, rueda, explorador, ventanas y diálogos; `:mouse` suelta el ratón para la selección nativa.
- **Configuración** en `~/.config/fos/config.toml` (tema `dark`/`light`/`auto`, iconos, `tabwidth`, `indent`, `scrolloff`, ratón y atajos por contexto) con avisos por clave en lugar de fallar; `Space c o` la abre y `Space c r` (o guardarla) la recarga.
- **Buscar en el proyecto** (`Space s g`, `Space s w`): con `rg` si está instalado y con un buscador propio si no; los resultados llegan mientras escribes.
- **Distribución:** binarios para Linux y macOS (amd64 y arm64) con sumas SHA256 mediante GoReleaser, `scripts/install.sh`, `go install` y un workflow de release que solo acepta etiquetas de `main`.
- **Licencia MIT** (`LICENSE`).

### Cambiado

- **Iconos desactivados por defecto:** sin Nerd Font se veían cuadros vacíos que pisaban los nombres; se activan con `--icons=nerd` o `icons = "nerd"`.
- **Comando `fos`:** el programa se ejecuta con `fos`; *fosforo* se mantiene como nombre del producto y de la marca.

## [0.0.1] - 2026-10-20

**Fundación.** Todavía no hay programa que instalar: esta versión deja listo el diseño, la planificación y el proceso de trabajo.

### Añadido

- **Planificación con OpenSpec:** 9 cambios (125 tareas) que cubren desde el esqueleto del proyecto hasta la configuración y la búsqueda en el proyecto.
- **Diseño visual** con OpenPencil: 12 pantallas, explorador minimalista sin flechas, 38 iconos por tipo de archivo, paleta propia «Tinta» en tema oscuro y claro y modo sin Nerd Font.
- **Marca:** el nombre *fosforo* y su logo unicolor (`< >` con una llama sutil), en SVG y PNG.
- **Documentación:** `README.md`, guía técnica en `docs/`, guía de diseño y `CONTRIBUTING.md`.
- **Proceso de trabajo:** ramas `main` y `dev` protegidas, flujo por pull request con specs, plantilla de PR y política de versiones.

[Sin publicar]: https://github.com/Javier-GarciaP/fosforo/compare/v0.8.0...HEAD
[0.8.0]: https://github.com/Javier-GarciaP/fosforo/releases/tag/v0.8.0
[0.0.1]: https://github.com/Javier-GarciaP/fosforo/releases/tag/v0.0.1
