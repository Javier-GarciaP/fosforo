# Design

## Context

Ya existen registro de comandos, keymap con prefijos (incluido `Space`), lista de buffers, árbol de archivos y resaltado. Falta una capa
de interfaz que se superponga a la vista base y capture la entrada mientras está abierta.

## Goals / Non-Goals

**Goals:**
- Un único componente `Picker` genérico para archivos, buffers y comandos.
- Superposición (overlay) reutilizable también por diálogos.

**Non-Goals:**
- Reordenación por uso frecuente o historial persistente.

## Decisions

- **Componente genérico:** `ui.Picker[T]` con `Source` (síncrona o asíncrona por canal), `Scorer`, `Preview` y `OnAccept`. Los tres usos solo cambian la fuente y la acción.
- **Coincidencia difusa:** subsecuencia con puntuación estilo fzf (bonus por inicio de palabra, tras `/` y consecutivos). Se empieza con una librería probada;
  si no ordena bien, se escribe una propia detrás de la interfaz `Scorer`.
- **Superposición:** `ui.Overlay(base, popup, x, y)` que recorta cadenas ANSI respetando el ancho de celdas (Lip Gloss/`x/ansi`); el atenuado del fondo se logra
  volviendo a pintar la vista base con la variante "atenuada" de la paleta, porque una terminal no mezcla transparencias.
- **Entrada modal:** mientras hay un overlay abierto, `Update` lo enruta primero; el keymap del editor queda suspendido.
- **Carga de archivos:** `fs.Walk` en una goroutine que emite lotes de resultados como mensajes; respeta la lista de ignorados del explorador.
- **Menú líder:** el temporizador de 300 ms reutiliza el mecanismo de tiempo de espera del keymap; el contenido se genera recorriendo el trie bajo el prefijo actual.

- **Decisiones tomadas al implementar:**
  - Un solo `picker` en `editor` (y `ui.PickerView` en la vista) sirve para archivos, buffers, recientes, paleta y atajos; cambian la fuente y la acción al aceptar.
    No se usó genérico `Picker[T]`: los elementos son siempre `pickItem` (texto, texto de comparación, detalle y dato).
  - `internal/fuzzy`: subsecuencia con programación dinámica y puntuación tipo fzf (bonus de inicio de palabra, tras `/` o `_`, camelCase, caracteres seguidos y
    coincidencia en el nombre de archivo); *smart-case*; empate por texto más corto y luego orden original. Implementación propia (sin dependencia nueva).
    Medido: ~160 ms para 50 000 rutas, por eso las listas de más de 3000 elementos se puntúan **en segundo plano** (con una sola puntuación a la vez y descarte de resultados viejos).
  - Sin texto escrito, archivos se ordenan por ruta más corta; buffers, recientes y comandos conservan su orden original.
  - `fs.Walk` recorre el proyecto en una goroutine y envía lotes de 500 por un canal; el picker pide el siguiente lote con un `tea.Cmd`. Cerrar el picker cancela el recorrido.
  - La superposición se dibuja sobre la misma cuadrícula de celdas de `ui` (no hace falta recortar cadenas ANSI): `grid.dim` mezcla los colores del fondo con su color de
    fondo para atenuarlo, y la ventana se pinta encima. Caracteres anchos y estilos mezclados quedan cubiertos por las pruebas.
  - El contador de la ventana es **coincidencias/total** (`5/42`), como en la maqueta 05.
  - Vista previa: primeras 100 líneas (máx. 64 KiB leídos), coloreada con `syntax`, con nota para binarios, no UTF-8 y archivos ilegibles.
  - Menú líder: dos temporizadores distintos, definidos juntos en `model.go`: **300 ms** (`LeaderDelay`) para mostrar el panel y **1000 ms** (`DefaultTimeout`) para que
    caduque una secuencia incompleta. Una vez visible el panel, la espera **no caduca** (se cierra con `Esc`, una tecla sin asignar o completando la secuencia).
    El contenido sale de `keymap.Children` + el registro (`leaderEntries`); `Command.Short` da la etiqueta corta (si falta, se usa la descripción).
    Una prueba comprueba que todo lo que cuelga de `Space` está registrado y que todo atajo `Space …` declarado está en el árbol.
  - Sugerencias en `:`: las mismas entradas de la paleta, que también comparan con los atajos (escribir `wq` sugiere «Guardar y salir»). Con sugerencias a la vista `↑`/`↓` las recorren
    (y `Enter` ejecuta la elegida); con el prompt vacío o al navegar el historial siguen recorriendo el historial.
  - Comandos nuevos: `file.find`, `file.recent` (archivos abiertos en la sesión), `file.new` (abre `:e `), `buffer.pick`, `palette.open`, `help.keys`. Los comandos que necesitan contexto
    (explorador, modo Insert, selección Visual) no hacen nada —o solo avisan— si se lanzan desde la paleta sin él.
  - `registry`/`defaultKeymap` se construyen en `init()` para evitar un ciclo de inicialización (la paleta consulta el registro).

## Risks / Trade-offs

- [Recorte de ANSI con caracteres anchos produce artefactos] → pruebas golden con CJK y estilos mezclados.
- [Listas enormes y puntuación en cada pulsación] → puntuar en segundo plano y limitar el número de resultados visibles (p. ej. 200).
- [Preview de archivos grandes o binarios] → leer solo las primeras ~100 líneas y omitir binarios.
- [Dos temporizadores distintos (1000 ms de secuencia y 300 ms de ayuda) se confunden] → documentarlos en un solo lugar y probar ambos.

## Open Questions

- Orden inicial del picker de archivos: ¿priorizar archivos abiertos recientemente? Se asume orden por puntuación y, a igualdad, ruta más corta.
