# Tasks

## 1. Registro de comandos

- [ ] 1.1 Implementar `Registry` (registrar, buscar por nombre, listar con descripción y atajos); verificar con pruebas de tabla
- [ ] 1.2 Migrar guardar, salir y movimiento del cambio anterior a comandos con nombre; verificar que el comportamiento previo no cambia (pruebas existentes pasan)
- [ ] 1.3 Manejar comandos desconocidos con el mensaje `E492`; verificar con una prueba que el texto no cambia

## 2. Modos y teclas

- [ ] 2.1 Implementar el trie del keymap y `Resolve` con pendiente/completo/no mapeado; verificar con pruebas de tabla (`gg`, `dd`, secuencias parciales)
- [ ] 2.2 Implementar la máquina de modos con etiqueta y color en la statusline; verificar con pruebas golden contra las maquetas 01, 02 y 03
- [ ] 2.3 Implementar conteos (`3dd`, `5j`); verificar con pruebas de tabla
- [ ] 2.4 Implementar el tiempo de espera de 1000 ms con invalidación; verificar con `teatest` y un reloj simulado
- [ ] 2.5 Ignorar teclas no mapeadas en Normal y hacer `Esc` inmediato; verificar con `teatest`

## 3. Edición

- [ ] 3.1 Implementar motions (`h j k l w b e 0 ^ $ gg G {n}G`, `Ctrl+d/u`) devolviendo rangos; verificar con pruebas de tabla con Unicode
- [ ] 3.2 Implementar los operadores `d c y` sobre rangos y `x D C p P`; verificar con pruebas de tabla
- [ ] 3.3 Implementar el registro sin nombre y la copia al portapapeles (OSC 52); verificar el contenido enviado en una prueba y manualmente en tmux
- [ ] 3.4 Implementar deshacer/rehacer con agrupación de sesiones de Insert; verificar con pruebas de ida y vuelta
- [ ] 3.5 Implementar Visual `v`/`V` con resaltado y operadores; verificar con golden contra la maqueta 03 y pruebas de borrado
- [ ] 3.6 Implementar `/`, `n`, `N` con los errores `E486`; verificar con pruebas de tabla

## 4. Línea de comandos

- [ ] 4.1 Implementar el prompt `:` (abrir, editar, cancelar, historial simple); verificar con `teatest`
- [ ] 4.2 Implementar `:w`, `:q`, `:q!`, `:wq`, `:x`, `:e` y `:<n>`; verificar con pruebas de tabla y el diálogo de la maqueta 10
- [ ] 4.3 Actualizar `docs/README.md` con la tabla de atajos real; verificar que cada atajo listado funciona manualmente
