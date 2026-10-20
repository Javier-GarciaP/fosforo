# Cómo contribuir

Este documento describe **cómo se trabaja** en el proyecto: el flujo de ramas con git y cómo se usan las specs
(OpenSpec) para planificar, implementar y validar cada cambio. Vale igual para una persona sola que para un equipo.

> Resumen en una línea: **cada cambio nace de una spec, vive en su propia rama, entra a `dev` por PR, y `dev` pasa a
> `main` solo cuando hay una versión estable.**

## 1. Ramas

```
main  ●─────────────────────────────────●───────────────●   ← solo versiones estables (cada punto = etiqueta vX.Y.Z)
       \                               ↑ PR "release"   ↑
dev     ●──●────●───────●──────●───────●───────●────────●   ← integra TODO; siempre la versión más reciente
           ↑    ↑       ↑      ↑
           │    │       │      └─ feat/picker-and-palette ─┐   ← una rama por cambio, creada desde dev
           │    │       └──────── feat/syntax-highlighting ┤      y cerrada (borrada) al validarse el PR
           │    └──────────────── feat/modal-editing ──────┤
           └───────────────────── docs/project-foundation ─┘
```

| Rama | Qué contiene | Quién escribe en ella |
| --- | --- | --- |
| `main` | **Versiones estables.** Cada commit de `main` es una versión publicable y lleva etiqueta `vX.Y.Z`. | Solo PRs desde `dev` ("release"). Nunca commits directos. |
| `dev` | **Integración.** La versión más reciente del proyecto; puede cambiar a diario pero debe compilar y pasar CI. | Solo PRs desde ramas de cambio. Nunca commits directos. |
| `<tipo>/<cambio>` | **Un cambio concreto**, con el mismo nombre que su cambio de OpenSpec. | La persona que lo implementa. Vive poco y se borra al integrarse. |

### Nombre de las ramas de cambio

`<tipo>/<nombre-del-cambio-de-openspec>`, todo en minúsculas y con guiones:

| Tipo | Cuándo | Ejemplo |
| --- | --- | --- |
| `feat/` | Funcionalidad nueva | `feat/core-text-editing` |
| `fix/` | Corregir un error | `fix/cursor-wide-chars` |
| `docs/` | Solo documentación o diseño | `docs/project-foundation` |
| `refactor/` | Cambiar estructura sin cambiar comportamiento | `refactor/keymap-trie` |
| `chore/` | Herramientas, CI, dependencias | `chore/ci-go-cache` |

## 2. Todo cambio pasa por una spec

**No hay cambios sin spec**, ni siquiera las integraciones o la documentación grande: cualquier cambio o integración se
describe primero como un *cambio de OpenSpec* en `openspec/changes/<nombre>/`. Así queda escrito **qué** se va a hacer
y **cómo se comprueba**, antes de tocar código.

Cada cambio tiene cuatro artefactos:

| Archivo | Responde a |
| --- | --- |
| `proposal.md` | ¿Por qué? ¿Qué cambia? ¿Qué queda fuera? |
| `specs/<capacidad>/spec.md` | ¿Qué debe hacer el sistema? Requisitos (`SHALL`/`MUST`) con escenarios `WHEN`/`THEN` |
| `design.md` | ¿Cómo se hará? Decisiones, alternativas descartadas, riesgos |
| `tasks.md` | ¿En qué pasos? Casillas `- [ ]`, cada una con su forma de verificarla |

### Comandos de OpenSpec que usarás

| Acción | Con Claude Code | Con la CLI |
| --- | --- | --- |
| Pensar una idea antes de proponerla | `/opsx:explore` | — |
| Crear el cambio con sus 4 artefactos | `/opsx:propose "idea"` | `openspec new change <nombre>` |
| Revisar y corregir los artefactos | `/opsx:update` | editar los `.md` |
| Implementar las tareas | `/opsx:apply` | — |
| Ver estado y progreso | — | `openspec list`, `openspec status --change <nombre>` |
| Validar | — | `openspec validate <nombre>` |
| Archivar al terminar | `/opsx:archive` | `openspec archive <nombre>` |

