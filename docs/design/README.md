# Diseño visual con OpenPencil

Las maquetas del editor están hechas con **[OpenPencil](https://github.com/ZSeven-W/openpencil)**, una herramienta de
diseño vectorial de código abierto (MIT) cuyos archivos `.op` son JSON legible y versionable con git. Esta guía explica
cómo **ver**, **modificar** y **regenerar** el diseño.

## Qué hay en esta carpeta

```
docs/design/
├── editor.op          ← el diseño completo (12 pantallas + el tablero de marca). Se abre con OpenPencil.
├── exports/           ← una imagen PNG por pantalla, para verlas sin abrir nada
│   ├── 00-guia-de-estilo.png
│   ├── 01-layout-principal-normal.png
│   └── ... (hasta 12-marca-candidatos.png)
├── src/               ← scripts que GENERAN el diseño (fuente de verdad de las maquetas)
│   ├── palette.js     ← GENERADO: paleta Tinta (oscuro y claro)
│   ├── icons.js       ← GENERADO: iconos de archivo como trazados vectoriales
│   ├── lib.js         ← piezas compartidas: líneas, explorador, statusline, popups, tabla...
│   └── 00-… a 12-….js ← una pantalla por archivo
└── tools/             ← generadores de palette.js e icons.js (Python)
    ├── make-palette.py
    └── make-icons.py
```

> **Regla:** `src/` es la fuente de verdad. `editor.op` y `exports/` son resultados generados. Si cambias una
> pantalla a mano en la aplicación, traslada ese cambio al script correspondiente, o la próxima regeneración lo borrará.

## Qué tienes instalado

| Pieza | Ruta | Para qué |
| --- | --- | --- |
| `op` (CLI, v0.8.4) | `~/.local/bin/op` | Controlar OpenPencil desde la terminal (y que lo controle un agente) |
| `openpencil-desktop` | `~/.local/bin/openpencil-desktop` | La aplicación de escritorio que usa `op start` |
| `OpenPencil.AppImage` | `~/.local/bin/OpenPencil.AppImage` | La aplicación de escritorio descargada aparte |
| Lanzador de GNOME | `~/.local/share/applications/openpencil-terminal-ide.desktop` | Abre este diseño desde el menú de aplicaciones |

## Ver el diseño

Tres formas, de la más rápida a la más completa:

1. **Solo mirar:** abre las imágenes de [`exports/`](exports/). No necesitas nada instalado.
2. **Desde el menú de aplicaciones:** busca **"OpenPencil - terminal-IDE"**. Abre `editor.op` directamente.
3. **Desde la terminal**, estando en la raíz del repositorio:

   ```bash
   op start --file docs/design/editor.op
   ```

   Se abre la ventana de OpenPencil con todas las pantallas en el lienzo, en una cuadrícula de 3 columnas.

   También puedes abrir `OpenPencil.AppImage`; la documentación indica que los archivos `.op` se asocian a la aplicación
   (doble clic para abrir) y que el menú de archivo incluye "Open Recent" y "Save As".

Para cerrar el servidor de control que arranca con la aplicación: `op stop`.

## Cómo se usa la aplicación

> La documentación oficial describe solo una parte de la interfaz. Lo que sigue es **lo que está documentado**; lo demás
> se marca como no documentado en vez de suponerlo. En caso de duda, mira los menús de la propia aplicación.

### Herramientas y atajos documentados

| Tecla | Acción | Tecla | Acción |
| --- | --- | --- | --- |
| `V` | Seleccionar | `Cmd+S` | Guardar |
| `R` | Rectángulo | `Cmd+Shift+S` | Guardar como |
| `O` | Elipse | `Cmd+Z` | Deshacer |
| `L` | Línea | `Cmd+Shift+Z` | Rehacer |
| `T` | Texto | `Cmd+C / X / V / D` | Copiar / cortar / pegar / duplicar |
| `F` | Marco (frame) | `Cmd+G` | Agrupar |
| `P` | Pluma | `Cmd+Shift+G` | Desagrupar |
| `H` | Mano (desplazar el lienzo) | `Cmd+Shift+P` | Exportar (PNG, JPG, WEBP, PDF) |
| `Del` | Borrar | `Cmd+Shift+C` | Panel de código |
| `[` `]` | Cambiar el orden de capas | `Cmd+Shift+V` | Panel de variables |
| Flechas | Mover 1 px | `Cmd+J` | Chat de IA |

La tabla oficial solo habla de `Cmd`. **En Linux suele corresponder a `Ctrl`**, pero no está documentado: compruébalo con
un atajo inofensivo (por ejemplo, `Ctrl+Z`).

### Lienzo

- El lienzo es infinito, con zoom, guías de alineación y ajuste (snapping).
- Para moverte por él, la herramienta **Mano (`H`)** está documentada. Los gestos de ratón/trackpad y los atajos de zoom
  **no están documentados**; prueba rueda del ratón / pellizco.
- Cada pantalla es un **frame de nivel superior** llamado `00 Guia de estilo`, `01 Layout principal (Normal)`, etc.
  Selecciona uno para ver sus capas y propiedades.

### Qué no está documentado

El README oficial no explica los paneles de capas y propiedades, la selección múltiple, el zoom ni dónde está el menú
"Open". Es una herramienta joven (versión 0.8): conviene explorarla un rato en vez de fiarse de una guía.

## Cómo está construido el diseño

- **Cuadrícula de terminal:** cada pantalla mide **120 × 36 celdas** de **9 × 18 px** (1080 × 648 px). La fuente es
  *JetBrains Mono* a **15 px**, que da exactamente 9 px por carácter. Así una columna del diseño equivale a una columna
  real de terminal.
- **Paleta «Tinta»** (nombre provisional): propia, en oscuro y claro, generada en `src/palette.js` y mostrada en la pantalla `00`.
  Ver [Paleta e iconos](#paleta-e-iconos).
- **Colores por modo:** NORMAL azul, INSERT verde, VISUAL malva, COMMAND melocotón, PREVIEW verde azulado.
- **Disposición de las pantallas en el lienzo:**

  | | Columna 1 | Columna 2 | Columna 3 |
  | --- | --- | --- | --- |
  | Fila 1 | 00 Guía de estilo | 01 Layout principal | 02 Modo Insert |
  | Fila 2 | 03 Modo Visual | 04 Menú líder | 05 Picker de archivos |
  | Fila 3 | 06 Paleta de comandos | 07 Markdown vista previa | 08 Markdown dividido |
  | Fila 4 | 09 Tema claro | 10 Diálogo cambios sin guardar | 11 Iconos y explorador |
  | Fila 5 | 12 Marca: candidatos (provisional; se sustituye al elegir nombre) | | |

## Paleta e iconos

### Paleta «Tinta»

Inspirada en la sensación de tu referencia (fondo azul-tinta, panel más oscuro que el editor, acentos suaves), pero **propia**:
no copia los valores de ninguna paleta existente. Se **deriva en OKLCH** con `tools/make-palette.py`:

- Todos los acentos comparten **la misma luminosidad y croma** y solo cambian de tono, así que ninguno grita más que otro.
- Las superficies (fondo, panel, selección, bordes) comparten el tono azul-violeta de la marca con croma muy bajo.
- El script **comprueba el contraste** (WCAG): texto ≥ 7:1, acentos ≥ 4,5:1 y comentarios ≥ 3:1; si algo falla, no genera nada.

Para cambiar la identidad ajusta `BRAND_HUE` y `ACCENT_HUES` en el script y ejecútalo:

```bash
python3 docs/design/tools/make-palette.py        # reescribe src/palette.js e imprime los contrastes
```

### Iconos de archivo

El explorador **no usa flechas**: la estructura se lee por la sangría (2 columnas por nivel) y el estado de cada carpeta por su
icono (cerrada / abierta). Cada archivo lleva el icono de su tipo, resuelto en este orden: **nombre exacto** (`Makefile`,
`go.mod`, `.gitignore`…) → **extensión** → icono por defecto. La pantalla [`11`](exports/11-iconos-y-explorador.png) muestra
los 38 iconos y qué nombres o extensiones activan cada uno.

Los iconos son **glifos de Nerd Fonts** (familias Seti y Material Design): en el editor real son caracteres de texto, así que
**necesitan una Nerd Font** en la terminal. Aquí se convierten a trazados vectoriales con `tools/make-icons.py` para dibujarlos
sin instalar fuentes:

```bash
# Descargas (en un directorio temporal, no en el repositorio):
#   https://github.com/ryanoasis/nerd-fonts/releases/latest/download/NerdFontsSymbolsOnly.tar.xz   (extraer)
#   https://raw.githubusercontent.com/ryanoasis/nerd-fonts/master/glyphnames.json
pip install fonttools
python3 docs/design/tools/make-icons.py <ruta>/SymbolsNerdFont-Regular.ttf <ruta>/glyphnames.json   # reescribe src/icons.js
```

La tabla de qué icono usa cada tipo vive en `tools/make-icons.py` (diccionario `SET`), que es la referencia para implementarla.
Sin Nerd Font se usa el modo `icons = "none"`: las carpetas llevan `/` al final, las abiertas en azul con sus hijos sangrados y las
cerradas atenuadas (ver pantalla `11`).

## Modificar el diseño

El flujo recomendado, que sirve tanto a una persona como a un agente (Claude) conectado por MCP:

1. **Edita el script** de la pantalla en `src/` (o `src/lib.js` si cambia algo compartido).
2. **Arranca el servidor sin ventana** sobre un archivo limpio y aplica los scripts.
3. **Guarda y exporta**, y revisa las imágenes.
4. **Commit** de `src/`, `editor.op` y `exports/` en el mismo PR (ver [`CONTRIBUTING.md`](../../CONTRIBUTING.md)).

```bash
# Desde la raíz del repositorio
op stop                                   # por si hay una instancia abierta (comparten el puerto 3100)
rm -f docs/design/editor.op
op start --headless --file docs/design/editor.op &
sleep 4

# Aplicar cada pantalla: palette.js + icons.js + lib.js + pantalla, concatenados en un archivo temporal .js
mkdir -p /tmp/op-build
for f in docs/design/src/[0-9]*.js; do
  n=$(basename "$f")
  cat docs/design/src/palette.js docs/design/src/icons.js docs/design/src/lib.js "$f" > "/tmp/op-build/$n"
  op design "@/tmp/op-build/$n" > /dev/null
done

op save_document filePath="$PWD/docs/design/editor.op"

# PNG de cada pantalla, copiados a exports/ con nombres limpios (01-layout-principal-normal.png)
rm -rf /tmp/op-png && mkdir -p /tmp/op-png docs/design/exports
op export-frames --output-dir /tmp/op-png --format png --scale 1
for f in /tmp/op-png/*.png; do
  b=$(basename "$f" .png); n=${b#*-}
  n=$(echo "$n" | iconv -f utf8 -t ascii//TRANSLIT | tr 'A-Z' 'a-z' | sed 's/[^a-z0-9]\+/-/g; s/^-//; s/-$//')
  cp "$f" "docs/design/exports/$n.png"
done
op stop
```

Para **ver el resultado en la aplicación**, termina con `op start --file docs/design/editor.op`.

### Que un agente controle OpenPencil

`op` es un cliente de un servidor MCP local: `http://127.0.0.1:3100/mcp`. Una sesión de Claude Code puede llamar a las
mismas herramientas (`op tools` lista las 166 disponibles: `batch_design`, `insert_node`, `get_screenshot`,
`export_nodes`...). El comando `op install` instala la *skill* de OpenPencil en los agentes detectados.

> ⚠️ **Seguridad:** ese endpoint **no tiene autenticación**. No uses `op start --web --host 0.0.0.0` ni lo expongas
> en una red: cualquiera que lo alcance podría modificar el diseño.

## Trampas conocidas (aprendidas al construir estas maquetas)

- **El primer hijo queda encima.** En un frame, el orden de dibujo es inverso al de creación: lo primero que se crea se
  pinta *encima*. Los popups y el fondo atenuado deben crearse **antes** que la aplicación (ver `src/05-picker-archivos.js`).
- **Los glifos de caja (`┌ │ ┬ …`) se renderizan con otra altura de línea** y desalinean las filas. Las tablas del diseño
  se dibujan con bordes (`tableGrid` en `lib.js`); en el editor real sí se usarán caracteres de caja.
- **No uses el carácter backtick dentro de un script**: el preprocesador del DSL lo confunde con una plantilla. Constrúyelo
  con `String.fromCharCode(96)` (ver `src/08-markdown-split.js`).
- **Un texto vacío ocupa ancho.** Filtra los fragmentos `''` (ver `hl()` en `lib.js`).
- **Solo una instancia a la vez** en el puerto 3100: si `op start` falla, ejecuta `op status` y `op stop`.
- **`op design -` (stdin) no admite `--script`.** Usa un archivo `.js` con `@ruta`: con esa extensión se activa el modo
  script automáticamente.

## Solución de problemas

| Síntoma | Qué hacer |
| --- | --- |
| `op start` dice que no encuentra el binario de escritorio | Comprueba que existe `~/.local/bin/openpencil-desktop` o define `OPENPENCIL_DESKTOP_BIN` |
| No se abre la ventana | `op status`; si hay un servidor sin ventana (`--headless`), `op stop` y vuelve a `op start --file …` |
| El lanzador no aparece en el menú de GNOME | Ejecuta `update-desktop-database ~/.local/share/applications`; el archivo `.desktop` debe empezar con `[Desktop Entry]` (no `[Entry]`) |
| `op design` devuelve `errors` y no aplica nada | El lote es transaccional: corrige las líneas indicadas y reenvía **todo** el lote |
| El AppImage no arranca | Asegúrate de que tiene permiso de ejecución (`chmod +x`); la documentación oficial no menciona requisitos de FUSE |

## Referencias

- Repositorio: <https://github.com/ZSeven-W/openpencil>
- Referencia del CLI: `crates/op-cli/src/usage.txt` en el repositorio, o `op help`
- Herramientas MCP disponibles: `op tools`
