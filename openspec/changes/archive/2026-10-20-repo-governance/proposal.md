# Proposal

## Why

El repositorio ya existe en GitHub pero cualquiera (incluidos los administradores) podía subir directamente a `main` y `dev`, y no había
una política de versiones. Antes de empezar a implementar hace falta fijar cómo se integra el trabajo y cómo se numeran las versiones,
para que el proceso se cumpla solo y el historial sea legible desde el primer día.

## What Changes

- Proteger `main` y `dev`: todo entra por pull request, sin pushes directos, sin force-push y sin borrado, también para administradores.
  Se exigen **0 aprobaciones** para que quien mantiene el proyecto pueda integrar sus propios PR.
- Ajustes de integración del repositorio: squash para ramas de cambio, merge commit para `dev` → `main`, ramas borradas al integrar.
- Política de versiones: versionado semántico en beta (`0.MINOR.PATCH`) hasta la primera versión estable `1.0.0`, con una tabla que
  relaciona cada hito de la hoja de ruta con su versión.
- `CHANGELOG.md` (formato *Keep a Changelog*) y publicación de cada versión como *pre-release* mientras `MAJOR` sea 0.
- `README.md` de presentación en la raíz y `CONTRIBUTING.md` más corto y directo.

## Capabilities

### New Capabilities
- `versioning`: numeración de versiones, relación con la hoja de ruta, changelog y etiquetas.
- `branch-workflow`: ramas protegidas, estrategias de integración y flujo de publicación a `main`.

### Modified Capabilities

Ninguna. La capacidad `distribution` (de `config-search-release`) sigue definiendo los binarios; aquí solo se fijan las reglas de numeración.

## Non-goals

- Exigir comprobaciones de CI: no se activan hasta que `bootstrap-project` cree el workflow (exigirlas antes bloquearía todos los merges).
- Restringir desde qué rama se puede abrir un PR a `main`: GitHub no lo permite con reglas de rama; se resolverá con una comprobación de CI.
- Elegir la licencia del proyecto.

## Impact

- Configuración del repositorio en GitHub (ramas protegidas y opciones de merge).
- Documentos: `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `docs/README.md`.
