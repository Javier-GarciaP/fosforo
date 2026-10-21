# Tasks

## 1. Buffer

- [x] 1.1 Implementar la carga de archivos preservando CRLF/LF y salto final; verificar con pruebas de tabla de ida y vuelta byte a byte
- [x] 1.2 Detectar solo lectura (NUL, UTF-8 inválido, >10 MiB); verificar con archivos de prueba
- [x] 1.3 Implementar `Apply(Edit)` e inserción/borrado por grafemas; verificar con pruebas Unicode (CJK, emoji compuesto, combinantes)
- [x] 1.4 Implementar la marca de modificado; verificar que cambia con una edición y se limpia al guardar
- [x] 1.5 Implementar el guardado atómico con permisos; verificar el guardado correcto y el fallo en un directorio sin escritura

## 2. Cursor y columnas

- [x] 2.1 Implementar movimiento por grafemas y columna visual con tabs de 4 y caracteres anchos; verificar con pruebas de tabla
- [x] 2.2 Implementar la columna deseada en movimiento vertical; verificar con el escenario de líneas cortas y largas

## 3. Vista

- [x] 3.1 Implementar el render del área de texto con gutter y `~`; verificar con pruebas golden contra la maqueta 01
- [x] 3.2 Implementar la statusline y la línea de comandos; verificar con pruebas golden (`12:35`, `[+]`, solo lectura)
- [x] 3.3 Implementar el scroll vertical con margen 3 y horizontal; verificar con pruebas del modelo que mueven el cursor
- [x] 3.4 Manejar el redimensionado; verificar con `teatest` enviando mensajes de tamaño

## 4. Edición e integración

- [x] 4.1 Implementar la escritura (insertar, Enter con sangría, Backspace/Delete, unión de líneas); verificar con `teatest`
- [x] 4.2 Implementar flechas, Home, End, PageUp y PageDown con límites; verificar con `teatest`
- [x] 4.3 Implementar `Ctrl+S` y `Ctrl+Q` con diálogo de cambios sin guardar; verificar con `teatest` y comparando con la maqueta 10
- [x] 4.4 Abrir la ruta pasada como argumento; verificar abriendo un archivo real con `fos <ruta>` y guardando una edición
- [x] 4.5 Actualizar `docs/README.md` con el estado real de funciones; verificar que la lista coincide con lo implementado
