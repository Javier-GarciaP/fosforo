# distribution Specification

## Purpose
Define cómo se construye, versiona y publica el editor para que cualquier persona pueda instalarlo de forma reproducible.

## Requirements

### Requirement: Versioned binaries
Each release SHALL publish static binaries for Linux and macOS on amd64 and arm64, with SHA256 checksums, and `fos --version` MUST print the semantic version and short commit.

#### Scenario: Versión impresa
- **WHEN** se ejecuta `fos --version` en un binario de release
- **THEN** imprime algo como `fos 0.8.0 (abc1234)`

### Requirement: Installation from source
The project SHALL be installable with `go install` from the repository.

#### Scenario: Instalar con Go
- **WHEN** se ejecuta `go install <módulo>/cmd/fos@latest` con Go instalado
- **THEN** el comando `fos` queda disponible

### Requirement: Release tied to main
A release MUST be created only from a tag `vX.Y.Z` on `main`, following the branching workflow in `CONTRIBUTING.md`, and the release workflow SHALL fail if the tag does not point to a commit on `main`.

#### Scenario: Etiqueta fuera de main
- **WHEN** se crea una etiqueta de versión sobre un commit que no está en `main`
- **THEN** el workflow de release falla sin publicar nada

### Requirement: Installation documentation
`docs/README.md` SHALL document installation from releases and from source, and the optional dependency on `rg`.

#### Scenario: Instrucciones completas
- **WHEN** una persona sigue el README en una máquina limpia
- **THEN** consigue ejecutar `fos` sin pasos no documentados

### Requirement: One-line installation
The project SHALL provide an installation script that downloads the binary of the latest release for the current system, verifies its SHA256 checksum, installs it in a user-writable directory and tells the person how to put it on the `PATH`; `README.md` MUST document it together with `go install`, manual download, updating and uninstalling.

#### Scenario: Instalar con un comando
- **WHEN** una persona ejecuta el comando de instalación del README en Linux o macOS
- **THEN** `fos --version` funciona sin más pasos que los documentados

#### Scenario: Suma incorrecta
- **WHEN** la suma SHA256 de la descarga no coincide con `checksums.txt`
- **THEN** el script no instala nada y termina con error
