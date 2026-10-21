# Tasks

## 1. Componente base

- [ ] 1.1 Implementar el `Scorer` difuso con resaltado de coincidencias; verificar con pruebas de tabla (`ktr` → `keymap/trie.go`)
- [ ] 1.2 Implementar `Overlay` con recorte ANSI consciente del ancho y el fondo atenuado; verificar con golden (CJK y estilos mezclados)
- [ ] 1.3 Implementar `Picker[T]` (entrada, lista, selección, `Enter`/`Esc`, contador); verificar con `teatest`
- [ ] 1.4 Enrutar la entrada al overlay mientras esté abierto; verificar que el keymap del editor no recibe teclas

## 2. Picker de archivos y buffers

- [ ] 2.1 Implementar la fuente de archivos asíncrona con lotes e ignorados; verificar con un directorio sintético de muchos archivos que la entrada responde
- [ ] 2.2 Implementar `Space f f` y abrir el archivo elegido; verificar con `teatest`
- [ ] 2.3 Implementar la vista previa coloreada con límite de líneas y omisión de binarios; verificar con golden contra la maqueta 05
- [ ] 2.4 Implementar `Space b b` marcando buffers modificados; verificar con `teatest`

## 3. Paleta y menú líder

- [ ] 3.1 Implementar la paleta sobre el registro con atajos a la derecha; verificar con golden contra la maqueta 06
- [ ] 3.2 Implementar sugerencias en el prompt `:` y la regla de `Enter`; verificar con `teatest` (`:wq` sin elegir sugerencia)
- [ ] 3.3 Implementar el panel del menú líder con temporizador de 300 ms generado desde el trie; verificar con golden contra la maqueta 04 y con `teatest` y reloj simulado
- [ ] 3.4 Implementar grupos descendentes y `Esc`; verificar con `teatest`
- [ ] 3.5 Actualizar `docs/README.md` con el menú líder real; verificar que el panel y la documentación coinciden
