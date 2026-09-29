#!/bin/sh
# Contrato universal del pizarrón — funciona en opencode, openclaw, claude code, llama.cpp, curl puro.
# Uso: PIZARRON_URL=http://127.0.0.1:8000 sh api/contract.sh
set -eu
U="${PIZARRON_URL:-http://127.0.0.1:8000}"
fail(){ echo "FAIL: $1" >&2; exit 1; }
echo "== pizarron contract vs $U =="

ID=$(curl -s -f "$U/next") || fail "GET /next"
echo "next -> $ID"
case "$ID" in ''|*[!0-9]*) fail "next no numérico: $ID";; esac

printf '%s' "Agent-A: hallazgo $ID" | curl -s -f -X POST --data-binary @- "$U/write?id=$ID" >/dev/null || fail "POST /write"
echo "write OK"

GOT=$(curl -s -f "$U/read?id=$ID") || fail "GET /read"
[ "$GOT" = "Agent-A: hallazgo $ID" ] || fail "read mismatch: $GOT"
echo "read OK"

H=$(curl -s -f "$U/healthz") || fail "GET /healthz"
echo "healthz -> $H"

L=$(curl -s -f "$U/list?limit=5") || fail "GET /list"
echo "list OK (muestra):"; echo "$L" | head -n 3

# SSE: esperamos 1 evento hello/write en <5s
if curl -s -N --max-time 4 "$U/stream" | grep -qE "^(event: hello|event: write|: ping)"; then
  echo "stream OK"
else
  fail "GET /stream sin eventos"
fi

# Viewer solo-lectura sirve HTML sin exponer /write en el DOM como formulario
V=$(curl -s -f "$U/") || fail "GET / (viewer)"
echo "$V" | grep -q "solo lectura" || fail "viewer sin marca solo-lectura"
echo "$V" | grep -q "/write" && fail "viewer no debe referenciar /write" || true
echo "viewer OK (sin escrituras)"

echo "ALL OK — ID=$ID"
