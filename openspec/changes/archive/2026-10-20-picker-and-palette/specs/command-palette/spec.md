# Spec Delta

## Purpose
Define la paleta que permite buscar y ejecutar cualquier comando del editor por su nombre o descripción, mostrando su atajo.

## ADDED Requirements

### Requirement: Palette listing
`Space Space` SHALL open a palette listing every registered command by its Spanish description with its key binding aligned to the right, filtered fuzzily as the user types.

#### Scenario: Buscar guardar
- **WHEN** se abre la paleta y se escribe `gua`
- **THEN** aparecen los comandos de guardar, con `Gua` resaltado y sus atajos (`:w  Space w`)

### Requirement: Palette execution
`Enter` SHALL run the selected command through the command registry, exactly as if it had been triggered by its key binding.

#### Scenario: Ejecutar desde la paleta
- **WHEN** se selecciona "Guardar archivo" y se pulsa `Enter`
- **THEN** el archivo se guarda igual que con `Ctrl+S`

### Requirement: Suggestions in the command line
While typing in the `:` prompt, the editor SHALL show the same list as suggestions; `Enter` runs the highlighted suggestion if one was chosen with `Up`/`Down`, otherwise it runs the text as typed.

#### Scenario: Escribir comando completo
- **WHEN** se escribe `:wq` y se pulsa `Enter` sin elegir ninguna sugerencia
- **THEN** se ejecuta `:wq`
