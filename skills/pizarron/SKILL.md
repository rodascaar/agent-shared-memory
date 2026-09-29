---
name: pizarron
description: Shared blackboard memory for agent teams — sequential IDs, plain-text notes, sub-ms, human read-only viewer. Use when coordinating multi-step work across agents or sessions, sharing findings, or when the user mentions pizarron, blackboard, or shared agent memory.
---

# Pizarrón — skill de coordinación entre agentes

Memoria compartida en texto plano. Tres operaciones, cero dependencias.
Si el servidor no está arriba, arráncalo tú (no pidas nada al humano).

## 1. Arrancar (idempotente)

```sh
sh agent/bootstrap.sh   # imprime PIZARRON_URL y viewer; reutiliza si ya corre
```

Guarda la `PIZARRON_URL` impresa y úsala en todo lo que sigue. Para apagar:
`sh agent/down.sh`.

## 2. Operaciones

```sh
export PIZARRON_URL=http://127.0.0.1:8000
ID=$(sh agents/pz.sh next)                          # reserva ID (consume aunque no escribas)
printf 'tu hallazgo...' | sh agents/pz.sh write "$ID"  # publica
sh agents/pz.sh read "$ID"                          # lee (también #anterior/#siguiente)
sh agents/pz.sh list 20                             # últimos 20 (para el viewer, no para ti)
```

Equivalencias curl directas (si tu harness no corre shell scripts):

```sh
ID=$(curl -s "$PIZARRON_URL/next")
printf '...' | curl -s -X POST --data-binary @- "$PIZARRON_URL/write?id=$ID"
curl -s "$PIZARRON_URL/read?id=$ID"
```

## 3. Protocolo de equipo

- **Un ID por tarea.** Llama `next` una vez al empezar; escribe al terminar.
- **Lee antes de actuar:** `#N-1` anterior, `#N+1` siguiente. Los huecos son
  normales (un `next` sin `write` deja hueco, como SQL).
- **Texto plano, corto, con prefijo de rol:** `Agent-Search: ...`,
  `Agent-Build: ...`, `RESUELTO: ...` al cerrar una línea de trabajo.
- **Referencia cruzada:** cita `#ID` en tus notas para trazar decisiones.
- **Nada de JSON** salvo que tu harness lo exija (`Accept: application/json`).
- **El viewer `/` es del personal (solo lectura).** Nunca llames `next`/`write`
  desde código de visualización ni escribas en nombre de humanos.

## 4. Fallos (fail-fast)

- `curl` exit 7 / timeout → servidor caído: corre `sh agent/bootstrap.sh`.
- `missing id` (400) → olvidaste `?id=`. `id not found` (404) → hueco o ID futuro.
- Si `api/contract.sh` no da `ALL OK`, detente y muestra `.run/pizarron.out`.
