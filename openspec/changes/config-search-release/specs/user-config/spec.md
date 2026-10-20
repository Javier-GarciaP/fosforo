# Spec Delta

## Purpose
Define cómo el usuario personaliza el editor mediante un archivo de configuración, temas y remapeo de teclas, sin que los errores impidan arrancar.

## ADDED Requirements

### Requirement: Configuration file
The system SHALL read an optional TOML file at `$XDG_CONFIG_HOME/ted/config.toml` (or `~/.config/ted/config.toml`) over embedded defaults, and MUST start normally when the file does not exist.

#### Scenario: Sin archivo
- **WHEN** no existe el archivo de configuración
- **THEN** el editor arranca con los valores por defecto

### Requirement: Validation without failure
Invalid or unknown settings SHALL be reported in the command line with the setting name and reason, and the editor MUST fall back to the default for that setting instead of failing to start.

#### Scenario: Valor inválido
- **WHEN** el archivo contiene `tabwidth = "ocho"`
- **THEN** se muestra un aviso que nombra `tabwidth` y se usa el valor por defecto 4

### Requirement: Editor settings
The configuration SHALL support `theme` (`dark`, `light` or `auto`), `icons` (`nerd` or `none`, default `nerd`), `tabwidth`, `indent` (`tabs` or `spaces`), `scrolloff` and `mouse`.

#### Scenario: Tema automático
- **WHEN** `theme = "auto"` y la terminal informa de fondo claro
- **THEN** se usa el tema claro; si la terminal no informa, se usa el oscuro

#### Scenario: Desactivar los iconos
- **WHEN** se configura `icons = "none"`
- **THEN** el explorador muestra las carpetas con `/` final y no imprime ningún icono, sin reiniciar tras `config.reload`

### Requirement: Key remapping
The configuration SHALL allow binding and unbinding keys per context (`normal`, `insert`, `visual`, `explorer`, `preview`) to named commands, and MUST report an unknown command name without applying that binding.

#### Scenario: Remapear una tecla
- **WHEN** se configura `normal."gh" = "cursor.line_start"`
- **THEN** pulsar `gh` en Normal mueve el cursor al inicio de la línea

#### Scenario: Comando inexistente
- **WHEN** se asigna un nombre de comando que no existe
- **THEN** se informa del nombre inválido y el resto de atajos sigue funcionando

### Requirement: Reload and open
The commands `config.open` and `config.reload` SHALL open the configuration file in the editor and re-apply it without restarting, and `Space t` MUST toggle between dark and light for the current session.

#### Scenario: Recargar
- **WHEN** se edita el archivo y se ejecuta `config.reload`
- **THEN** los cambios válidos se aplican de inmediato
