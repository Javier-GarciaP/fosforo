# Spec Delta

## Purpose
Define el nombre del comando ejecutable y su relación con el nombre de marca.

## ADDED Requirements

### Requirement: Command name
The executable SHALL be named `fos` and the product SHALL keep the brand name *fosforo* in documentation, repository and logo.

#### Scenario: Ejecutar el editor
- **WHEN** una persona escribe `fos` en la terminal tras instalar el programa
- **THEN** se abre el editor sobre el directorio actual

#### Scenario: Binario y módulo
- **WHEN** se compila el proyecto
- **THEN** el punto de entrada está en `cmd/fos` y el binario resultante se llama `fos`

### Requirement: No conflict with common tools
The name `fos` MUST have been checked against the package managers and tools of the developer ecosystem before being adopted, and any
residual overlap SHALL be documented.

#### Scenario: Verificación registrada
- **WHEN** se consulta `design.md` de este cambio
- **THEN** figura el resultado de la comprobación y la decisión tomada
