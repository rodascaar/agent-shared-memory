# Pizarrón ligero — memoria compartida para agentes, visible para personas

Un binario, cero dependencias, texto plano. Cientos de agentes se coordinan con
3 operaciones; el personal mira en tiempo real desde el navegador (solo lectura).

## Clona y dile a tu agente (cero manos)

```sh
git clone <este-repo> && cd <este-repo>
# dile a tu agente (opencode, claude code, openclaw, ...):
# "Lee AGENTS.md y deja el pizarrón corriendo"
```

El agente ejecuta `sh agent/bootstrap.sh` (instala Go oficial solo si falta,
compila, arranca, verifica) y usa `skills/pizarron/SKILL.md`. No hay nada que
hacer a mano. Detalle agente: `AGENTS.md`.

## En 30 segundos (humanos)

```sh
# 1. Instalar (macOS/Linux, ~/.local/bin)
sh install.sh
# — o compilar tú mismo (requiere Go >= 1.21):
# make build

# 2. Arrancar (manual, queda en primer plano; Ctrl-C para parar)
PIZARRON_ADDR=127.0.0.1:8000 pizarron

# 3. Usar (cualquier lenguaje, aquí curl)
export PIZARRON_URL=http://127.0.0.1:8000
ID=$(curl -s "$PIZARRON_URL/next")                                  # → 1
printf 'hallazgo...' | curl -s -X POST --data-binary @- "$PIZARRON_URL/write?id=$ID"
curl -s "$PIZARRON_URL/read?id=$ID"

# 4. Mirar (personal, auto-actualiza)
open http://127.0.0.1:8000/
```

## Para agentes (todos los harnesses)

Pega `agents/AGENTS.snippet.md` en tu `AGENTS.md` o system prompt. Funciona en
opencode, openclaw, claude code, llama.cpp y bash puro: el único requisito es
poder ejecutar `curl` (o cualquier HTTP). Detalle por harness en
`docs/30-matriz-harness.md`.

## Configuración (todo opcional, todo por entorno)

| Variable        | Default            | Para qué                                  |
|-----------------|--------------------|-------------------------------------------|
| `PIZARRON_ADDR` | `127.0.0.1:8000`   | Dónde escucha (localhost = seguro)        |
| `PIZARRON_LOG`  | `./pizarron.log`   | Archivo append-only de persistencia       |
| `VIEW_TOKEN`    | (vacío = abierto)  | `?token=...` para ver `/` y `/stream`     |

`pizarron -version` imprime la versión. `/healthz` la incluye (`ok N=.. version=..`).

## Verificar que todo funciona

```sh
make test        # vet + tests unitarios (<1s)
make test-e2e    # levanta el binario y corre api/contract.sh
make build-all   # binarios darwin/linux (+windows) en dist/ + SHA256SUMS
```

## Desinstalar

```sh
rm "$(which pizarron)" ./pizarron.log   # eso es todo; no hay servicios ni DB
```

## Diseño (resumen)

`text/plain` por defecto (ahorra tokens, sin `{}`), IDs secuenciales atómicos,
RAM-first (`<1ms`), persistencia append-only en background con replay al
arranque. Spec completa: `docs/SPEC-v2.md`. Investigación y comparativas: `docs/`.
