#!/bin/sh
# bootstrap.sh — ciclo de vida zero-touch del pizarrón (no interactivo, idempotente).
# Asegura Go (descarga oficial pinneada + sha256 solo si falta), compila, elige
# puerto libre, arranca, verifica con el contrato e imprime PIZARRON_URL.
# Uso: sh agent/bootstrap.sh
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="$(tr -d '[:space:]' < "$ROOT/VERSION")"
GO_MIN_MINOR="21"
GO_PIN="go1.27.1"
GO_BASE="https://go.dev/dl"
BIN_DIR="$ROOT/.bin"
TOOLS_DIR="$ROOT/.tools"
RUN_DIR="$ROOT/.run"
BIN="$BIN_DIR/pizarron"
PIDFILE="$RUN_DIR/pizarron.pid"
OUTLOG="$RUN_DIR/pizarron.out"
APPLOG="$RUN_DIR/pizarron.log"

sha_of() { # $1=os-arch -> sha256 pinneado (go.dev/dl/?mode=json, 2026-09-29)
  case "$1" in
    darwin-arm64) echo "ee215d57e0ec269c60cc9ceca68e6bda321ba9ee5afe24f4b0988703c2d87d12";;
    darwin-amd64) echo "8f8f52c6649542cf027bbc9b9c68d1ec042f9f34808a40413f0b8b3f66f3caa4";;
    linux-arm64)  echo "3450b45a3f9ee8568792736a5c5e70a1f2e9b36c35a8f74958c03e51d7d92bec";;
    linux-amd64)  echo "63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445";;
    *) echo "plataforma sin Go pinneado: $1 (usa dist/ precompilado o instala Go >= 1.21)" >&2; exit 1;;
  esac
}

go_minor() { # imprime minor de `go version` o "none"
  if ! command -v go >/dev/null 2>&1; then echo "none"; return; fi
  ver="$(go version 2>/dev/null | awk '{print $3}')" || { echo "none"; return; }
  echo "$ver" | sed -E 's/^go1\.([0-9]+).*/\1/;t;d' | grep -E '^[0-9]+$' || echo "none"
}

ensure_go() {
  minor="$(go_minor)"
  if [ "$minor" != "none" ] && [ "$minor" -ge "$GO_MIN_MINOR" ]; then return 0; fi
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  case "$ARCH" in x86_64|amd64) ARCH="amd64";; arm64|aarch64) ARCH="arm64";; *) echo "arch no soportada: $ARCH" >&2; exit 1;; esac
  case "$OS" in darwin|linux) ;; *) echo "SO no soportado para auto-Go: $OS (instala Go >= 1.21 o usa dist/)" >&2; exit 1;; esac
  command -v curl >/dev/null 2>&1 || { echo "sin go y sin curl: instala Go >= 1.21 a mano" >&2; exit 1; }
  file="$GO_PIN.$OS-$ARCH.tar.gz"
  echo "go ausente/viejo -> descargando oficial $file ..."
  mkdir -p "$TOOLS_DIR"
  curl -fsSL -o "$TOOLS_DIR/$file" "$GO_BASE/$file" || { echo "descarga falló" >&2; exit 1; }
  want="$(sha_of "$OS-$ARCH")"
  if command -v sha256sum >/dev/null 2>&1; then got="$(sha256sum "$TOOLS_DIR/$file" | awk '{print $1}')";
  else got="$(shasum -a 256 "$TOOLS_DIR/$file" | awk '{print $1}')"; fi
  [ "$got" = "$want" ] || { echo "sha256 NO coincide (esperaba $want): aborto" >&2; exit 1; }
  tar -xzf "$TOOLS_DIR/$file" -C "$TOOLS_DIR"
  export PATH="$TOOLS_DIR/go/bin:$PATH"
  echo "go listo: $(go version)"
}

port_free() { # $1=puerto -> 0 si libre
  curl -s -m 1 "http://127.0.0.1:$1/healthz" >/dev/null 2>&1 && return 1 || return 0
}

pick_port() {
  if [ -n "${PIZARRON_ADDR:-}" ]; then
    echo "$PIZARRON_ADDR" | sed -E 's/.*:([0-9]+)$/\1/'
    return
  fi
  p=8000
  while [ "$p" -le 8100 ]; do
    if port_free "$p"; then echo "$p"; return; fi
    p=$((p + 1))
  done
  echo "sin puerto libre 8000-8100" >&2; exit 1
}

alive_pid() { [ -n "${1:-}" ] && kill -0 "$1" 2>/dev/null; }

# Reutiliza si ya corre sano.
if [ -f "$PIDFILE" ]; then
  old="$(cat "$PIDFILE")"
  if alive_pid "$old"; then
    echo "pizarrón ya corre (PID $old). Exports sugeridos:"
    echo "export PIZARRON_URL=${PIZARRON_URL:-http://127.0.0.1:8000}"
    exit 0
  fi
fi

ensure_go
mkdir -p "$BIN_DIR" "$RUN_DIR"
echo "compilando $VERSION ..."
(cd "$ROOT" && CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=$VERSION" -o "$BIN" ./cmd/pizarron)

PORT="$(pick_port)"
case "$PORT" in *:* ) ADDR="$PORT";; *) ADDR="127.0.0.1:$PORT";; esac
export PIZARRON_URL="http://$ADDR"

echo "arrancando en $ADDR ..."
PIZARRON_ADDR="$ADDR" PIZARRON_LOG="$APPLOG" nohup "$BIN" > "$OUTLOG" 2>&1 &
echo $! > "$PIDFILE"

tries=0
until curl -s -m 2 "$PIZARRON_URL/healthz" >/dev/null 2>&1; do
  tries=$((tries + 1))
  if [ "$tries" -ge 20 ]; then echo "no levantó; ver $OUTLOG" >&2; tail -n 20 "$OUTLOG" >&2; exit 1; fi
  sleep 0.5
done

echo "verificando contrato ..."
PIZARRON_URL="$PIZARRON_URL" sh "$ROOT/api/contract.sh" || {
  echo "contrato FALLÓ; apagando. Log: $OUTLOG" >&2
  sh "$ROOT/agent/down.sh" >/dev/null 2>&1 || true
  exit 1
}

echo ""
echo "PIZARRON_LISTO url=$PIZARRON_URL viewer=$PIZARRON_URL/"
echo "export PIZARRON_URL=$PIZARRON_URL"
echo "skill: skills/pizarron/SKILL.md | atajo: sh agents/pz.sh next|read|write"
