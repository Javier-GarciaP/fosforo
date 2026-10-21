# Spec Delta

## Purpose
Define cómo responde el editor al mouse en cada zona de la interfaz, manteniendo el teclado como vía principal.

## ADDED Requirements

### Requirement: Cursor placement
A left click in the text area SHALL move the cursor to the clicked character, accounting for the line-number gutter, tabs, wide characters and scroll, and a click below the last line MUST place the cursor on the last line.

#### Scenario: Clic sobre carácter ancho
- **WHEN** se hace clic en la segunda celda de un carácter CJK de ancho doble
- **THEN** el cursor se coloca sobre ese carácter

#### Scenario: Clic bajo el texto
- **WHEN** se hace clic en una fila `~` posterior al final del archivo
- **THEN** el cursor va a la última línea

### Requirement: Drag selection
Dragging with the left button in the text area SHALL start a character-wise Visual selection from the press point to the pointer, and a double click SHALL select the word under the pointer.

#### Scenario: Seleccionar arrastrando
- **WHEN** se arrastra desde el inicio de `Lookup` hasta su final
- **THEN** la palabra queda seleccionada en modo Visual y `y` la copia

### Requirement: Wheel scrolling
The mouse wheel SHALL scroll the pane under the pointer by 3 lines per notch (editor, explorer, preview, picker and palette lists) without changing focus, and MUST keep the editor cursor inside the visible area.

#### Scenario: Rueda sobre el explorador
- **WHEN** se gira la rueda con el puntero sobre el explorador y el foco está en el editor
- **THEN** se desplaza el explorador y el foco sigue en el editor

### Requirement: Explorer clicks
A click on a directory in the explorer SHALL toggle it, and a click on a file SHALL open it and move focus to the editor.

#### Scenario: Abrir archivo con clic
- **WHEN** se hace clic en `trie.go` en el explorador
- **THEN** el archivo se abre en el editor

### Requirement: Overlay clicks
In pickers, palettes and dialogs, a click on an item SHALL select it, a double click or a click on a dialog button SHALL accept it, and a click outside the overlay MUST close it.

#### Scenario: Cerrar picker con clic fuera
- **WHEN** hay un picker abierto y se hace clic fuera de su ventana
- **THEN** el picker se cierra sin ejecutar nada

### Requirement: Mouse toggle
The system SHALL provide the command `mouse.toggle` to disable and re-enable mouse capture, so the terminal's native text selection can be used.

#### Scenario: Desactivar el mouse
- **WHEN** se ejecuta `mouse.toggle`
- **THEN** los clics dejan de afectar al editor y el terminal permite seleccionar texto con el mouse
