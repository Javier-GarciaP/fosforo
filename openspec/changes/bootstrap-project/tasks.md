# Tasks

## 1. Toolchain y módulo

- [ ] 1.1 Instalar Go estable y verificar con `go version`
- [ ] 1.2 Ejecutar `go mod init` con la ruta de módulo acordada y verificar que existe `go.mod` con `go` y `toolchain` fijados
- [ ] 1.3 Añadir Bubble Tea (y `teatest` como dependencia de pruebas) y verificar con `go mod tidy` sin diferencias

## 2. Esqueleto y arranque

- [ ] 2.1 Crear `cmd/fosforo/` y los paquetes vacíos de `internal/` con `doc.go`; verificar con `go build ./...`
- [ ] 2.2 Implementar el modelo Bubble Tea mínimo (marcador + salir con `q`/`Ctrl+C`); verificar con un test `teatest` que `q` termina el programa
- [ ] 2.3 Implementar `--version`, `--help` y el argumento `<ruta>`; verificar con pruebas de tabla (versión → código 0, flag desconocido → código 2)
- [ ] 2.4 Restaurar la terminal ante pánico y señales; verificar con una prueba que fuerza un pánico y comprueba el estado de la terminal en un pseudo-terminal
- [ ] 2.5 Mostrar el aviso de tamaño mínimo 40x10; verificar con `teatest` enviando mensajes de redimensionado

## 3. Calidad y CI

- [ ] 3.1 Crear `Makefile` con `build`, `test`, `vet` y `run`; verificar ejecutando cada objetivo
- [ ] 3.2 Crear `.github/workflows/ci.yml` con vet, test y build; verificar que el workflow pasa en un PR de prueba a `dev`
- [ ] 3.3 Actualizar `docs/README.md` con los comandos de desarrollo reales; verificar que un clon limpio compila siguiendo solo el README
