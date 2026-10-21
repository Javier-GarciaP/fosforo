# Spec Delta

## Purpose
Retira los atajos transitorios de `core-text-editing`, que el modo Normal y la línea de comandos sustituyen.

## REMOVED Requirements

### Requirement: Transitional save and quit
**Reason:** `Ctrl+S` y `Ctrl+Q` eran provisionales hasta que existiera la edición modal. Ahora salir es `:q`, `:q!`, `:wq` o `:x`, y `Ctrl+S` pasa a ser `file.save` en todos los modos (ver `modal-input`).
**Migration:** Usar `:q` (con el mismo diálogo de cambios sin guardar), `:wq` o `:q!`. `Ctrl+S` conserva su comportamiento.
