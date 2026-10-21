# Tasks

## 1. Renderizado

- [x] 1.1 Añadir Glamour y crear la función `Render(md, ancho, tema) []línea`; verificar con una prueba golden de un documento de ejemplo
- [x] 1.2 Crear los estilos `github-dark` y `github-light` desde la paleta; verificar con golden contra la maqueta 07 (encabezados con regla, código, cita, tareas, tabla)
- [x] 1.3 Alinear el estilo de código con el mapa de colores de `syntax-highlighting`; verificar que un bloque Go usa los mismos colores que el editor
- [x] 1.4 Degradar imágenes (`[imagen: alt]`) y omitir HTML; verificar con pruebas de tabla
- [x] 1.5 Verificar el ancho con CJK y emoji y el ajuste de línea; verificar con golden
- [x] 1.6 Probar con tres READMEs reales (fixtures) y anotar discrepancias; verificar que no hay desbordes de ancho

## 2. Panel de vista previa

- [x] 2.1 Implementar el panel de solo lectura con scroll (`j k Ctrl+d Ctrl+u gg G`) y keymap `preview`; verificar con `teatest`
- [x] 2.2 Implementar el render asíncrono con caché por (versión, ancho, tema) y recálculo al redimensionar; verificar que escribir no se bloquea
- [x] 2.3 Abrir `.md`/`.markdown` en preview por defecto con modo `PREVIEW` en la statusline; verificar con `teatest` y golden contra la maqueta 07
- [x] 2.4 Resolver tablas anchas (truncado o scroll horizontal) según la prueba con READMEs reales; verificar con un fixture de tabla ancha
- [x] 2.5 Desactivar la vista previa para documentos de más de 2 MiB con aviso; verificar con un archivo sintético

## 3. Alternancia y modo dividido

- [x] 3.1 Implementar `markdown.toggle` (`e`, `Space m p`) con aterrizaje del cursor en el encabezado cercano; verificar con `teatest`
- [x] 3.2 Implementar el modo dividido (`Space m s`) con el layout de la maqueta 08; verificar con golden
- [x] 3.3 Extraer las anclas de encabezados del AST y sincronizar el scroll; verificar con pruebas de tabla de mapeo (con y sin encabezados)
- [x] 3.4 Implementar la actualización en vivo con debounce de ~150 ms; verificar con `teatest` y reloj simulado
- [x] 3.5 Actualizar `docs/README.md` (cómo ver y editar Markdown y sus límites); verificar manualmente abriendo `docs/README.md` con el editor
