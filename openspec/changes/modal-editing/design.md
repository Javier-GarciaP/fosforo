# Design

## Context

`core-text-editing` aporta `buffer` (con `Apply(Edit)`), `ui` y un `editor.Model` con atajos provisionales. Este cambio
introduce la capa de intención (comandos, modos, teclas) entre la entrada de Bubble Tea y el buffer.

## Goals / Non-Goals

**Goals:**
- Que añadir un comando nuevo sea registrar una función y una tabla de atajos, sin tocar el bucle principal.
- Que operador + motion funcione para cualquier motion futura sin código extra.

**Non-Goals:**
- Compatibilidad total con Vim: se implementa un subconjunto útil y documentado.

## Decisions

- **Registro de comandos (`internal/command`):** `Registry` con `Register(Command{Name, Desc, Run})`. `Run` recibe el estado del
  editor y devuelve un efecto (nuevo estado y `tea.Cmd` opcional) para que sea testable sin terminal.
- **Keymap (`internal/keymap`):** trie por modo; `Resolve(mode, pendiente, tecla)` devuelve `{Comando, Conteo, Pendiente|Completo|NoMapeado}`.
  Las tablas por defecto van embebidas en código y luego serán sobreescribibles desde configuración.
- **Operador + motion = rango:** toda motion devuelve un rango (inicio, fin, lineal/carácter); los operadores actúan sobre rangos y
  Visual simplemente aporta un rango ya seleccionado. Así `d`, `c`, `y` funcionan con todas.
- **Deshacer:** pila de grupos de `Edit` inversos en `buffer`; el modo Insert abre un grupo al entrar y lo cierra al salir.
- **Tiempo de espera:** `tea.Tick` de 1000 ms que se invalida con un contador de secuencia (evita carreras).
- **Portapapeles:** secuencia OSC 52 (funciona por SSH/tmux) con respaldo a la utilidad del sistema si existe.
- **Búsqueda:** `regexp` de Go sobre el buffer; sin resaltado persistente en v1.

- **Decisiones tomadas al implementar:**
  - `command.Registry` es genérico en el estado (`Registry[*editor.Model]`) para que `command` no dependa de `editor`; los comandos de `editor` solo
    devuelven un `tea.Cmd` cuando hace falta (portapapeles, temporizador): salir es un indicador del modelo (`Quitting()`), lo que simplifica las pruebas.
  - Las motions son una tabla (`motionDefs`) de la que salen a la vez el comando, las teclas de Normal/Visual/operador y el rango de los operadores.
    El árbol de teclas «operator» solo contiene motions; `dd`/`cc`/`yy` se detectan repitiendo la tecla del operador.
  - Reglas de Vim aplicadas: motion exclusiva que acaba en la columna 0 termina en la línea anterior (así `dw` no une líneas), `cw` actúa como `ce`
    y el cursor en Normal nunca queda después del último carácter.
  - Deshacer: `buffer` guarda grupos de `Edit`; `editor.withUndo` abre un grupo por orden y lo mantiene abierto mientras el modo es Insert. «Modificado»
    se calcula por versión (ediciones aplicadas) frente a la versión guardada, así deshacer hasta lo guardado quita `[+]`.
  - Temporizador: cada tecla incrementa un contador y el `tea.Tick` lleva el valor; un `TimeoutMsg` viejo se ignora.
  - `Ctrl+C` actúa como `Esc` (no sale) y `Ctrl+Q` queda sin asignar. En Insert las flechas cruzan de línea (como en VS Code).
  - Portapapeles: se envía OSC 52 y, si existe, `wl-copy`/`xclip`/`xsel`/`pbcopy`; solo `y` copia (borrar no toca el portapapeles).
  - La búsqueda es por línea (no hay coincidencias entre líneas); `:e` sustituye el único buffer hasta que lleguen los buffers múltiples.
  - Fuera de la lista cerrada se dejó a propósito: `X`, `s`, `?`, `.`, objetos de texto y `p` en Visual.

## Risks / Trade-offs

- [Teclados en español: teclas muertas (`~`, `` ` ``, `^`) y `<`/`>` en posiciones distintas] → probar con distribución española desde el inicio; no depender de combinaciones raras.
- [`Esc` y secuencias de escape de flechas se confunden en algunas terminales] → usar el análisis de teclas de Bubble Tea y probar en tmux y SSH.
- [Tiempo de espera molesto en secuencias cortas] → solo se espera si existe un prefijo posible; configurable en el futuro.
- [Alcance de Vim muy grande] → lista cerrada de comandos en este cambio; lo demás se anota como futuro.

## Open Questions

- Si el uso real lo pide, `.` (repetir) y objetos de texto serán el siguiente cambio candidato.
