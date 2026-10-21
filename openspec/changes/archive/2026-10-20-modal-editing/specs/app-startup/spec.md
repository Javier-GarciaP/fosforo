# Spec Delta

## Purpose
La salida normal de la aplicación pasa de `Ctrl+Q`/`Ctrl+C` a la línea de comandos `:q`.

## MODIFIED Requirements

### Requirement: Terminal state restoration
The application SHALL use the alternate screen while running and MUST restore the terminal (cursor visible, normal screen, echo and line mode) on every exit path, including normal exit, runtime panic, `SIGINT` and `SIGTERM`.

#### Scenario: Salida normal
- **WHEN** el usuario ejecuta `:q` sin cambios pendientes de guardar
- **THEN** la aplicación termina con código 0 y la terminal queda como estaba antes de ejecutarla

#### Scenario: Pánico durante la ejecución
- **WHEN** ocurre un pánico no controlado mientras la aplicación está abierta
- **THEN** la terminal se restaura, se imprime el error en la salida de errores y el código de salida es distinto de 0
