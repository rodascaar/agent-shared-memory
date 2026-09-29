#!/bin/sh
# pz.sh — wrapper fino del pizarrón (thin-over-curl, sin magia).
# Requiere: curl y PIZARRON_URL. Con --show imprime el curl exacto en vez de correrlo.
#   sh agents/pz.sh next | read ID | write ID | append ID | list [N] | health | url
set -eu
SHOW="no"
if [ "${1:-}" = "--show" ]; then SHOW="yes"; shift; fi
U="${PIZARRON_URL:-http://127.0.0.1:8000}"
cmd="${1:-help}"; shift 2>/dev/null || true

run() { if [ "$SHOW" = "yes" ]; then echo "+ $*"; else "$@"; fi; }

case "$cmd" in
  next)   run curl -s "$U/next";;
  read)   run curl -s "$U/read?id=${1:?falta ID}";;
  write)  run curl -s -X POST --data-binary @- "$U/write?id=${1:?falta ID}";;
  append) run curl -s -X POST --data-binary @- "$U/append?id=${1:?falta ID}";;
  list)   run curl -s "$U/list?limit=${1:-30}";;
  health) run curl -s "$U/healthz";;
  url)    echo "$U";;
  *) echo "uso: sh agents/pz.sh [--show] next|read ID|write ID|append ID|list [N]|health|url" >&2; exit 1;;
esac
