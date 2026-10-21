# app-startup Specification

## Purpose
Define cómo arranca, se comporta y termina la aplicación en la terminal, garantizando que la terminal del usuario quede siempre en buen estado.

## Requirements

### Requirement: Terminal state restoration
The application SHALL use the alternate screen while running and MUST restore the terminal (cursor visible, normal screen, echo and line mode) on every exit path, including normal exit, runtime panic, `SIGINT` and `SIGTERM`.

#### Scenario: Salida normal
- **WHEN** el usuario pulsa `q` o `Ctrl+C` en la pantalla inicial
- **THEN** la aplicación termina con código 0 y la terminal queda como estaba antes de ejecutarla

#### Scenario: Pánico durante la ejecución
- **WHEN** ocurre un pánico no controlado mientras la aplicación está abierta
- **THEN** la terminal se restaura, se imprime el error en la salida de errores y el código de salida es distinto de 0

### Requirement: Command-line interface
The application SHALL accept `--version`, `--help` and an optional positional `<path>` argument.

#### Scenario: Mostrar versión
- **WHEN** se ejecuta `fos --version`
- **THEN** imprime la versión en una línea y termina con código 0 sin abrir la pantalla alternativa

#### Scenario: Flag desconocido
- **WHEN** se ejecuta `fos --no-existe`
- **THEN** imprime un mensaje de uso y termina con código 2

#### Scenario: Ruta como argumento
- **WHEN** se ejecuta `fos README.md`
- **THEN** la aplicación recibe la ruta `README.md` para abrirla (el comportamiento de apertura lo define `core-text-editing`)

### Requirement: Minimum terminal size
The application SHALL show a "terminal demasiado pequeña" message while the terminal is smaller than 40 columns by 10 rows, and MUST resume normal rendering once it is resized to at least that size.

#### Scenario: Terminal pequeña
- **WHEN** la terminal mide 30x8
- **THEN** se muestra el aviso con el tamaño mínimo requerido en lugar de la interfaz

#### Scenario: Se agranda la terminal
- **WHEN** la terminal pasa de 30x8 a 80x24
- **THEN** se vuelve a dibujar la interfaz normal sin reiniciar la aplicación
