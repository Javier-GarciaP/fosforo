# Tasks

## 1. Configuración

- [ ] 1.1 Implementar la carga TOML con defaults embebidos y la ruta XDG; verificar con pruebas (sin archivo, archivo válido)
- [ ] 1.2 Implementar la validación con errores por clave y respaldo al valor por defecto; verificar con pruebas de tabla (`tabwidth = "ocho"`)
- [ ] 1.3 Aplicar `theme`, `icons`, `tabwidth`, `indent`, `scrolloff` y `mouse` en el editor; verificar con `teatest` (incluye alternar `icons` con `config.reload`)
- [ ] 1.4 Implementar el tema `auto` (OSC 11) con tiempo máximo; verificar con una prueba que simula respuesta y ausencia de respuesta
- [ ] 1.5 Implementar el remapeo por contexto con rechazo de comandos inexistentes; verificar con pruebas de tabla y `teatest`
- [ ] 1.6 Implementar `config.open`, `config.reload` y `Space t`; verificar recargando un archivo editado

## 2. Búsqueda en proyecto

- [ ] 2.1 Implementar el backend `rg --json` con streaming, cancelación y límite de 1000 resultados; verificar con un proyecto de ejemplo
- [ ] 2.2 Implementar el buscador de respaldo con exclusiones; verificar que da los mismos resultados esenciales que `rg` en el proyecto de ejemplo
- [ ] 2.3 Implementar la interfaz (`Space s g`, `Space s w`) reutilizando el picker y abrir en la línea del resultado; verificar con `teatest`

## 3. Distribución

- [ ] 3.1 Inyectar versión y commit con `-ldflags` y mostrarlos en `--version`; verificar con una prueba del binario compilado
- [ ] 3.2 Configurar GoReleaser (Linux y macOS, amd64 y arm64, checksums); verificar con `goreleaser release --snapshot --clean`
- [ ] 3.3 Crear `.github/workflows/release.yml` con la comprobación de que la etiqueta pertenece a `main`; verificar con una etiqueta de prueba fuera de `main` (debe fallar)
- [ ] 3.4 Documentar instalación, configuración, búsqueda y la dependencia opcional de `rg` en `docs/README.md`; verificar siguiéndolo en una máquina limpia
