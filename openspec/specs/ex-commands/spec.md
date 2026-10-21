# ex-commands Specification

## Purpose
Define la línea de comandos `:` para guardar, salir, abrir archivos y saltar a una línea.

## Requirements

### Requirement: Command line prompt
Pressing `:` in Normal mode SHALL open a prompt in the command line, run the entered command on Enter and cancel on `Esc`.

#### Scenario: Cancelar
- **WHEN** se escribe `:wq` pero se pulsa `Esc` antes de Enter
- **THEN** no se ejecuta nada y se vuelve a Normal

### Requirement: File commands
The system SHALL implement `:w [path]`, `:q`, `:q!`, `:wq`, `:x` and `:e <path>`.

#### Scenario: Guardar y salir
- **WHEN** se ejecuta `:wq` en un buffer modificado
- **THEN** el archivo se guarda y la aplicación termina con código 0

#### Scenario: Salir sin guardar con aviso
- **WHEN** se ejecuta `:q` con cambios sin guardar
- **THEN** se muestra `E37: cambios sin guardar (usa :q! para descartarlos)` y el diálogo de la maqueta 10; la aplicación no termina

#### Scenario: Descartar cambios
- **WHEN** se ejecuta `:q!`
- **THEN** la aplicación termina sin guardar

### Requirement: Go to line
Entering `:<n>` SHALL move the cursor to line `n`, clamped to the last line.

#### Scenario: Línea fuera de rango
- **WHEN** se ejecuta `:9999` en un archivo de 17 líneas
- **THEN** el cursor va a la línea 17
