# buffer-management Specification

## Purpose
Define cómo se mantienen varios archivos abiertos a la vez, cómo se cambia y cierra entre ellos y cómo se protegen los cambios del usuario.

## Requirements

### Requirement: Multiple open buffers
The system SHALL keep any number of files open as buffers, reusing the existing buffer when a file that is already open is opened again, and MUST remember the cursor and scroll position of each buffer.

#### Scenario: Abrir un archivo ya abierto
- **WHEN** se abre `trie.go` desde el explorador estando ya abierto
- **THEN** se activa el buffer existente con su cursor y scroll anteriores

### Requirement: Switching and closing
The system SHALL provide the commands `buffer.next`, `buffer.prev` and `buffer.close` (`Space b n`, `Space b p`, `Space b d`, and `:bn`, `:bp`, `:bd`), and MUST ask for confirmation before closing a buffer with unsaved changes.

#### Scenario: Cerrar buffer modificado
- **WHEN** se cierra un buffer con cambios sin guardar
- **THEN** se muestra el diálogo de cambios sin guardar y el buffer solo se cierra tras elegir guardar o descartar

#### Scenario: Cerrar el último buffer
- **WHEN** se cierra el único buffer abierto
- **THEN** el editor muestra un buffer vacío sin nombre y no termina la aplicación

### Requirement: Quit with modified buffers
Quitting SHALL list every modified buffer in the confirmation dialog and offer to save all, discard all or cancel.

#### Scenario: Salir con dos buffers modificados
- **WHEN** se ejecuta `:q` con dos buffers modificados
- **THEN** el diálogo nombra ambos archivos y la aplicación no termina hasta que el usuario decida

### Requirement: External change detection
The system SHALL detect when a file changed on disk after being loaded: an unmodified buffer MUST reload automatically with a notice, and saving a modified buffer SHALL ask for confirmation before overwriting.

#### Scenario: Archivo cambiado fuera
- **WHEN** un archivo abierto y sin modificar cambia en disco y se vuelve al editor
- **THEN** el buffer se recarga y se muestra un aviso en la línea de comandos

#### Scenario: Guardar sobre cambios externos
- **WHEN** se guarda un buffer modificado cuyo archivo cambió en disco
- **THEN** se pide confirmación antes de sobrescribir y, si se cancela, el archivo en disco no se toca

### Requirement: Buffer indicator
The statusline SHALL show the buffer position (`2/3`) when more than one buffer is open.

#### Scenario: Tres buffers
- **WHEN** hay tres buffers abiertos y el segundo está activo
- **THEN** la statusline muestra `2/3`
