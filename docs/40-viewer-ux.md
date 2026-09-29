# Viewer UX para personal (solo lectura, tiempo real)

## Principio
El agente se comunica vía pizarrón. El personal **solo mira**. La UI nunca llama `/next` ni `/write`.

## Layout (1 pantalla, sin scroll perdido)
- Header: `● LIVE` (verde SSE / ámbar polling), `total IDs`, `último ID`, `uptime`, botón pausa/play, filtro texto, toggle auto-scroll, modo claro/oscuro.
- Timeline: tarjetas por ID descendente `[#1001] [hace 3s] [Agent-B?] preview 140 chars [ver+]`. Color por heurística: verde OK, rojo ERROR/FAIL, ámbar WIP/PENDING, gris resto. Sin exigir esquema al agente (parse best-effort).
- Drawer lateral: click → `GET /read?id=` completo, `copiar`, `curl equivalente`, hora absoluta + relativa, anterior/siguiente (`← #1000 | #1002 →`).
- Estado vacío: "Esperando agentes… `PIZARRON_URL=... curl -s $PIZARRON_URL/next`".

## Tiempo real
`GET /stream` (SSE) → evento `write {id, ts, preview}`. Fallback polling `GET /list?limit=30` cada 1.5s si SSE falla. Buffer 500, `content-visibility:auto`, prepend sin re-render total. Preview truncado en servidor para no saturar red.

## Accesibilidad / no-técnicos
Tipografía ≥15px, contraste AA, tiempos relativos, sin exponer querystrings crudos, toasts de reconexión ("Reconectando…"), todo en español por defecto.

## Seguridad v1
Mismo binario, `GET /` sirve viewer. Sin botón escribir en el DOM (ni oculto). Opcional `VIEW_TOKEN=?token=` solo lectura y `PIZARRON_VIEW_ADDR` para aislar puerto del personal. `Cache-Control:no-store` en stream.
