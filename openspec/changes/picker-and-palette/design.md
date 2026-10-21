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

## Risks / Trade-offs

- [Recorte de ANSI con caracteres anchos produce artefactos] → pruebas golden con CJK y estilos mezclados.
- [Listas enormes y puntuación en cada pulsación] → puntuar en segundo plano y limitar el número de resultados visibles (p. ej. 200).
- [Preview de archivos grandes o binarios] → leer solo las primeras ~100 líneas y omitir binarios.
- [Dos temporizadores distintos (1000 ms de secuencia y 300 ms de ayuda) se confunden] → documentarlos en un solo lugar y probar ambos.

## Open Questions

- Orden inicial del picker de archivos: ¿priorizar archivos abiertos recientemente? Se asume orden por puntuación y, a igualdad, ruta más corta.
