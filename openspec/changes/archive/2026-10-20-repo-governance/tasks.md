# Tasks

## 1. Repositorio

- [x] 1.1 Proteger `main` y `dev` (PR obligatorio, 0 aprobaciones, sin force-push ni borrado, también administradores); verificado con la API y rechazando un push directo de prueba a `dev` ("Changes must be made through a pull request")
- [x] 1.2 Configurar las opciones de merge (squash y merge commit, sin rebase, borrar ramas al integrar); verificado con la API del repositorio

## 2. Documentación

- [x] 2.1 Escribir `README.md` de presentación con la tabla de versiones de la hoja de ruta; verificar que los enlaces e imágenes relativos existen
- [x] 2.2 Simplificar `CONTRIBUTING.md` (flujo, ramas, reglas y versiones); verificar que los enlaces existen y que coincide con la protección configurada
- [x] 2.3 Crear `CHANGELOG.md` con la entrada de la versión `0.0.1`; verificar que sigue el formato Keep a Changelog
- [x] 2.4 Añadir la columna de versión a la hoja de ruta de `docs/README.md` y ajustar el ejemplo de versión de `distribution`; verificar con `openspec validate --all`

## Workflow follow-up

- Publicar la versión `0.0.1` (PR `dev` → `main`, etiqueta `v0.0.1` y GitHub Release marcada como pre-release) es la primera ejecución de este proceso y se hace después de integrar este cambio.
