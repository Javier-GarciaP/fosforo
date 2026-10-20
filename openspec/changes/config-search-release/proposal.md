# Proposal

## Why

Para ser usable a diario, el editor necesita personalización básica (tema, teclas, tabulación), búsqueda de texto en todo el proyecto
y una forma estable de instalarse. Es el cierre de la v1 y enlaza con el flujo de ramas: un PR de `dev` a `main` genera una versión.

## What Changes

- Archivo de configuración TOML con valores por defecto embebidos, validación con mensajes claros y recarga.
- Tema oscuro, claro y automático; remapeo de teclas; ancho de tabulación y sangría.
- Búsqueda de texto en el proyecto (`Space s g`, `Space s w`) con `rg` si existe y un buscador propio de respaldo.
- Distribución: binarios para Linux y macOS, `go install`, versión con commit y un workflow de release al etiquetar `main`.

## Capabilities

### New Capabilities
- `user-config`: configuración del usuario, temas y remapeo de teclas.
- `project-search`: búsqueda de texto en el proyecto con navegación a resultados.
- `distribution`: construcción, versionado y publicación del binario.

### Modified Capabilities

Ninguna en `openspec/specs/` aún; al implementarlo se amplían los comandos del registro con `config.*` y `search.*`.

## Non-goals

- Plugins, scripting o configuración en Lua.
- Reemplazo de texto en todo el proyecto.
- Soporte oficial de Windows en v1 y firma/notarización de macOS.
- Gestores de paquetes (Homebrew, AUR) más allá de documentar la instalación.

## Impact

- Código nuevo: `internal/config`, `internal/search`; flujo de release en `.github/workflows/release.yml` y configuración de GoReleaser.
- Dependencias: lector TOML; opcionalmente `rg` en el sistema.
- Depende de: todos los cambios anteriores.
