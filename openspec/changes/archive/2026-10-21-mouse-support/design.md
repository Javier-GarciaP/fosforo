# Design

## Context

Bubble Tea entrega eventos de mouse (pulsar, soltar, movimiento, rueda) con coordenadas de celda. Los paneles (explorador, editor,
preview, overlays) ya existen y se dibujan con funciones puras.

## Goals / Non-Goals

**Goals:**
- Que dibujar y localizar un clic usen la misma geometría (sin duplicar cálculos).
- Que el mouse ejecute comandos del registro.

**Non-Goals:**
- Gestos complejos o interacción con barras de scroll.

## Decisions

- **Geometría compartida:** `ui.ComputeLayout(size, estado) Layout` devuelve rectángulos de cada panel; `View` los usa para dibujar y `Update` para
  localizar el panel bajo el puntero. Una sola fuente de verdad evita desfases.
- **Pantalla → texto:** función inversa de la columna visual (gutter, tabs a 4, anchos) que devuelve (línea, grafema); se prueba con las mismas tablas que el cursor.
- **Doble clic:** dos pulsaciones en la misma celda dentro de 400 ms. La palabra se obtiene con la misma definición de palabra que `w`/`b`.
- **Arrastre:** el evento de movimiento con botón pulsado se limita a la tasa de dibujo (coalescencia) para no saturar.
- **Eventos como comandos:** `cursor.move_to(línea, col)`, `view.scroll(panel, n)`, `visual.select_range(a, b)`, `explorer.open(ruta)`, `overlay.click(i)`.
- **Captura configurable:** activada por defecto; `mouse.toggle` quita la captura para recuperar la selección nativa del terminal.

- **Decisiones tomadas al implementar:**
  - `ui.ComputeLayout` (zonas: explorador, editor, vista previa, statusline, línea de comandos) y `ui.PickerGeom`/`ui.DialogGeom` (caja, filas y botones) son la única fuente de verdad: lo que se dibuja
    y lo que se pulsa salen de las mismas funciones (`Areas`, `pickerRect`, `DialogGeom`), y una prueba comprueba que coinciden. `ui.PosAt` es la conversión pantalla → (línea, grafema) y una prueba
    la compara con `CursorCell` para cada posición.
  - Los eventos se traducen a comandos del registro (`cursor.move_to`, `visual.select_range`, `view.scroll`, `explorer.click`, `mouse.toggle`) con posiciones en `command.Args` (`Line/Col/Line2/Col2`).
    `Open`/aceptar en picker y diálogos reutilizan `pickerAccept` y `chooseDialog`, las mismas rutas que el teclado.
  - Doble clic: dos pulsaciones en la misma celda en menos de 400 ms (`DoubleClickInterval`, con reloj inyectable); un tercer clic empieza otra cuenta. La palabra usa `charClass` (la de `w`/`b`).
  - Arrastre: `MouseModeCellMotion` solo informa de movimiento con el botón pulsado y se ignoran las muestras repetidas de la misma celda (Bubble Tea ya limita el dibujo a su ritmo).
  - Rueda: 3 líneas por muesca en el panel bajo el puntero. En el editor se mueve la vista y el cursor se acerca a lo visible respetando el margen de 3 líneas; en el explorador solo cambia el desplazamiento
    (la selección no se mueve y, mientras no cambie, `scrollExplorer` no la vuelve a traer a la vista); en la vista previa completa se desplaza la vista previa y en el modo dividido el fuente (la vista previa lo sigue);
    sobre el picker mueve la selección. Con un diálogo abierto la rueda y los clics fuera de los botones se ignoran (es modal).
  - Un clic fuera de la ventana de un picker o paleta la cierra sin ejecutar nada; un clic fuera de un diálogo no hace nada.
  - Pregunta abierta resuelta: un clic en el explorador **abre** el archivo (y alterna las carpetas), como en VS Code; la raíz no se cierra.
  - `mouse.toggle` (también `:mouse` y la paleta) cambia `View.MouseMode` entre `CellMotion` y `None`, de modo que el terminal recupera el ratón y la selección nativa funciona. Con el ratón activado se puede
    seguir seleccionando de forma nativa con Shift+arrastrar en la mayoría de terminales y en `tmux` (con `mouse on` de tmux, Shift o la tecla del terminal).

## Risks / Trade-offs

- [Con la captura activa, la selección nativa del terminal exige Shift+arrastrar] → documentarlo y ofrecer `mouse.toggle`.
- [Terminales o multiplexores sin mouse SGR] → degradar a teclado sin errores.
- [Clic en medio de un carácter ancho] → resolver al carácter que contiene la celda.
- [Eventos de movimiento masivos] → coalescer por fotograma.

## Open Questions

- ¿Un clic en el explorador abre el archivo (como en VS Code) o solo lo selecciona hasta el doble clic? Se asume apertura con un clic.
