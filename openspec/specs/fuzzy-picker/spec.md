# fuzzy-picker Specification

## Purpose
Define la selección rápida de elementos (archivos y buffers) escribiendo parte de su nombre, con resultados ordenados y vista previa.

## Requirements

### Requirement: Fuzzy filtering
The picker SHALL filter its items as the user types, matching the typed characters as an ordered subsequence of the item text, ranking tighter and path-aware matches first and highlighting the matched characters.

#### Scenario: Subsecuencia
- **WHEN** se escribe `ktr` con los archivos `internal/keymap/trie.go` y `docs/README.md`
- **THEN** solo `internal/keymap/trie.go` aparece, con `k`, `t` y `r` resaltadas

### Requirement: Keyboard control
The picker SHALL move the selection with `Up`/`Down` or `Ctrl+p`/`Ctrl+n`, accept with `Enter` and cancel with `Esc`, restoring the previous state when cancelled.

#### Scenario: Cancelar
- **WHEN** se abre el picker y se pulsa `Esc`
- **THEN** el picker se cierra sin abrir nada y el cursor del editor no cambia

### Requirement: File picker
`Space f f` SHALL open a picker listing the project files, excluding the ignored directories, and open the chosen file in the editor.

#### Scenario: Abrir por nombre
- **WHEN** se pulsa `Space f f`, se escribe `trie` y se pulsa `Enter`
- **THEN** se abre `internal/keymap/trie.go`

### Requirement: Buffer picker
`Space b b` SHALL list the open buffers, marking the modified ones, and switch to the chosen one.

#### Scenario: Cambiar de buffer
- **WHEN** se elige el segundo buffer de la lista
- **THEN** ese buffer pasa a ser el activo

### Requirement: Preview and counters
The file picker SHALL show a syntax-highlighted preview of the selected file and a `selected/total` counter, as in the design.

#### Scenario: Vista previa
- **WHEN** se selecciona un archivo en la lista
- **THEN** su contenido inicial aparece coloreado en el panel derecho

### Requirement: Responsive on large projects
The picker MUST open immediately even when the file list is still loading, completing the list in the background without blocking input.

#### Scenario: Proyecto con muchos archivos
- **WHEN** se abre el picker en un proyecto de decenas de miles de archivos
- **THEN** la entrada responde de inmediato y la lista se completa progresivamente
