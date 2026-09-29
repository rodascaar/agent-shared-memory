# Historia: Blackboard → Linda → Agentes LLM

## 1. Hearsay-II (Erman et al., 1970s, speech recognition)
Pizarra central donde especialistas (KS: acústica, sintaxis, semántica) leen/escriben hipótesis por niveles. Sin llamadas directas entre ellos. Lección: **coordinación por observación, no por mensajes**.

## 2. Linda / Tuple spaces (Gelernter, 1985)
Primitivas `out/in/rd` sobre espacio compartido asociativo. Desacopla en tiempo, espacio y referencia. Lección directa para el pizarrón: **IDs o tuplas predecibles > búsqueda compleja; atomicidad del `in` evita colisiones** (nuestro `/next` atómico es el `in` simplificado a contador).

## 3. Fowler: "An Accidental Blackboard" (2024-2025)
Describe cómo sistemas multi-agente LLM redescubren la pizarra sin planearlo: scratch files / contexto compartido como fuente de verdad. Advertencia: sin IDs secuenciales ni RAM-first degenera en caos de archivos. Nuestro aporte: **contador atómico + texto plano** lo hace deliberado y barato en tokens.

## 4. Papers LLM-MAS recientes
- `arxiv 2510.18893`: observation-driven coordination (Linda + blackboard + stigmergy) aplicada a agentes autónomos.
- `arxiv 2507.01701`: blackboard architecture para LLM-MAS con roles (planner, researcher, writer...). Confirma que compartir todo el contexto funciona mejor que pasar mensajes.
- Zylos `Awareness Boards Revisited` (2026): "reduce the mechanism before you build it" — mismo espíritu que nuestra spec: quitar JSON/DB antes de escalar.

## Conclusión para versiones
v0 (fichero) = Linda mínima. v1 (HTTP+RAM) = Hearsay-II sin niveles fijos. v2 (wrapper MCP) = compatibilidad sin traicionar el mecanismo. No añadir consenso distribuido, embeddings obligatorios ni grafo pesado: eso ya lo hacen otros y rompe `<1ms / 15-50MB`.
