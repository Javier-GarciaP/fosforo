# Design

## Context

El repositorio es público, lo mantiene una sola cuenta y las integraciones las hace esa misma cuenta (con ayuda de Claude Code). Hay que
proteger las ramas sin que el flujo dependa de aprobaciones ajenas.

## Goals / Non-Goals

**Goals:**
- Que sea imposible saltarse el flujo por accidente, también siendo administrador.
- Que las versiones sigan un estándar reconocible (SemVer + Keep a Changelog).

**Non-Goals:**
- Reglas de CI y de calidad (llegan con `bootstrap-project`).

## Decisions

- **Protección clásica de ramas** (no *rulesets*): cubre lo necesario con `enforce_admins`, 0 aprobaciones, sin force-push ni borrado.
  Un administrador puede desactivar la regla temporalmente, pero no saltársela por descuido.
- **0 aprobaciones:** nadie puede aprobar su propio PR en GitHub; exigir una aprobación bloquearía al único mantenedor.
- **Comprobaciones de CI obligatorias solo cuando existan:** si se exigen antes, GitHub esperaría una comprobación que nunca llega.
- **Squash a `dev`, merge commit a `main`:** `dev` queda con un commit por cambio (legible); `main` conserva la historia de `dev` para poder
  etiquetar y revertir versiones enteras. Se desactiva el rebase merge para no tener una tercera forma.
- **SemVer con `0.x`:** es el estándar para software aún inestable; cada hito pasa a una versión `MINOR` y `1.0.0` marca la estabilidad.
  Se descartó añadir sufijos `-beta.N` porque `0.x` ya comunica inestabilidad y duplicaría la información.
- **Release como pre-release** mientras `MAJOR` = 0, para que GitHub no la presente como "última versión estable".

## Risks / Trade-offs

- [Cualquier persona con permisos de administrador puede relajar la protección] → documentarlo y revisar la configuración al añadir colaboradores.
- [Un PR a `main` desde una rama distinta de `dev` no se bloquea] → añadir una comprobación de CI en `bootstrap-project` que lo verifique.
- [Un squash pierde la historia detallada de cada rama] → aceptado: el PR conserva la discusión y los commits.
