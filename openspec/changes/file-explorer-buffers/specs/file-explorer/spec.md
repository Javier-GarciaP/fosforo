# Spec Delta

## Purpose
Define el panel de explorador de archivos: cómo se muestra el árbol, cómo se navega con teclado y qué operaciones de archivo ofrece.

## ADDED Requirements

### Requirement: Toggle and focus
The system SHALL show or hide the explorer with `Space e` and move focus between explorer and editor with `Tab` or `Ctrl+w`, indicating the focused pane.

#### Scenario: Alternar explorador
- **WHEN** el explorador está oculto y se pulsa `Space e`
- **THEN** aparece a la izquierda con 24 columnas de ancho y recibe el foco

#### Scenario: Volver al editor
- **WHEN** el foco está en el explorador y se pulsa `Tab`
- **THEN** el foco pasa al editor sin cerrar el explorador

### Requirement: Tree display
The system SHALL display the project directory as a tree with directories first, then files in alphabetical order, hiding `.git` and `node_modules`, and loading the children of a directory only when it is expanded.

#### Scenario: Directorio grande
- **WHEN** el proyecto contiene un directorio con miles de archivos sin expandir
- **THEN** la apertura del explorador no lee su contenido y sigue respondiendo de inmediato

#### Scenario: Archivo actual visible
- **WHEN** se abre el explorador con un archivo en el editor
- **THEN** el árbol expande sus carpetas padre y resalta ese archivo

### Requirement: Minimal tree rendering
The explorer SHALL show the hierarchy only through indentation of 2 columns per level, without disclosure arrows or box-drawing guide lines, with the project name in bold as the first row, directory names in the theme's blue, file names in the default text color and the selected row highlighted.

#### Scenario: Estructura por sangría
- **WHEN** se muestra el archivo `internal/keymap/trie.go` con las carpetas expandidas
- **THEN** `trie.go` aparece sangrado 3 niveles (6 columnas) bajo la raíz, sin flechas ni líneas guía (ver `docs/design/exports/01-layout-principal-normal.png`)

### Requirement: Folder state icons
When icons are enabled, the explorer SHALL show a closed-folder icon for collapsed directories and an open-folder icon for expanded ones, so that the state is visible without arrows.

#### Scenario: Expandir una carpeta
- **WHEN** se expande `buffer` con `l`
- **THEN** su icono pasa de carpeta cerrada a carpeta abierta y sus hijos aparecen sangrados debajo

### Requirement: File type icons
When icons are enabled, each file SHALL show an icon chosen first by exact file name (for example `Makefile`, `go.mod`, `.gitignore`, `LICENSE`, `Dockerfile`), then by extension (case-insensitive), and otherwise a default file icon, colored with the active theme palette as listed in `docs/design/exports/11-iconos-y-explorador.png`.

#### Scenario: Por extensión
- **WHEN** se muestra `trie.go`
- **THEN** lleva el icono de Go en el color teal del tema

#### Scenario: Por nombre exacto
- **WHEN** se muestra `Makefile`, que no tiene extensión
- **THEN** lleva el icono de make

#### Scenario: Tipo desconocido
- **WHEN** se muestra `notas.xyz`
- **THEN** lleva el icono de archivo por defecto

### Requirement: Icon fonts and fallback
Icons SHALL be Nerd Font glyphs, each followed by one space, and the explorer MUST remain fully usable without them: with `icons = "none"`, directory names end with `/`, expanded directories are shown in blue and collapsed ones dimmed, and no icon characters are printed.

#### Scenario: Sin Nerd Font
- **WHEN** el ajuste `icons` vale `none`
- **THEN** se ve `keymap/` en azul con sus hijos sangrados y `buffer/` atenuado, sin ningún icono

### Requirement: Keyboard navigation
In the explorer, `j` and `k` SHALL move the selection, `Enter` or `l` open a file or expand a directory, and `h` collapse a directory or move to its parent.

#### Scenario: Abrir archivo
- **WHEN** se selecciona `trie.go` y se pulsa `Enter`
- **THEN** el archivo se abre en el editor y el foco pasa al editor

### Requirement: File operations
The explorer SHALL create files and directories with `a` (a trailing `/` creates a directory), rename with `r` and delete with `d`, asking for confirmation before deleting and showing any error inline.

#### Scenario: Crear archivo
- **WHEN** se pulsa `a`, se escribe `util/strings.go` y se confirma
- **THEN** se crea el archivo con sus carpetas intermedias y aparece seleccionado en el árbol

#### Scenario: Borrar con confirmación
- **WHEN** se pulsa `d` sobre un archivo y se cancela el diálogo
- **THEN** el archivo no se borra

#### Scenario: Renombrar archivo abierto
- **WHEN** se renombra un archivo que está abierto en un buffer
- **THEN** el buffer pasa a usar la nueva ruta sin perder cambios sin guardar

### Requirement: Refresh
The explorer SHALL refresh its contents with `R`, after saving a file, and when the application regains focus.

#### Scenario: Archivo creado fuera del editor
- **WHEN** se crea un archivo desde otra terminal y se pulsa `R`
- **THEN** el archivo aparece en el árbol
