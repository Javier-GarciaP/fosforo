# Tasks

## 1. Dibujo (paquete `ui`)

- [x] 1.1 Crear `internal/ui/start_view.go` con el logo "FOSFORO" (letras de bloque), `StartEntries` y `drawStart`, centrando en el área del editor y con los tres niveles de tamaño. Prueba primero (explicada): con un área grande el resultado contiene el logo y las ocho acciones; con un área baja solo la lista; con un área mínima nada.
- [x] 1.2 Añadir el campo `Start bool` a `EditorView` y usarlo en `RenderEditor` en lugar de los `~`/líneas cuando la pantalla cabe. Prueba (explicada): con `Start` el texto "FOSFORO" aparece y con `Start` falso no; el logo usa el color `Brand` del tema oscuro y del claro.
- [x] 1.3 Hacer que `CursorCell` no muestre cursor mientras la pantalla de inicio está dibujada. Prueba (explicada): devuelve `ok=false`.

## 2. Decisión (paquete `editor`)

- [x] 2.1 En `internal/editor/model.go`, calcular `Start` al construir la vista: buffer sin nombre, vacío, sin modificar, sin vista previa Markdown y no en modo Insert. Prueba de modelo (explicada, con teatest): abrir carpeta muestra la pantalla; abrir un archivo la oculta; pulsar `i` la oculta; cerrar el último buffer la vuelve a mostrar.
- [x] 2.2 Prueba (explicada) que recorre `ui.StartEntries` y comprueba que cada atajo está asignado en el mapa de teclas al comando esperado (`file.find`, `file.new`, `file.recent`, `search.project`, `explorer.toggle`, `config.open`, `help.keys`, `app.quit`).
- [x] 2.3 Prueba (explicada): los clics y las teclas normales (`:`) se comportan igual con la pantalla visible (no hay zona clicable nueva).

## 3. Documentación y cierre

- [x] 3.1 Actualizar `docs/README.md` (tabla de funciones) y `CHANGELOG.md` en inglés sencillo.
- [x] 3.2 Ejecutar `go vet ./...` y `go test ./...`, y revisar visualmente con `fos .` en tema oscuro y claro y en una terminal pequeña.
