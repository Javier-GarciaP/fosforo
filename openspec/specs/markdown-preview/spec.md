# markdown-preview Specification

## Purpose
Define cómo se muestran los archivos Markdown renderizados dentro del editor, cómo se alterna con el fuente y cómo se mantienen sincronizados.

## Requirements

### Requirement: Preview by default
Opening a `.md` or `.markdown` file SHALL show it as a rendered preview in a read-only pane, with the statusline showing the `PREVIEW` mode.

#### Scenario: Abrir un README
- **WHEN** se abre `README.md` desde el explorador o con `fos README.md`
- **THEN** se muestra el documento renderizado y la statusline indica `PREVIEW`, sin modificar el archivo

### Requirement: GitHub-style rendering
The preview SHALL render headings (with an underline rule for levels 1 and 2), paragraphs, bulleted and ordered lists, task lists (`☑` / `☐`), blockquotes, fenced code blocks with syntax highlighting, inline code, bold, italic, strikethrough, links, tables with borders and horizontal rules, using the active theme.

#### Scenario: Lista de tareas
- **WHEN** el documento contiene `- [x] Hecho` y `- [ ] Pendiente`
- **THEN** se muestran como `☑ Hecho` y `☐ Pendiente`

#### Scenario: Tabla
- **WHEN** el documento contiene una tabla con encabezado y filas
- **THEN** se muestra con bordes y las columnas alineadas según el documento

#### Scenario: Bloque de código
- **WHEN** hay un bloque ```` ```go ````
- **THEN** se muestra sobre un fondo propio con colores de sintaxis de Go

### Requirement: Unsupported content
Content that a terminal cannot display MUST degrade gracefully: images SHALL show their alt text as `[imagen: alt]`, and raw HTML or diagrams MUST be omitted or shown dimmed without breaking the layout.

#### Scenario: Imagen
- **WHEN** el documento contiene `![logo](logo.png)`
- **THEN** se muestra `[imagen: logo]`

### Requirement: Width adaptation
The preview SHALL wrap text to the pane width and re-render when the pane is resized.

#### Scenario: Redimensionar
- **WHEN** la terminal se estrecha mientras se ve un párrafo largo
- **THEN** el párrafo se vuelve a ajustar al nuevo ancho sin perder la posición aproximada

### Requirement: Preview navigation
The preview pane SHALL scroll with `j`/`k`, `Ctrl+d`/`Ctrl+u`, `gg` and `G`.

#### Scenario: Ir al final
- **WHEN** se pulsa `G` en la vista previa
- **THEN** se muestra el final del documento

### Requirement: Switching between preview, source and split
`e` or `Space m p` SHALL switch between preview and source editing, `Space m s` SHALL show source and preview side by side, and the cursor MUST land near the corresponding position of the document after switching.

#### Scenario: Editar desde la vista previa
- **WHEN** se está viendo la sección "Instalación" y se pulsa `e`
- **THEN** se abre el fuente en Normal con el cursor en el encabezado `## Instalación`

#### Scenario: Modo dividido
- **WHEN** se pulsa `Space m s`
- **THEN** el fuente aparece a la izquierda y la vista previa a la derecha, como en la maqueta 08

### Requirement: Live update and scroll synchronization
In split mode the preview MUST re-render shortly after edits (debounced) and SHALL follow the source scroll position, aligning on headings.

#### Scenario: Editar en el fuente
- **WHEN** se cambia el texto de un párrafo en el fuente
- **THEN** la vista previa refleja el cambio en menos de un segundo sin perder la posición de scroll

#### Scenario: Scroll sincronizado
- **WHEN** se desplaza el fuente hasta otra sección
- **THEN** la vista previa se desplaza a esa misma sección

### Requirement: Large documents
Documents larger than 2 MiB SHALL open as source with a notice instead of rendering a preview.

#### Scenario: Documento enorme
- **WHEN** se abre un Markdown de 5 MiB
- **THEN** se abre en modo fuente y se informa de que la vista previa está desactivada por tamaño