`openspec archive` mueve el cambio a `openspec/changes/archive/` y **fusiona sus specs en `openspec/specs/`**, que es el
estado actual del comportamiento del sistema.

## 3. Flujo completo de un cambio

```
 dev ──► rama ──► spec ──► implementar ──► validar ──► PR a dev ──► revisión + CI ──► merge ──► borrar rama
          (1)      (2)        (3)           (4)          (5)             (6)           (7)        (8)
```

### Paso a paso

**1. Crear la rama desde `dev` actualizado**

```bash
git switch dev && git pull --ff-only
git switch -c feat/core-text-editing
```

**2. Escribir (o completar) la spec del cambio** en `openspec/changes/<nombre>/`.
Si el cambio ya existe en el repositorio (como los de la hoja de ruta), léelo y ajústalo si hace falta.

```bash
openspec validate core-text-editing      # debe decir "is valid"
git add openspec && git commit -m "docs(core-text-editing): ajusta la spec"
```

**3. Implementar siguiendo `tasks.md`**, de arriba abajo. Después de cada tarea:
marca su casilla (`- [x]`), comprueba lo que dice su verificación y haz un commit pequeño.

**4. Validar antes de abrir el PR**

- Las pruebas pasan (`make test`, y `make vet`) y el programa compila.
- Todas las tareas del cambio están marcadas y verificadas.
- `openspec validate <nombre>` sigue siendo válido (si cambió algo del comportamiento, la spec se actualizó en el mismo PR).
- Si el cambio toca la interfaz, el diseño de `docs/design/` está actualizado (ver la [guía de diseño](docs/design/README.md)).
- `docs/README.md` refleja lo que realmente existe.
- Último commit de la rama: **archivar el cambio** (`openspec archive <nombre>`), para que `dev` siempre tenga las specs al día.

**5. Abrir el PR hacia `dev`** (nunca hacia `main`)

```bash
git push -u origin feat/core-text-editing
gh pr create --base dev --title "feat(core-text-editing): buffer, vista y escritura básica" --fill
```

El PR usa la plantilla de `.github/pull_request_template.md`.

**6. Revisión y CI.** El PR se aprueba cuando el CI está en verde y la persona revisora comprueba que lo implementado
cumple la spec (escenarios `WHEN/THEN`). Los comentarios se resuelven con commits en la misma rama.

**7. Integrar en `dev`** con **"Squash and merge"**: un cambio = un commit en `dev`, con el título del PR como mensaje.

**8. Cerrar la rama de implementación.** Una vez integrada, la rama ya no se usa:

```bash
git push origin --delete feat/core-text-editing    # o activar "borrar rama automáticamente al hacer merge" en GitHub
git switch dev && git pull --ff-only
git branch -d feat/core-text-editing
```

### Reglas de oro

1. Nada de commits directos a `main` ni a `dev`.
2. Una rama = un cambio de OpenSpec. Si descubres trabajo nuevo a mitad de camino, abre **otro** cambio y otra rama.
3. La rama sale de `dev` y vuelve a `dev`. Solo `dev` se integra en `main`.
4. Un PR no se integra con el CI en rojo ni con tareas sin verificar.
5. Las ramas viven poco: cuanto más tiempo pasan abiertas, más duelen los conflictos. Actualízalas con
   `git merge dev` (o `git rebase dev` si nadie más la usa) con frecuencia.

## 4. De `dev` a `main`: versiones estables

`main` solo recibe `dev` cuando existe una **versión estable**. No hay un calendario fijo: se publica cuando se cumple
la lista siguiente.

**Criterios de una versión estable**

