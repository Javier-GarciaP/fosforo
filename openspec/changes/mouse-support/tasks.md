# Tasks

## 1. Geometría y conversión

- [x] 1.1 Extraer `ComputeLayout` y hacer que la vista lo use; verificar que las pruebas golden existentes no cambian
- [x] 1.2 Implementar pantalla → (línea, grafema) con gutter, tabs, anchos y scroll; verificar con pruebas de tabla (incluye CJK y clic bajo el texto)
- [x] 1.3 Implementar la búsqueda del panel bajo el puntero; verificar con pruebas de tabla de coordenadas

## 2. Editor

- [x] 2.1 Implementar clic para mover el cursor (`cursor.move_to`); verificar con `teatest` enviando eventos de mouse
- [x] 2.2 Implementar arrastre → Visual y doble clic → palabra; verificar con `teatest` (arrastre y copia con `y`)
- [x] 2.3 Implementar la rueda por panel con cursor siempre visible; verificar con `teatest` sobre editor, explorador y preview

## 3. Explorador y overlays

- [x] 3.1 Implementar clics en el explorador (abrir/expandir); verificar con `teatest`
- [x] 3.2 Implementar clics en picker, paleta y diálogos, y cierre al hacer clic fuera; verificar con `teatest`
- [x] 3.3 Implementar `mouse.toggle`; verificar que al desactivarlo no se reciben eventos y manualmente que la selección nativa funciona en tmux y SSH
- [x] 3.4 Actualizar `docs/README.md` (mouse y selección nativa con Shift); verificar la documentación en un terminal real
