#!/bin/sh
# down.sh — apaga el pizarrón levantado por bootstrap.sh (por pidfile).
# Uso: sh agent/down.sh
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIDFILE="$ROOT/.run/pizarron.pid"

if [ ! -f "$PIDFILE" ]; then echo "no hay pidfile: nada que apagar"; exit 0; fi
pid="$(cat "$PIDFILE")"
if kill -0 "$pid" 2>/dev/null; then
  kill "$pid"
  tries=0
  while kill -0 "$pid" 2>/dev/null && [ "$tries" -lt 20 ]; do sleep 0.25; tries=$((tries + 1)); done
  if kill -0 "$pid" 2>/dev/null; then echo "PID $pid no murió (mátalo a mano)" >&2; exit 1; fi
  echo "apagado (PID $pid)"
else
  echo "PID $pid ya no corría; limpiando pidfile"
fi
rm -f "$PIDFILE"
