# Spec Delta

## MODIFIED Requirements

### Requirement: Editor settings
The configuration SHALL support `theme` (`dark`, `light` or `auto`), `icons` (`nerd` or `none`, default `none`), `tabwidth`, `indent` (`tabs` or `spaces`), `scrolloff` and `mouse`.

#### Scenario: Tema automático
- **WHEN** `theme = "auto"` y la terminal informa de fondo claro
- **THEN** se usa el tema claro; si la terminal no informa, se usa el oscuro

#### Scenario: Iconos desactivados por defecto
- **WHEN** no hay configuración ni `--icons`
- **THEN** el explorador muestra las carpetas con `/` final y no imprime ningún icono

#### Scenario: Activar los iconos
- **WHEN** se configura `icons = "nerd"` o se ejecuta `fos --icons=nerd`
- **THEN** el explorador muestra los iconos de Nerd Font, sin reiniciar tras `config.reload`

#### Scenario: Desactivar los iconos
- **WHEN** se configura `icons = "none"`
- **THEN** el explorador muestra las carpetas con `/` final y no imprime ningún icono, sin reiniciar tras `config.reload`
