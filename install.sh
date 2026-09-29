#!/bin/sh
# Instalador del pizarrón ligero — POSIX sh, sin dependencias salvo curl o go.
#   sh install.sh                          # usa binario precompilado si hay release, si no compila
#   sh install.sh --from-source            # compila desde este checkout (requiere go >= 1.21)
#   sh install.sh --version v0.2.0 --prefix ~/.local/bin
# Tras instalar, el arranque es manual (decisión v1): ver README.md.
set -eu

SCRIPT_DIR_INIT="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR_INIT/VERSION" ]; then VERSION="$(tr -d '[:space:]' < "$SCRIPT_DIR_INIT/VERSION")";
else VERSION="v0.1.0"; fi
PREFIX="$HOME/.local/bin"
FROM_SOURCE="no"
# Cambiar al publicar el primer release (ver Makefile: make build-all).
RELEASE_BASE="https://github.com/TU_USUARIO/pizarron/releases/download"

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2;;
    --prefix) PREFIX="$2"; shift 2;;
    --from-source) FROM_SOURCE="yes"; shift;;
    -h|--help) sed -n '2,7p' "$0"; exit 0;;
    *) echo "flag desconocida: $1 (ver --help)" >&2; exit 1;;
  esac
done

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64";;
  arm64|aarch64) ARCH="arm64";;
  *) echo "arquitectura no soportada: $ARCH" >&2; exit 1;;
esac
case "$OS" in
  darwin|linux) BIN="pizarron-${OS}-${ARCH}";;
  *) echo "SO no soportado: $OS (compila con --from-source)" >&2; exit 1;;
esac

mkdir -p "$PREFIX"
DEST="$PREFIX/pizarron"

install_from_release() {
  command -v curl >/dev/null 2>&1 || return 1
  echo "descargando $RELEASE_BASE/$VERSION/$BIN ..."
  curl -fsSL -o "$DEST" "$RELEASE_BASE/$VERSION/$BIN" || return 1
  chmod +x "$DEST"
}

install_from_source() {
  command -v go >/dev/null 2>&1 || {
    echo "go no encontrado (>= 1.21 requerido). Instala Go o usa un binario de releases." >&2
    return 1
  }
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  if [ ! -f "$SCRIPT_DIR/go.mod" ]; then
    echo "no hay checkout fuente en $SCRIPT_DIR (falta go.mod)" >&2
    return 1
  fi
  echo "compilando desde fuente ($SCRIPT_DIR) ..."
  (cd "$SCRIPT_DIR" && CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=$VERSION" -o "$DEST" ./cmd/pizarron)
}

if [ "$FROM_SOURCE" = "yes" ]; then
  install_from_source || exit 1
else
  install_from_release || {
    echo "sin release para $BIN; intentando compilar desde fuente ..."
    install_from_source || exit 1
  }
fi

INSTALLED="$("$DEST" -version)"
echo "instalado: $DEST ($INSTALLED)"
echo "verifica: PIZARRON_ADDR=127.0.0.1:8000 $DEST & curl -s http://127.0.0.1:8000/healthz"
echo "viewer personal: http://127.0.0.1:8000/  |  agentes: ver agents/AGENTS.snippet.md"
case ":$PATH:" in
  *":$PREFIX:"*) ;;
  *) echo "nota: $PREFIX no está en tu PATH; añade: export PATH=\"$PREFIX:\$PATH\"" ;;
esac
