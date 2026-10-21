# Tasks

## 1. Motor

- [ ] 1.1 Añadir Chroma y definir `Highlighter`, `Span` y las categorías abstractas; verificar con una prueba que tokeniza un fragmento Go conocido
- [ ] 1.2 Implementar la detección de lenguaje por nombre/extensión con texto plano por defecto; verificar con pruebas de tabla
- [ ] 1.3 Implementar el mapa categoría → color para tema oscuro y claro; verificar con una prueba golden contra la guía de estilo (maqueta 00)

## 2. Integración en la vista

- [ ] 2.1 Hacer que `ui` reciba tramos por línea y los pinte conservando fondos de línea actual y selección; verificar con golden contra las maquetas 01 y 03
- [ ] 2.2 Implementar la tokenización asíncrona con versión y descarte de resultados antiguos; verificar con `teatest` que escribir no espera al resaltado
- [ ] 2.3 Implementar el debounce y el desplazamiento de tramos entre ediciones; verificar con una prueba que inserta y borra líneas
- [ ] 2.4 Verificar construcciones multilínea (comentario de bloque, cadena cruda) tras ediciones; verificar con pruebas de tabla
- [ ] 2.5 Aplicar los límites (2 MiB, 2000 caracteres por línea); verificar con archivos sintéticos
- [ ] 2.6 Recolorear al cambiar de tema sin retokenizar; verificar con una prueba que alterna el tema
- [ ] 2.7 Actualizar `docs/README.md` con los lenguajes soportados; verificar que coincide con las pruebas golden
