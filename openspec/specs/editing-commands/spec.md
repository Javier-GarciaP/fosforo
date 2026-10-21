# editing-commands Specification

## Purpose
Define las operaciones de edición en modo Normal y Visual: movimientos, operadores, pegado, deshacer y búsqueda dentro del archivo.

## Requirements

### Requirement: Motions
The system SHALL support the motions `h j k l`, `w b e`, `0 ^ $`, `gg G` and `{n}G`, and `Ctrl+d` / `Ctrl+u` for half-page scrolling.

#### Scenario: Saltar a una línea
- **WHEN** se escribe `12G`
- **THEN** el cursor va a la línea 12

#### Scenario: Siguiente palabra
- **WHEN** se pulsa `w` con el cursor al inicio de una palabra
- **THEN** el cursor avanza al inicio de la siguiente palabra

### Requirement: Operators
The system SHALL support the operators `d`, `c` and `y` combined with any motion, doubled for a whole line (`dd`, `cc`, `yy`), plus `x`, `D`, `C`, `p` and `P`.

#### Scenario: Borrar hasta fin de palabra
- **WHEN** se escribe `dw` sobre la primera palabra de una línea
- **THEN** se elimina esa palabra y su espacio siguiente

#### Scenario: Cambiar línea
- **WHEN** se escribe `cc`
- **THEN** la línea se vacía y se entra en Insert

### Requirement: Registers and clipboard
The system SHALL store the last deleted or yanked text in an unnamed register used by `p`/`P`, and MUST also copy yanked text to the system clipboard when the terminal supports it.

#### Scenario: Copiar y pegar línea
- **WHEN** se escribe `yy` y luego `p`
- **THEN** la línea se duplica debajo de la actual

### Requirement: Undo and redo
The system SHALL undo with `u` and redo with `Ctrl+r`, treating everything typed during one Insert session as a single undo step.

#### Scenario: Deshacer una sesión de Insert
- **WHEN** se escribe `ihola<Esc>` y luego `u`
- **THEN** desaparece `hola` en un solo paso

### Requirement: Visual mode
The system SHALL provide character-wise (`v`) and line-wise (`V`) selection, highlighted as in the design, on which `d`, `c` and `y` operate.

#### Scenario: Borrar selección
- **WHEN** se pulsa `V`, se mueve el cursor dos líneas abajo y se pulsa `d`
- **THEN** se eliminan las tres líneas seleccionadas y se vuelve a Normal

### Requirement: Search in file
The system SHALL search the current buffer with `/pattern`, jump to the next match on Enter and cycle matches with `n` and `N`.

#### Scenario: Buscar y repetir
- **WHEN** se escribe `/trie`, Enter y luego `n`
- **THEN** el cursor salta a la primera coincidencia y después a la siguiente

#### Scenario: Sin coincidencias
- **WHEN** el patrón no aparece en el archivo
- **THEN** se muestra `E486: patrón no encontrado: <patrón>` y el cursor no se mueve
