# command-registry Specification

## Purpose
Define el registro central de comandos con nombre, que es el único punto por el que cualquier acción del editor puede invocarse desde teclas, línea de comandos, paleta o mouse.

## Requirements

### Requirement: Single command registry
Every user-invocable action SHALL be a command registered under a unique name of the form `area.verb` and MUST be invocable by name regardless of how it was triggered.

#### Scenario: Mismo comando desde dos orígenes
- **WHEN** se ejecuta `file.save` pulsando `Ctrl+S` y luego escribiendo `:w`
- **THEN** ambos caminos producen exactamente el mismo resultado

### Requirement: Command metadata
Each command SHALL declare a Spanish description and its default key bindings, and the registry SHALL be listable so that other features (paleta, ayuda) can show them.

#### Scenario: Listar comandos
- **WHEN** se pide la lista de comandos
- **THEN** cada entrada incluye nombre, descripción y atajos por defecto

### Requirement: Unknown command handling
Invoking an unregistered command name MUST NOT change editor state and SHALL show an error in the command line.

#### Scenario: Comando inexistente
- **WHEN** el usuario escribe `:nada`
- **THEN** se muestra `E492: no es un comando del editor: nada` y el texto no cambia
