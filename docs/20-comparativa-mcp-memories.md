# Comparativa: memorias MCP existentes vs Pizarrón Ligero

| Proyecto | Transporte | Deps / RAM | Aporte a copiar | Lo que evitamos |
|----------|------------|------------|-----------------|-----------------|
| `peekwez/mcp-blackboard` (FastAPI+SSE+Redis) | HTTP/SSE MCP | Python + Redis, imagen <90MB, scheduler APScheduler | SSE realtime, housekeeping (`remove_stale_files`) | Redis obligatorio, embeddings opcionales que encarecen |
| `doobidoo/mcp-memory-service` | REST + MCP + CLI + dashboard, OAuth | SQLite-vec/ONNX lite (~80% más chico que torch) | **Un servicio, todo transporte**. Funciona con LangGraph/CrewAI/AutoGen/Claude/OpenCode. Patrón a imitar en v2 | Vector DB por defecto; nosotros texto plano primero |
| `JJerryChoi/apex-memory` (46kB npm, SQLite FTS5) | MCP stdio | ms arranque, zero-infra | Ultra-ligero, decay/importancia, FTS5. Prueba que <50MB es viable | Solo stdio; nosotros HTTP puro base |
| `Beledarian/mcp-local-memory` | stdio + SSE remoto, SQLite+FTS5+vec | local-first, ~8320/sse | Híbrido FTS+vec con fallback a FTS, lifecycle auditable | Requiere Node, embedding download inicial |
| `scarletkc/Perenna` (Git-backed Markdown) | MCP stdio + Streamable HTTP | 1 repo git, índice Vexor reconstruible | Persistencia humana-legible (Markdown en git). Inspira nuestro append-only log `.log` legible | Git como DB frena `<1ms` en caliente; lo dejamos en background |

## Decisión
- Core v1: **ninguna de esas deps**. Solo stdlib Go, `sync.Map` + `atomic`, log append-only en background (como Perenna pero sin git en hot path).
- v2: wrapper MCP fino estilo `mcp-memory-service` (mismos 3 tools), sin Redis/vectores obligatorios.
- Persistencia: `pizarron.log` formato `ID\tTS\tLEN\nBODY` + snapshot. Reconstruible con `cat`, como Perenna pero más simple.
