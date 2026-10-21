# Spec Delta

## Purpose
Define los modos de edición, cómo se interpretan las secuencias de teclas y los conteos, y cómo se indica el modo al usuario.

## ADDED Requirements

### Requirement: Editing modes
The system SHALL start in Normal mode and provide Insert, Visual (character and line) and Command modes, showing the current mode in the statusline with its own color.

#### Scenario: Entrar y salir de Insert
- **WHEN** el usuario pulsa `i` en Normal y luego `Esc`
- **THEN** pasa a Insert (etiqueta `INSERT` en verde) y vuelve a Normal con el cursor una posición a la izquierda

#### Scenario: Teclas de entrada a Insert
- **WHEN** se pulsa `a`, `I`, `A`, `o` u `O` en Normal
- **THEN** se entra en Insert después del cursor, al inicio de la línea, al final de la línea, o en una línea nueva debajo o encima

### Requirement: Key sequence resolution
The system SHALL resolve multi-key sequences with a prefix tree and, when a sequence is incomplete, wait up to 1000 ms for the next key, showing the pending keys in the command line.

#### Scenario: Secuencia completa
- **WHEN** se pulsa `g` y luego `g` en Normal
- **THEN** el cursor salta a la primera línea

#### Scenario: Secuencia incompleta
- **WHEN** se pulsa `g` y pasan más de 1000 ms sin otra tecla
- **THEN** la secuencia pendiente se descarta sin cambiar el texto

### Requirement: Counts
The system SHALL accept a numeric count prefix before a command or motion and repeat it that many times.

#### Scenario: Borrar varias líneas
- **WHEN** se escribe `3dd` en Normal
- **THEN** se eliminan 3 líneas desde la actual

### Requirement: Unmapped keys in Normal mode
Keys without a binding in Normal or Visual mode MUST be ignored and MUST NOT insert text.

#### Scenario: Letra sin función
- **WHEN** se pulsa una tecla sin asignación en Normal
- **THEN** el buffer no cambia

### Requirement: Immediate Escape
Pressing `Esc` SHALL take effect without an artificial delay.

#### Scenario: Escape rápido
- **WHEN** se pulsa `Esc` en Insert
- **THEN** el modo cambia a Normal en el mismo ciclo de actualización

### Requirement: Save shortcut in every mode
`Ctrl+S` SHALL run `file.save` in Normal, Insert and Visual modes, and `Ctrl+Q` MUST no longer be bound once `:q` exists.

#### Scenario: Guardar desde Insert
- **WHEN** el usuario pulsa `Ctrl+S` mientras escribe en Insert
- **THEN** el archivo se guarda y el modo sigue siendo Insert
