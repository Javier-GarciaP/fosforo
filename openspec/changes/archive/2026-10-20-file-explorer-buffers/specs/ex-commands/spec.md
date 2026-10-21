# Spec Delta

## Purpose
`:e` abre archivos en buffers (ya no sustituye el único buffer) y se añaden los comandos de buffers.

## MODIFIED Requirements

### Requirement: File commands
The system SHALL implement `:w [path]`, `:q`, `:q!`, `:wq`, `:x`, `:e [path]`, `:bn`, `:bp` and `:bd [!]`, where `:e <path>` opens the file in a buffer (reusing it if it is already open) and `:e` without a path reloads the current file.

#### Scenario: Guardar y salir
- **WHEN** se ejecuta `:wq` en un buffer modificado y no hay otros buffers modificados
- **THEN** el archivo se guarda y la aplicación termina con código 0

#### Scenario: Salir sin guardar con aviso
- **WHEN** se ejecuta `:q` con cambios sin guardar
- **THEN** se muestra `E37: cambios sin guardar (usa :q! para descartarlos)` y el diálogo de la maqueta 10; la aplicación no termina

#### Scenario: Descartar cambios
- **WHEN** se ejecuta `:q!`
- **THEN** la aplicación termina sin guardar

#### Scenario: Abrir otro archivo sin perder cambios
- **WHEN** se ejecuta `:e otro.txt` con cambios sin guardar en el buffer actual
- **THEN** `otro.txt` se abre en un buffer nuevo y el buffer anterior conserva sus cambios

#### Scenario: Recargar con cambios
- **WHEN** se ejecuta `:e` sin ruta en un buffer modificado
- **THEN** se muestra `E37: cambios sin guardar (usa :e! para descartarlos)` y el buffer no cambia; con `:e!` se recarga el archivo desde disco
