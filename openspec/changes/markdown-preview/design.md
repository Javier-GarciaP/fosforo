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

## Risks / Trade-offs

- [Glamour envuelve y colorea con ANSI propio; el ancho puede discrepar en caracteres anchos] → pruebas golden con CJK y emoji; recorte con utilidades conscientes del ancho.
- [Tablas anchas no caben] → recorte con indicador `…` o scroll horizontal del panel; decidir tras probar READMEs reales.
- [Anclas por encabezado dan saltos en documentos sin encabezados] → recurrir a proporción global del documento.
- [Los documentos con mucho HTML (README con badges) pierden contenido] → mostrar un aviso discreto de contenido omitido.

## Open Questions

- ¿Tablas anchas con scroll horizontal o con truncado? Se resuelve en la tarea 2.4 con ejemplos reales.
