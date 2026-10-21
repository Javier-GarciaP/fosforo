# 🤝 Contribuir a fosforo

Aquí se explica **cómo se trabaja** en el proyecto. Son tres ideas:

1. **Todo cambio nace de una spec** (con [OpenSpec](https://github.com/Fission-AI/OpenSpec)).
2. **Cada cambio vive en su propia rama** y entra a `dev` por *pull request*.
3. **`main` solo recibe versiones publicadas.**

```
idea ──► spec ──► rama ──► código ──► PR a dev ──► ✅ integrar ──► (cuando toca) PR a main ──► versión
```

## 🌳 Las ramas

| Rama | Qué contiene | Quién escribe en ella |
| --- | --- | --- |
| `main` | Las versiones publicadas. Cada punto lleva una etiqueta `vX.Y.Z`. | Solo PR desde `dev`. |
| `dev` | Todo el trabajo integrado: siempre lo más reciente. | Solo PR desde ramas de cambio. |
| `tipo/nombre-del-cambio` | Un cambio concreto. Vive poco y se borra al integrarse. | Quien lo implementa. |

Los tipos son `feat/` (funcionalidad), `fix/` (error), `docs/` (documentación o diseño), `refactor/` y `chore/`. Ejemplo: `feat/core-text-editing`.

> 🛡️ **`main` y `dev` están protegidas:** no se puede subir directamente (tampoco los administradores), ni forzar, ni borrarlas. Todo entra por PR, y **no hace falta aprobación**, así que quien mantiene el proyecto puede integrar sus propios PR.

## 🚶 Paso a paso

```bash
# 1. Rama nueva desde dev
git switch dev && git pull --ff-only
git switch -c feat/core-text-editing

# 2. Spec: crear o ajustar el cambio en openspec/changes/<nombre>/
openspec validate core-text-editing         # debe decir "is valid"

# 3. Implementar siguiendo tasks.md. Marca cada tarea [x] cuando cumpla su verificación.
#    Commits pequeños; cada uno deja el proyecto compilando.

# 4. Último commit de la rama: archivar el cambio
openspec archive core-text-editing

# 5. Subir y abrir el PR hacia dev (nunca hacia main)
git push -u origin feat/core-text-editing
gh pr create --base dev --fill
```

**Antes de abrir el PR** comprueba que: las pruebas pasan, todas las tareas están marcadas, `openspec validate` es válido, y si tocaste la interfaz, el diseño de `docs/design/` está actualizado.

**Integrar:** con el CI en verde y sin conflictos, **squash and merge** (un cambio = un commit en `dev`). La rama se borra sola al integrar.

### Comandos de OpenSpec

| Para... | Con Claude Code | Con la CLI |
| --- | --- | --- |
| Pensar una idea | `/opsx:explore` | — |
| Crear el cambio con sus 4 documentos | `/opsx:propose "idea"` | `openspec new change <nombre>` |
| Implementar las tareas | `/opsx:apply` | — |
| Ver estado | — | `openspec list` · `openspec status --change <nombre>` |
| Validar / archivar | `/opsx:archive` | `openspec validate <nombre>` · `openspec archive <nombre>` |

Cada cambio tiene `proposal.md` (por qué y qué), `specs/` (qué debe hacer, con escenarios), `design.md` (cómo) y `tasks.md` (pasos con su verificación).

## ✅ Reglas de oro

1. Nada de commits directos a `main` ni a `dev`.
2. Una rama = un cambio. ¿Aparece trabajo nuevo a mitad? Abre **otro** cambio y otra rama.
3. La rama sale de `dev` y vuelve a `dev`. Solo `dev` pasa a `main`.
4. No se integra con el CI en rojo ni con tareas sin verificar.
5. Las ramas viven poco: actualízalas con `git merge dev` a menudo.

## 🏷️ Versiones

fosforo usa [Versionado Semántico](https://semver.org/lang/es/): `0.MINOR.PATCH` mientras está en beta y `1.0.0` para la primera versión estable. La tabla con lo que trae cada versión está en el [README](README.md).

| Cambio | Sube | Ejemplo |
| --- | --- | --- |
| Algo nuevo que se puede usar | el número del medio | `0.2.0` → `0.3.0` |
| Solo corrección de errores | el último número | `0.3.0` → `0.3.1` |

**Para publicar una versión:**

1. Añade su entrada en [`CHANGELOG.md`](CHANGELOG.md) (por PR a `dev`).
2. Abre un PR **`dev` → `main`** titulado `release: vX.Y.Z` y intégralo con **merge commit** (sin squash).
3. Crea la etiqueta sobre `main` y la *Release* (marcada como *pre-release* mientras la versión sea `0.x`):

```bash
git switch main && git pull --ff-only
git tag -a v0.1.0 -m "v0.1.0" && git push origin v0.1.0
gh release create v0.1.0 --prerelease --title "v0.1.0" --notes-file notas.md
```

## ✍️ Commits y PR

Formato `tipo(alcance): descripción`, donde el **alcance** es el nombre del cambio de OpenSpec.

`feat(modal-editing): añade los operadores d, c e y` · `fix(buffer): corrige el cursor en emojis` · `docs(readme): aclara la instalación`

Tipos: `feat`, `fix`, `docs`, `test`, `refactor`, `ci`, `chore`. El título del PR sigue el mismo formato, porque será el mensaje del commit en `dev`. El PR se rellena con la [plantilla](.github/pull_request_template.md).

## 🎨 Diseño visual

Si el cambio altera la interfaz, actualiza `docs/design/` **en el mismo PR** (script en `src/`, regenerar `editor.op` y `exports/`). La [guía de diseño](docs/design/README.md) explica cómo.

## 🔮 Pendiente de decidir

- **Hotfix sobre `main`:** el flujo no contempla una corrección urgente que salte `dev`. Se decidirá cuando haya una versión publicada con usuarios.
- **Comprobación de CI** de que los PR a `main` vienen de `dev`: llega con `bootstrap-project`.
- **Licencia** del proyecto.
