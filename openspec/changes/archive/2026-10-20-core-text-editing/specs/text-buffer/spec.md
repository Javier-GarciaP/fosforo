# Spec Delta

## Purpose
Define cómo se carga, modifica y guarda el texto de un archivo, incluyendo codificación, finales de línea y seguridad ante fallos.

## ADDED Requirements

### Requirement: Loading files
The system SHALL load UTF-8 text files preserving their line-ending style (LF or CRLF) and whether they end with a final newline.

#### Scenario: Archivo CRLF
- **WHEN** se abre un archivo cuyas líneas terminan en CRLF y se guarda sin editar
- **THEN** el contenido guardado es idéntico byte a byte al original

#### Scenario: Archivo inexistente
- **WHEN** se abre una ruta que no existe
- **THEN** se muestra un buffer vacío con ese nombre y el archivo se crea en el primer guardado

### Requirement: Read-only protection
The system MUST open files that contain NUL bytes, invalid UTF-8 or are larger than 10 MiB as read-only and SHALL tell the user why.

#### Scenario: Archivo binario
- **WHEN** se abre un archivo con bytes NUL
- **THEN** el buffer queda en solo lectura, la statusline lo indica y cualquier intento de editar muestra un aviso sin modificar el texto

### Requirement: Atomic saving
The system SHALL save by writing a temporary file in the same directory and renaming it over the target, preserving the original permissions, so a failed save never corrupts the existing file.

#### Scenario: Guardado correcto
- **WHEN** el usuario guarda un buffer modificado
- **THEN** el archivo en disco contiene el texto actual, los permisos no cambian y el indicador de modificado desaparece

#### Scenario: Fallo al guardar
- **WHEN** el directorio no admite escritura
- **THEN** el archivo original queda intacto, el buffer sigue marcado como modificado y se muestra el error en la línea de comandos

### Requirement: Modified tracking
The system SHALL mark a buffer as modified after any text change and clear the mark after a successful save.

#### Scenario: Marca de modificado
- **WHEN** se inserta un carácter en un buffer limpio
- **THEN** el buffer pasa a estado modificado (`[+]`)

### Requirement: Unicode-aware cursor
The system SHALL move the cursor by user-perceived characters (grapheme clusters) and MUST account for display width (wide characters occupy two cells, tabs advance to the next tab stop of 4).

#### Scenario: Carácter ancho
- **WHEN** el cursor pasa sobre un carácter CJK de ancho doble
- **THEN** avanza una posición lógica y dos celdas de pantalla

#### Scenario: Emoji compuesto
- **WHEN** se borra hacia atrás un emoji formado por varios puntos de código
- **THEN** se elimina el emoji completo en una sola pulsación
