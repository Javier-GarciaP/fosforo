# Spec Delta

## Purpose
Define cómo se protegen las ramas de larga vida y cómo entra el trabajo en ellas, para que el flujo `cambio → dev → main` se cumpla sin depender de la disciplina individual.

## ADDED Requirements

### Requirement: Protected long-lived branches
The branches `main` and `dev` SHALL only accept changes through pull requests: direct pushes (including those of administrators), force pushes and deletions MUST be rejected.

#### Scenario: Push directo rechazado
- **WHEN** alguien, incluido un administrador, intenta subir un commit directamente a `dev`
- **THEN** GitHub lo rechaza con el mensaje "Changes must be made through a pull request" y la rama no cambia

#### Scenario: Integrar por pull request
- **WHEN** se abre un PR hacia `dev` sin conflictos y se integra
- **THEN** la integración se permite sin necesidad de aprobaciones

### Requirement: Self-merge allowed
Branch protection SHALL require zero approving reviews, so that the maintainers can merge their own pull requests, and SHALL require status checks only once a CI workflow exists.

#### Scenario: Mantenedor integra su propio PR
- **WHEN** quien abrió un PR lo integra tras comprobar que no hay conflictos ni comprobaciones fallidas
- **THEN** el PR se integra sin esperar la aprobación de otra persona

### Requirement: Merge strategies
Change branches SHALL be integrated into `dev` with squash merge, `dev` SHALL be integrated into `main` with a merge commit, rebase merges MUST be disabled and merged branches SHALL be deleted automatically.

#### Scenario: Cambio a dev
- **WHEN** se integra un PR de una rama de cambio en `dev`
- **THEN** queda un único commit con el título del PR y la rama se borra

#### Scenario: Release a main
- **WHEN** se integra el PR de `dev` en `main`
- **THEN** se crea un merge commit que conserva la historia de `dev`

### Requirement: Change branches
Every change SHALL be developed in its own branch created from `dev` and named `<tipo>/<nombre-del-cambio>`, and SHALL be closed once its pull request is merged.

#### Scenario: Rama de un cambio
- **WHEN** se empieza el cambio `core-text-editing`
- **THEN** se crea `feat/core-text-editing` desde `dev` y se borra al integrarse

### Requirement: Release only from dev
A pull request into `main` SHALL come only from `dev` (or from a `release/*` branch cut from `dev`) and only when a version is ready to be published.

#### Scenario: PR de publicación
- **WHEN** hay una versión lista
- **THEN** se abre un PR `release: vX.Y.Z` de `dev` a `main`
