# Spec Delta

## ADDED Requirements

### Requirement: One-line installation
The project SHALL provide an installation script that downloads the binary of the latest release for the current system, verifies its SHA256 checksum, installs it in a user-writable directory and tells the person how to put it on the `PATH`; `README.md` MUST document it together with `go install`, manual download, updating and uninstalling.

#### Scenario: Instalar con un comando
- **WHEN** una persona ejecuta el comando de instalación del README en Linux o macOS
- **THEN** `fos --version` funciona sin más pasos que los documentados

#### Scenario: Suma incorrecta
- **WHEN** la suma SHA256 de la descarga no coincide con `checksums.txt`
- **THEN** el script no instala nada y termina con error
