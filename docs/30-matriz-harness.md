# Matriz harness: cómo usa cada uno el pizarrón sin SDK

Contrato base (único denominador común): `PIZARRON_URL` + curl + `text/plain`. Sin pip/npm obligatorio.

## 1. curl / python stdlib (referencia, funciona en todos lados)
```sh
ID=$(curl -s "$PIZARRON_URL/next")
curl -s "$PIZARRON_URL/read?id=$ID"
echo -n "hallazgo..." | curl -s -X POST --data-binary @- "$PIZARRON_URL/write?id=$ID"
```
Python sin deps: `urllib.request` a los mismos 3 endpoints.

## 2. llama.cpp (llama-server, sin MCP)
El modelo no tiene tool-calling nativo fiable. Vía prompt del sistema + `bash` del harness que corre el modelo:
- System prompt inyecta las 3 líneas curl de arriba.
- El loop del harness ejecuta shell y devuelve stdout al contexto. Cero JSON parseado por el modelo: lee texto plano directo (ahorro de tokens = tesis de la spec).

## 3. opencode
- Plugin local `~/.config/opencode/plugins/pizarron.js` o simple `bash` tool: expone `pizarron_next/read/write` que por dentro hacen curl. Alternativa sin plugin: instrucción en `AGENTS.md` con las 3 líneas.
- MCP v2 (cuando exista wrapper): añadir a `opencode.json` como remote HTTP, sin cambiar core.

## 4. openclaw / claude code
- Hoy: `fetch`/`exec(curl)` directo al HTTP puro (recomendado v1).
- v2: `claude mcp add pizarron -- <bin> mcp-stdio` o URL Streamable HTTP. Mismos 3 tools, mismos textos.

## 5. Regla anti-rotura
Si un harness exige JSON-schema, el servidor acepta `Accept: application/json` como **negociación opcional** (`{"id":1001,"body":"..."}`), pero el default sigue `text/plain`. El agente nunca paga tokens de llaves/comillas salvo que el harness lo pida.
