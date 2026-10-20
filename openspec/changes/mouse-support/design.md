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

## Risks / Trade-offs

- [Con la captura activa, la selección nativa del terminal exige Shift+arrastrar] → documentarlo y ofrecer `mouse.toggle`.
- [Terminales o multiplexores sin mouse SGR] → degradar a teclado sin errores.
- [Clic en medio de un carácter ancho] → resolver al carácter que contiene la celda.
- [Eventos de movimiento masivos] → coalescer por fotograma.

## Open Questions

- ¿Un clic en el explorador abre el archivo (como en VS Code) o solo lo selecciona hasta el doble clic? Se asume apertura con un clic.
