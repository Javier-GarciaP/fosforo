# Spec Delta

## Purpose
Define la ayuda visible de las secuencias de la tecla líder, para aprender atajos sin memorizarlos.

## ADDED Requirements

### Requirement: Leader help panel
When `Space` is pressed in Normal mode and no further key arrives within 300 ms, the editor SHALL show a panel listing the next keys with their descriptions, marking groups with `+`, as in the design.

#### Scenario: Panel tras esperar
- **WHEN** se pulsa `Space` y se espera 300 ms
- **THEN** aparece el panel con `f +archivos`, `e explorador`, `b +buffers`, `s +buscar`, `m +markdown` y el resto de entradas

#### Scenario: Usuario rápido
- **WHEN** se pulsa `Space f f` en menos de 300 ms
- **THEN** el picker se abre sin que el panel llegue a mostrarse

### Requirement: Descending groups
Pressing a group key SHALL replace the panel content with that group's entries, and `Esc` MUST cancel the sequence at any level.

#### Scenario: Entrar en un grupo
- **WHEN** el panel está visible y se pulsa `f`
- **THEN** el panel muestra las acciones de archivos (buscar, abrir reciente, nuevo)

### Requirement: Built from the registry
Every entry MUST correspond to a registered command or to a group, so that the panel stays consistent with the keymap.

#### Scenario: Atajo nuevo
- **WHEN** se registra un comando con atajo `Space x y`
- **THEN** aparece automáticamente en el panel dentro del grupo `x`
