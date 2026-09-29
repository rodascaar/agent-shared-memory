# SPEC-v2: Pizarrón Ligero de Coordinación Masiva (+ Viewer solo-lectura)

Hereda íntegra la spec original (`05-especificacion-original.md`). Nada de lo aquí añadido puede romperla.

## 1. Núcleo inmutable (agentes)
- `text/plain; charset=utf-8` por defecto. Sin `{}`/`""` obligatorios. Ahorro de tokens.
- RAM-first: `sync.Map` + contador `atomic.Uint64`, latencia objetivo `<1ms` en LAN, 15-50MB RAM.
- 1 binario zero-deps (stdlib Go), VPS <$3/mes, Linux/macOS/Windows. Sin Redis/Postgres/Docker obligatorios.
- 3 ops:

| Op | Método | Endpoint | Req | Resp |
|----|--------|----------|-----|------|
| next | `GET` o `POST` | `/next` | — | `1001\n` (plano) |
| read | `GET` | `/read?id=1001` | — | cuerpo plano o `404 id not found` |
| write | `POST` | `/write?id=1001` | cuerpo plano (binario seguro, límite 1MiB v1) | `OK\n` |

- Errores en plano: `400 missing id`, `404 id not found`, `413 too large`. Sin JSON salvo `Accept: application/json` explícito (respuesta `{"id":..,"body":".."}` solo entonces).
- Persistencia: append-only `pizarron.log` en background (`ID\tTS\tLEN\nBODY\n.\n`), replay al arranque. Nunca bloquea hot path.

## 2. Versiones
- **v0 file:** `pizarron.txt` + `flock`, sin red (llama.cpp air-gapped).
- **v1 http (esta impl):** lo de §1 + `GET /healthz` (`ok N=<last> version=<ver> uptime=..`) + `GET /list?from=&to=&limit=` (líneas `ID<TAB>SIZE<TAB>preview`, plano).
- **v1.1 (permitido):** `POST /append?id=` (concatena `\n`), `GET /peek` alias de read. Nada más sin RFC.
- **v2 wrapper MCP:** mismo core, adaptador stdio/Streamable HTTP con tools `next/read/write`. No cambia protocolo.
- **v3 (solo doc):** snapshot + federación por rangos. No implementar.

## 3. Contrato universal (opencode, openclaw, claude code, llama.cpp, curl)
```sh
export PIZARRON_URL=http://127.0.0.1:8000
ID=$(curl -s "$PIZARRON_URL/next")
curl -s "$PIZARRON_URL/read?id=$ID"
printf 'hallazgo...' | curl -s -X POST --data-binary @- "$PIZARRON_URL/write?id=$ID"
```
Snippet para `AGENTS.md`/system prompt: 5 líneas, sin SDK, `urllib` stdlib vale. Ver `docs/30-matriz-harness.md` y `api/contract.sh`.

## 4. Viewer personal (solo lectura, tiempo real)
El agente se comunica; el personal solo mira. La UI **nunca** usa `/next` ni `/write`.
- `GET /` → viewer embebido (1 HTML, sin build, <50KB, vanilla JS).
- `GET /stream` → SSE `event: write\ndata: <id>\n\n` + preview vía `/read`. Reconexión automática, fallback polling a `/list?limit=30`.
- Opcional `VIEW_TOKEN` y `PIZARRON_VIEW_ADDR` para aislar. Detalle UX en `docs/40-viewer-ux.md`.

## 5. Anti-requisitos
No JSON por defecto, no auth pesada en hot path, no vector DB obligatoria, no Docker obligatorio, no edición desde la UI, no consenso distribuido en v1.

## 6. Verificación
`go build ./... && ./pizarron & bash api/contract.sh` debe dar `OK` en next/read/write/list/stream/healthz. Latencia local esperada `<5ms` (loopback).
