# Design

## Context

El comando es lo que la persona escribe a diario; la marca es lo que ve en el README y la página del proyecto. Pueden diferir
(como `rg` y *ripgrep*, o `nvim` y *Neovim*).

## Decisions

- **Comando `fos`, marca *fosforo*.** Corto, de tres teclas, y empieza igual que la marca. Directorio del binario `cmd/fos`.
- **Sin alias `fosforo`** en la primera versión: un solo nombre de comando evita ambigüedad en la documentación.
- **Verificación de colisiones (2026-10-20):**
  - No existe `fos` en Homebrew (fórmula ni cask), Scoop, AUR ni como paquete de Debian; tampoco en esta máquina (Fedora).
  - Existen paquetes con ese nombre pero **sin relación con un editor ni un comando común**: `fos` en RubyGems (herramienta de archivado
    de carpetas, sin mantenimiento desde 2014), `fos` en PyPI (biblioteca de aprendizaje profundo, sin comando) y `fos` en npm
    (biblioteca JavaScript, última versión `0.0.10-a`, sin binario).
  - Conclusión: sin coincidencias fuertes. El riesgo residual es que alguien tenga instalada la gema `fos` de Ruby y deba elegir cuál queda
    primero en el `PATH`; se documenta.
- **Licencia MIT.** Sencilla, permisiva y compatible con las dependencias previstas (Bubble Tea, Lip Gloss, Glamour, Chroma: todas MIT).
  Alternativa descartada: Apache-2.0, por añadir complejidad sin necesidad (no hay empresa ni patentes de por medio).
- **Titular del copyright:** el autor del repositorio, con año de creación 2026.

## Risks / Trade-offs

- [`fos` choca con la gema de Ruby en el `PATH`] → documentarlo en la instalación; la gema está abandonada.
- [Cambiar de nombre de comando más adelante es costoso] → el nombre vive en una sola constante en el código y en una sola mención por documento.
