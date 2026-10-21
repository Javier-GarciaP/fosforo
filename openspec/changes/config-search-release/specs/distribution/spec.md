# Spec Delta

## Purpose
Define cómo se construye, versiona y publica el editor para que cualquier persona pueda instalarlo de forma reproducible.

## ADDED Requirements

### Requirement: Versioned binaries
Each release SHALL publish static binaries for Linux and macOS on amd64 and arm64, with SHA256 checksums, and `fosforo --version` MUST print the semantic version and short commit.

#### Scenario: Versión impresa
- **WHEN** se ejecuta `fosforo --version` en un binario de release
- **THEN** imprime algo como `fosforo 0.8.0 (abc1234)`

### Requirement: Installation from source
The project SHALL be installable with `go install` from the repository.

#### Scenario: Instalar con Go
- **WHEN** se ejecuta `go install <módulo>/cmd/fosforo@latest` con Go instalado
- **THEN** el comando `fosforo` queda disponible

### Requirement: Release tied to main
A release MUST be created only from a tag `vX.Y.Z` on `main`, following the branching workflow in `CONTRIBUTING.md`, and the release workflow SHALL fail if the tag does not point to a commit on `main`.

#### Scenario: Etiqueta fuera de main
- **WHEN** se crea una etiqueta de versión sobre un commit que no está en `main`
- **THEN** el workflow de release falla sin publicar nada

### Requirement: Installation documentation
`docs/README.md` SHALL document installation from releases and from source, and the optional dependency on `rg`.

#### Scenario: Instrucciones completas
- **WHEN** una persona sigue el README en una máquina limpia
- **THEN** consigue ejecutar `fosforo` sin pasos no documentados
