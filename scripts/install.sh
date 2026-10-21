#!/bin/sh
# Instala fos desde la última release de GitHub (o la indicada con FOS_VERSION=0.8.0).
#   curl -fsSL https://raw.githubusercontent.com/Javier-GarciaP/fosforo/main/scripts/install.sh | sh
# Variables: FOS_VERSION (sin «v»), INSTALL_DIR (por defecto ~/.local/bin).
set -eu

REPO="Javier-GarciaP/fosforo"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

err() { echo "install.sh: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in linux | darwin) ;; *) err "sistema no soportado: $os (hay binarios para Linux y macOS)" ;; esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) err "arquitectura no soportada: $(uname -m)" ;;
esac

if have curl; then get() { curl -fsSL "$1"; }; getfile() { curl -fsSL -o "$2" "$1"; }
elif have wget; then get() { wget -qO- "$1"; }; getfile() { wget -qO "$2" "$1"; }
else err "hace falta curl o wget"; fi

version="${FOS_VERSION:-}"
if [ -z "$version" ]; then
  # la API de «latest» ignora las pre-releases (todo 0.x lo es), así que se toma la lista y se queda la primera
  version=$(get "https://api.github.com/repos/$REPO/releases" | sed -n 's/.*"tag_name": *"v\{0,1\}\([^"]*\)".*/\1/p' | head -n 1)
  [ -n "$version" ] || err "no se pudo averiguar la última versión"
fi
version="${version#v}"

name="fos_${version}_${os}_${arch}"
base="https://github.com/$REPO/releases/download/v$version"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Descargando fos $version para $os/$arch..."
getfile "$base/$name.tar.gz" "$tmp/$name.tar.gz" || err "no se pudo descargar $base/$name.tar.gz"
getfile "$base/checksums.txt" "$tmp/checksums.txt" || err "no se pudo descargar checksums.txt"

want=$(grep " $name.tar.gz\$" "$tmp/checksums.txt" | cut -d' ' -f1)
[ -n "$want" ] || err "$name.tar.gz no figura en checksums.txt"
if have sha256sum; then got=$(sha256sum "$tmp/$name.tar.gz" | cut -d' ' -f1)
elif have shasum; then got=$(shasum -a 256 "$tmp/$name.tar.gz" | cut -d' ' -f1)
else err "hace falta sha256sum o shasum para verificar la descarga"; fi
[ "$got" = "$want" ] || err "la suma SHA256 no coincide (esperada $want, obtenida $got): no se instala"

tar -xzf "$tmp/$name.tar.gz" -C "$tmp" fos
mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/fos" "$INSTALL_DIR/fos"
echo "Instalado: $INSTALL_DIR/fos ($("$INSTALL_DIR/fos" --version))"
case ":$PATH:" in *":$INSTALL_DIR:"*) ;; *) echo "Añade $INSTALL_DIR a tu PATH para poder ejecutar «fos»." ;; esac
