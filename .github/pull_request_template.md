## Cambio de OpenSpec

<!-- Nombre del cambio, igual que la rama. Ej.: `core-text-editing` -->
`openspec/changes/<nombre>/` (o dentro de `openspec/changes/archive/` si ya está archivado)

## Qué hace este PR

<!-- Resumen breve de lo implementado y por qué. -->

## Comprobaciones

- [ ] La rama sale de `dev` y este PR apunta a **`dev`** (no a `main`, salvo que sea un PR `release`)
- [ ] `openspec validate <nombre>` es válido
- [ ] Todas las tareas de `tasks.md` están marcadas y cada una cumple su verificación
- [ ] Las pruebas pasan (`make test`, `make vet`) y el programa compila
- [ ] Si cambió el comportamiento, la spec está actualizada en este mismo PR
- [ ] Si toca la interfaz: `docs/design/` actualizado (`src/`, `editor.op`, `exports/`)
- [ ] `docs/README.md` refleja lo que realmente existe
- [ ] Último commit: `openspec archive <nombre>`
- [ ] Título con formato `tipo(alcance): descripción` (será el commit en `dev` tras el squash)

## Cómo se probó

<!-- Pruebas automáticas añadidas y pasos manuales (terminal real, tmux, SSH...). -->

## Capturas o notas (opcional)

<!-- Si cambia la interfaz, imagen de la pantalla resultante comparada con la maqueta de docs/design/exports. -->
