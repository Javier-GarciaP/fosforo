# Spec Delta

## Purpose
Define la edición directa de texto con el teclado: escribir, borrar, saltar de línea y moverse, que más adelante pasa a ser el modo Insert.

## ADDED Requirements

### Requirement: Text insertion and deletion
The system SHALL insert printable characters at the cursor, split the line on Enter keeping the leading whitespace of the current line, and delete characters with Backspace and Delete, joining lines at line boundaries.

#### Scenario: Escribir y saltar de línea
- **WHEN** el usuario escribe `ab`, pulsa Enter y escribe `c`
- **THEN** el buffer contiene dos líneas, `ab` y `c`

#### Scenario: Unir líneas
- **WHEN** el cursor está al inicio de una línea y se pulsa Backspace
- **THEN** la línea se une con la anterior y el cursor queda en el punto de unión

### Requirement: Cursor movement keys
The system SHALL support the arrow keys, Home, End, PageUp and PageDown to move the cursor, clamping it to valid positions.

#### Scenario: Flecha abajo en línea más corta
- **WHEN** el cursor está en la columna 20 y la línea siguiente tiene 5 caracteres
- **THEN** el cursor baja a la columna 5 y al volver a bajar a una línea larga recupera la columna 20

### Requirement: Transitional save and quit
Until modal editing exists, the system SHALL save with `Ctrl+S` and quit with `Ctrl+Q`, and MUST ask for confirmation before quitting with unsaved changes.

#### Scenario: Salir con cambios
- **WHEN** se pulsa `Ctrl+Q` con cambios sin guardar
- **THEN** se muestra el diálogo de cambios sin guardar (ver `docs/design/exports/10-dialogo-cambios-sin-guardar.png`) y no se sale hasta confirmar
