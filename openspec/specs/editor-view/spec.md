# editor-view Specification

## Purpose
Define cómo se dibuja el editor en la terminal: área de texto, números de línea, statusline y comportamiento de scroll.

## Requirements

### Requirement: Screen layout
The system SHALL render, from top to bottom, the text area with a line-number gutter, a statusline and a one-row command line, using `~` for rows past the end of the file.

#### Scenario: Archivo corto
- **WHEN** se abre un archivo de 17 líneas en una terminal de 36 filas
- **THEN** las 17 líneas aparecen numeradas, las filas siguientes muestran `~`, y la penúltima fila es la statusline

### Requirement: Statusline content
The statusline SHALL show the current mode label, the file path, the modified marker `[+]`, a read-only marker when applicable, the cursor position as `line:column` and the position in the file as a percentage.

#### Scenario: Posición del cursor
- **WHEN** el cursor está en la línea 12, columna 35 de un archivo de 22 líneas
- **THEN** la statusline muestra `12:35` y el porcentaje correspondiente

### Requirement: Scrolling
The system MUST keep the cursor visible, scrolling vertically with a margin of 3 lines and horizontally for lines wider than the text area, without wrapping lines.

#### Scenario: Cursor llega al borde inferior
- **WHEN** el cursor baja hasta 3 líneas del final visible y existen más líneas
- **THEN** la vista se desplaza para mantener el margen

### Requirement: Resize handling
The system SHALL redraw correctly when the terminal is resized, keeping the cursor visible.

#### Scenario: Terminal más estrecha
- **WHEN** el ancho de la terminal se reduce con el cursor en una columna que ya no cabe
- **THEN** la vista se desplaza horizontalmente para mostrar el cursor
