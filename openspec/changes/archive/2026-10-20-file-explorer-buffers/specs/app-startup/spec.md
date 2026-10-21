# Spec Delta

## Purpose
La línea de comandos acepta una carpeta (que abre el explorador) y una opción para el modo de iconos.

## MODIFIED Requirements

### Requirement: Command-line interface
The application SHALL accept `--version`, `--help`, `--icons=nerd|none` and an optional positional `<path>` argument that may be a file or a directory.

#### Scenario: Mostrar versión
- **WHEN** se ejecuta `fos --version`
- **THEN** imprime la versión en una línea y termina con código 0 sin abrir la pantalla alternativa

#### Scenario: Flag desconocido
- **WHEN** se ejecuta `fos --no-existe`
- **THEN** imprime un mensaje de uso y termina con código 2

#### Scenario: Ruta como argumento
- **WHEN** se ejecuta `fos README.md`
- **THEN** la aplicación abre `README.md` en el editor

#### Scenario: Carpeta como argumento
- **WHEN** se ejecuta `fos carpeta/`
- **THEN** el explorador se abre con el foco sobre esa carpeta y el editor muestra un buffer vacío sin nombre

#### Scenario: Iconos desactivados
- **WHEN** se ejecuta `fos --icons=none` o con un valor distinto de `nerd` o `none`
- **THEN** con `none` el explorador no imprime iconos y con otro valor se muestra un error de uso y el código de salida es 2