- [ ] Todos los cambios previstos para la versión están **archivados** (`openspec list` no muestra cambios pendientes de ella).
- [ ] El CI de `dev` está en verde.
- [ ] Prueba manual de extremo a extremo hecha en una terminal real y dentro de `tmux`: abrir, editar, guardar, ver un `.md`.
- [ ] `docs/README.md` y las instrucciones de instalación están al día.
- [ ] Se decidió el número de versión (ver abajo) y se actualizaron las notas de cambios.

**Cómo se hace**

```bash
git switch dev && git pull --ff-only
git switch -c release/v0.1.0 dev          # opcional: una rama para ajustes finales (versión, notas)
git push -u origin release/v0.1.0
gh pr create --base main --head release/v0.1.0 --title "release: v0.1.0" --body "Notas de la versión..."
```

1. El PR `dev → main` se integra con **merge commit** (no squash), para conservar la historia de `dev`.
2. Después del merge se crea la etiqueta **sobre `main`**: `git tag -a v0.1.0 -m "v0.1.0"` y `git push origin v0.1.0`.
   La etiqueta dispara el workflow de release, que **falla si el commit no está en `main`**.
3. Se borra la rama de release y `dev` se actualiza con `main` si hubo ajustes finales.

**Numeración (versionado semántico `MAJOR.MINOR.PATCH`)**

- Mientras no haya una v1 estable: `0.MINOR.PATCH`. Funcionalidad nueva sube `MINOR`; correcciones suben `PATCH`.
- La primera versión completa de la hoja de ruta (cambios 1 a 9) es **`1.0.0`**.

## 5. Mensajes de commit y PRs

Formato (Conventional Commits, en español): `tipo(alcance): descripción en imperativo`.

| Tipo | Uso |
| --- | --- |
| `feat` | Funcionalidad nueva |
| `fix` | Corrección de un error |
| `docs` | Documentación, specs o diseño |
| `test` | Pruebas |
| `refactor` | Reestructurar sin cambiar el comportamiento |
| `ci` / `chore` | CI, herramientas, dependencias |

El **alcance** es el nombre del cambio de OpenSpec: `feat(modal-editing): añade operadores d, c e y`.
Los commits son pequeños y cada uno deja el proyecto compilando.

El título del PR sigue el mismo formato, porque es el mensaje del commit que quedará en `dev` tras el squash.

## 6. Diseño visual

Cualquier cambio visible en la interfaz actualiza también el diseño en `docs/design/`, **en el mismo PR**:

1. Edita el script de la pantalla en `docs/design/src/`.
2. Regenera `editor.op` y `exports/` como indica la [guía de diseño](docs/design/README.md).
3. Incluye `src/`, `editor.op` y `exports/` en el PR.

## 7. Protección de ramas (recomendado, a configurar en GitHub)

Estas reglas hacen que el flujo se cumpla solo. Se configuran en *Settings → Branches* del repositorio remoto:

| Regla | `main` | `dev` |
| --- | --- | --- |
| Exigir pull request para integrar | Sí | Sí |
| Exigir que el CI pase | Sí | Sí |
| Prohibir push directo (incluye administradores) | Sí | Sí |
| Exigir ramas actualizadas antes de integrar | Sí | Opcional |
| Borrar la rama al integrar el PR | — | Sí (en *General*) |

## 8. Preguntas abiertas del flujo

Decisiones que no se tomaron todavía y que conviene fijar cuando aparezcan:

- **Correcciones urgentes sobre `main`:** el flujo actual no contempla un *hotfix* que salte `dev`. Mientras no exista una
  versión publicada no hace falta; después habrá que decidir si se permite `hotfix/*` desde `main` con vuelta a `dev`.
- **Número de personas revisoras:** el flujo exige PR y CI en verde; cuántas aprobaciones se necesitan depende de quiénes sean
  las personas que colaboren.
- **Nombre del proyecto y del repositorio:** `ted` es provisional.
