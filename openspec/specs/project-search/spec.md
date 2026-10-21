# project-search Specification

## Purpose
Define la búsqueda de texto en todos los archivos del proyecto y la navegación a los resultados.

## Requirements

### Requirement: Live project search
`Space s g` SHALL open a search prompt that lists matches as `path:line: text` while typing, cancels the previous search when the query changes, and limits the list to 1000 results with a notice when truncated.

#### Scenario: Buscar mientras se escribe
- **WHEN** se escribe `Lookup` en la búsqueda de proyecto
- **THEN** aparecen las líneas que contienen `Lookup` con su ruta y número de línea

### Requirement: Jump to result
Pressing `Enter` on a result SHALL open the file at that line with the cursor on the match.

#### Scenario: Abrir resultado
- **WHEN** se elige `internal/keymap/trie.go:9`
- **THEN** el archivo se abre con el cursor en la línea 9

### Requirement: Search word under cursor
`Space s w` SHALL start the project search prefilled with the word under the cursor.

#### Scenario: Palabra bajo el cursor
- **WHEN** el cursor está sobre `Node` y se pulsa `Space s w`
- **THEN** la búsqueda se abre con `Node` ya escrito

### Requirement: Search backend
The system SHALL use `rg` when it is available, respecting ignore files, and otherwise use a built-in search that skips the ignored directories and binary files.

#### Scenario: Sin ripgrep
- **WHEN** `rg` no está instalado
- **THEN** la búsqueda funciona con el buscador integrado, con los mismos resultados esenciales
