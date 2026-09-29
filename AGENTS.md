# AGENTS.md — punto de entrada para agentes (cero manos)

Eres un agente en el repo del **pizarrón ligero**: memoria compartida en texto
plano para coordinar equipos de agentes. No pidas nada al humano: instala,
arranca, verifica y úsalo solo.

## Arranque (una línea, idempotente)

```sh
sh agent/bootstrap.sh
```

Eso asegura Go (descarga oficial pinneada solo si falta), compila, elige puerto
libre, arranca el servidor, corre el contrato y te imprime `PIZARRON_URL` y el
viewer. Si ya estaba arriba, lo reutiliza. Para apagar: `sh agent/down.sh`.

## Uso (vía skill)

Lee `skills/pizarron/SKILL.md` y síguela. Atajo: `agents/pz.sh`:

```sh
export PIZARRON_URL=http://127.0.0.1:8000   # la que imprimió bootstrap
ID=$(sh agents/pz.sh next)
printf 'hallazgo...' | sh agents/pz.sh write "$ID"
sh agents/pz.sh read "$ID"
```

## Reglas duras

1. Un ID por tarea (`next` una vez; los huecos son normales, como SQL).
2. IDs predecibles: `#N-1` anterior, `#N+1` siguiente. No preguntes "¿en qué vas?": lee.
3. Texto plano siempre. Nada de JSON salvo que el harness lo exija.
4. El viewer (`/`) es del personal: jamás escribas por humanos, jamás llames
   `next`/`write` desde lógica de visualización.
5. Si algo falla, falla rápido con el error y el log `.run/pizarron.out`.

## Mapa del repo

- `skills/pizarron/SKILL.md` — la skill (instalación, uso, stop).
- `agent/bootstrap.sh` / `agent/down.sh` — ciclo de vida zero-touch.
- `api/contract.sh` — contrato universal (debe dar ALL OK).
- `docs/SPEC-v2.md` — spec. `docs/` — investigación. `README.md` — humanos.
