# Design

## Context

Hay buffers, vista, resaltado y comandos. Falta renderizar Markdown a líneas con estilo dentro de un panel, y relacionar posiciones
entre el fuente y el resultado.

## Goals / Non-Goals

**Goals:**
- Renderizado fiel a la estructura de GitHub usando los colores del tema del editor.
- Renderizado fuera del bucle de eventos, con caché.

**Non-Goals:**
- Un motor Markdown propio, soporte de imágenes en terminales avanzadas (Kitty/Sixel).

## Decisions

- **Motor: Glamour** (goldmark + GFM: tablas, tareas, tachado). Se descarta escribir un renderizador propio por coste y se acepta la fidelidad aproximada.
- **Estilos propios:** dos hojas de estilo Glamour (`github-dark`, `github-light`) generadas a partir de la paleta del tema para que preview y editor
  compartan colores. La regla bajo H1/H2, el fondo de bloques de código y la barra de citas se definen aquí.
- **Resaltado de código:** Glamour usa Chroma internamente; se alinea el estilo de Chroma con el mapa de colores de `syntax-highlighting`.
- **Render asíncrono con caché:** clave = (versión del buffer, ancho, tema); resultado como mensaje de Bubble Tea; el modo dividido aplica debounce de ~150 ms.
- **Sincronización de scroll por anclas:** se extraen del AST de goldmark las líneas de los encabezados; se mapea la posición del fuente a la anterior ancla y un
  desplazamiento proporcional hasta la siguiente. Se descartó el mapeo exacto por bloque (renderizar cada bloque por separado rompe listas y tablas).
- **Conmutación de cursor:** de preview a fuente se usa el encabezado visible más cercano; de fuente a preview, el scroll por anclas.
- **Degradación:** imágenes → `[imagen: alt]` mediante un *hook* sobre el AST; HTML crudo omitido; todo lo demás sigue el estilo.
- **Panel de preview:** un tipo de panel de solo lectura con su propio keymap (`preview`), sin modo Insert.

- **Decisiones tomadas al implementar:**
  - **Motor: goldmark con renderizador propio**, en lugar de Glamour. Glamour devuelve cadenas con secuencias ANSI que habría que volver a descomponer en celdas para dibujarlas
    en la cuadrícula del editor; con goldmark (CommonMark + GFM: tablas, tareas, tachado y autoenlaces) se recorre el AST y se generan líneas de fragmentos con un **rol** (encabezado,
    enlace, cita, código...) que cada tema traduce a colores. Así el ancho es exacto (CJK y emoji), los colores salen de la paleta del editor y las anclas de scroll se calculan en el mismo recorrido.
    No hay hojas de estilo `github-dark/light`: el tema oscuro y el claro son los de la paleta Tinta.
  - Bloques de código: se tokenizan con `syntax.Tokenize` (los mismos colores que el editor); tabuladores a 4 espacios y recorte con `…` en líneas más anchas que el panel.
  - Anclas: se registra un ancla por **cada bloque de primer nivel** (no solo encabezados), marcando cuáles son encabezados. El mapeo fuente→resultado interpola entre anclas y la inversa
    (`HeadingAt`) devuelve el encabezado más cercano por encima, o el primero del documento.
  - Tablas anchas: se **recortan con `…`** (columna más ancha primero, mínimo 3 celdas). No hay scroll horizontal. Si aun así no cupieran, la línea se recorta (red de seguridad: ninguna línea se desborda).
  - El texto se ajusta a `min(ancho del panel − 3, 80)` columnas, alineado a la izquierda con 2 de margen.
  - Imágenes → `[imagen: alt]`; HTML crudo y diagramas `mermaid` se omiten (`<br>` salta de línea) y al final se añade un aviso discreto `(contenido que la terminal no puede mostrar: omitido)`.
  - Render: documentos de hasta 64 KiB se renderizan al instante; los demás, y los cambios con el modo dividido, en segundo plano. A pantalla completa y al cambiar el ancho no hay espera; con ediciones en el
    modo dividido se espera el **debounce de 150 ms** y se conserva el último renderizado (con el scroll mapeado) mientras llega el nuevo.
  - La vista previa completa tiene su propio contexto de teclas (`preview`) con `j k Ctrl+d Ctrl+u PgDn PgUp gg G {n}gg`, `e`, `:` y `Ctrl+S`. `e` solo alterna ahí: en el fuente sigue siendo «fin de palabra».
  - En el modo dividido la vista previa sigue la línea superior del fuente (`doc.OutLine(top)`).
  - Pruebas con tres documentos reales (README de presentación, `docs/README.md` y `CONTRIBUTING.md`, copiados como fixtures) en anchos de 30 a 160: sin desbordes. **Discrepancias conocidas**:
    el README de presentación usa HTML (`<picture>`, `<p align>`, insignias) que se omite con aviso; los emoji compuestos (🛡️) pueden ocupar distinto ancho según la terminal;
    no hay centrado ni imágenes.

## Risks / Trade-offs

- [Glamour envuelve y colorea con ANSI propio; el ancho puede discrepar en caracteres anchos] → pruebas golden con CJK y emoji; recorte con utilidades conscientes del ancho.
- [Tablas anchas no caben] → recorte con indicador `…` o scroll horizontal del panel; decidir tras probar READMEs reales.
- [Anclas por encabezado dan saltos en documentos sin encabezados] → recurrir a proporción global del documento.
- [Los documentos con mucho HTML (README con badges) pierden contenido] → mostrar un aviso discreto de contenido omitido.

## Open Questions

- ¿Tablas anchas con scroll horizontal o con truncado? Se resuelve en la tarea 2.4 con ejemplos reales.
