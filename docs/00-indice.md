# Índice de investigación — Pizarrón Ligero

Idea central (inmutable): memoria compartida async, `text/plain`, IDs secuenciales atómicos, RAM-first `<1ms`, 1 binario zero-deps. Spec original: `05-especificacion-original.md`.

| Doc | Tema | Pregunta que responde |
|-----|------|-----------------------|
| `10-historia-blackboard-linda.md` | Hearsay-II (70s), Linda (Gelernter 85), Fowler accidental blackboard, arxiv 2510.18893 / 2507.01701 | ¿Por qué espacio compartido desacoplado y no mensajes directos? |
| `20-comparativa-mcp-memories.md` | mcp-blackboard, mcp-memory-service, apex-memory, mcp-local-memory, Perenna | ¿Qué copiar (multi-transporte, ultra-ligero) y qué evitar (Redis, Docker obligatorio, JSON pesado)? |
| `30-matriz-harness.md` | opencode, openclaw, claude code, llama.cpp, curl/python puro | ¿Cómo invoca cada harness el contrato sin SDK? |
| `40-viewer-ux.md` | UI solo-lectura tiempo real para personal | ¿Qué ve el personal sin poder escribir? |
| `SPEC-v2.md` | Spec consolidada | Contrato, versiones, anti-requisitos |
| `../api/contract.sh` | Contrato shell testeable | Prueba universal con curl |

Fuentes web consultadas (sep-2026): martinfowler.com/articles/exploring-gen-ai/an-accidental-blackboard, arxiv 2510.18893 / 2507.01701, github peekwez/mcp-blackboard, doobidoo/mcp-memory-service, JJerryChoi/apex-memory, Beledarian/mcp-local-memory, scarletkc/Perenna.
