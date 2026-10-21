# fosforo — editor de código en la terminal

Un editor ligero con la fluidez de [Neovim](https://neovim.io) y la comodidad de VS Code, hecho en Go.

## Características

- Teclado modal con tecla líder `Space`
- Explorador de archivos y paleta de comandos
- [x] Vista previa de Markdown
- [ ] Búsqueda en proyecto

## Instalación

```sh
go install github.com/usuario/fosforo@latest
fosforo README.md
```

> Nota: los .md se abren en vista previa.

| Tecla | Acción |
| ------- | -------- |
| `Space f f` | Buscar archivo |
| `Space m p` | Alternar vista previa |

Más información en [docs/README.md](docs/README.md)
