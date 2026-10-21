# versioning Specification

## Purpose
Define cómo se numeran, anotan y publican las versiones, de forma estándar y predecible, mientras el proyecto está en beta y después.

## Requirements

### Requirement: Semantic versioning in beta
Versions SHALL follow `MAJOR.MINOR.PATCH`; while the project is in beta, `MAJOR` MUST be 0, `MINOR` SHALL increase when a new capability is delivered and `PATCH` when only fixes are delivered.

#### Scenario: Nueva funcionalidad
- **WHEN** se publica el explorador de archivos tras la versión `0.2.0`
- **THEN** la nueva versión es `0.3.0`

#### Scenario: Solo correcciones
- **WHEN** se publica una corrección de errores tras `0.3.0`
- **THEN** la nueva versión es `0.3.1`

### Requirement: Stable release
The version `1.0.0` SHALL be published only when all roadmap changes are archived, CI is green and the end-to-end manual test (open, edit, save, view a Markdown file) passes in a real terminal and inside `tmux`.

#### Scenario: Primera versión estable
- **WHEN** se cumplen todos los criterios de versión estable
- **THEN** se publica `1.0.0` y deja de marcarse como pre-release

### Requirement: Roadmap mapping
Each milestone of the roadmap SHALL map to one minor version, and the mapping MUST be published as a table in `README.md`.

#### Scenario: Consultar el mapa
- **WHEN** alguien lee `README.md`
- **THEN** encuentra una tabla que indica qué entrega cada versión `0.x`

### Requirement: Changelog
Every release SHALL add an entry to `CHANGELOG.md` following the Keep a Changelog format, written before the release pull request is opened.

#### Scenario: Entrada de versión
- **WHEN** se prepara la versión `0.1.0`
- **THEN** `CHANGELOG.md` contiene una sección `[0.1.0]` con la fecha y los cambios agrupados (Añadido, Cambiado, Corregido)

### Requirement: Tags and pre-releases
A release SHALL be a tag `vX.Y.Z` created on `main`, published as a GitHub Release, and MUST be flagged as pre-release while `MAJOR` is 0.

#### Scenario: Publicar una beta
- **WHEN** se publica `v0.1.0`
- **THEN** la etiqueta apunta a un commit de `main` y la Release está marcada como pre-release
