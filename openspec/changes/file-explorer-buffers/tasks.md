# Tasks

## 1. Modelo de buffers y foco

- [ ] 1.1 Introducir `Buffers`, `Active` y `Focus` en el estado del editor sin cambiar el comportamiento visible; verificar que todas las pruebas existentes pasan
- [ ] 1.2 Guardar cursor y scroll por buffer; verificar con una prueba que cambia de buffer y recupera la posición
- [ ] 1.3 Indexar el keymap por contexto (`editor/*`, `explorer`); verificar con pruebas de tabla que no hay colisiones entre contextos

## 2. Gestión de buffers

- [ ] 2.1 Implementar abrir con reutilización de buffer existente (`:e`, explorador); verificar con una prueba que no se duplican buffers
- [ ] 2.2 Implementar `buffer.next/prev/close` y sus atajos y `:bn :bp :bd`; verificar con `teatest`
- [ ] 2.3 Implementar el cierre con diálogo de cambios sin guardar y el caso del último buffer; verificar con `teatest` y la maqueta 10
- [ ] 2.4 Implementar la salida listando todos los buffers modificados; verificar con dos buffers modificados
- [ ] 2.5 Implementar la detección de cambios externos (mtime y tamaño) en activar, foco y guardar; verificar con una prueba que modifica el archivo en disco
- [ ] 2.6 Mostrar `n/m` en la statusline; verificar con golden

## 3. Árbol de archivos

- [ ] 3.1 Implementar `internal/fs` con carga perezosa, orden e ignorados; verificar con directorios temporales (incluye enlaces simbólicos y permisos)
- [ ] 3.2 Implementar `Create`, `Rename` y `Remove` con errores tipados; verificar con pruebas sobre un directorio temporal
- [ ] 3.3 Implementar la vista del explorador minimalista (sangría de 2 columnas, fila raíz, sin flechas, colores por tipo); verificar con golden contra las maquetas 01 y 07
- [ ] 3.4 Implementar la tabla de iconos y su resolución (nombre exacto → extensión → defecto) con los colores de la paleta; verificar con pruebas de tabla (`trie.go`, `Makefile`, `go.mod`, `notas.xyz`) y golden contra la maqueta 11
- [ ] 3.5 Implementar los iconos de carpeta cerrada/abierta y el modo `icons = "none"` (sufijo `/`, abiertas en azul, cerradas atenuadas); verificar con golden contra la pantalla 11 y comprobando que no se imprime ningún carácter de icono

## 4. Explorador interactivo

- [ ] 4.1 Implementar alternar (`Space e`), foco (`Tab`, `Ctrl+w`) y ancho de 24 columnas con ocultado automático en terminales estrechas; verificar con `teatest`
- [ ] 4.2 Implementar navegación `j k l h Enter` y revelar el archivo actual; verificar con `teatest`
- [ ] 4.3 Implementar `a`, `r`, `d` con prompts y confirmación; verificar con `teatest` sobre un directorio temporal, incluida la actualización de la ruta del buffer
- [ ] 4.4 Implementar el refresco (`R`, guardar, foco); verificar creando un archivo desde fuera
- [ ] 4.5 Actualizar `docs/README.md` (explorador y buffers); verificar que los atajos documentados coinciden con los implementados
