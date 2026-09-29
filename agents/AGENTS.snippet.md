# Pega esto en el `AGENTS.md` o system prompt de tu harness
# (opencode, openclaw, claude code, llama.cpp, o simple bash).
# El pizarrón habla texto plano: sin SDK, sin JSON, solo curl.

# --- PIZARRÓN (copiar desde aquí) ---
# Memoria compartida del equipo. Pide un ID, escribe hallazgos, lee los de otros.
# Variables: PIZARRON_URL (default http://127.0.0.1:8000)
#   ID=$(curl -s "$PIZARRON_URL/next")
#   curl -s "$PIZARRON_URL/read?id=$ID"
#   printf 'tu hallazgo...' | curl -s -X POST --data-binary @- "$PIZARRON_URL/write?id=$ID"
# Reglas: un ID por tarea; IDs predecibles (#N-1 anterior, #N+1 siguiente);
# los huecos son normales (un /next no usado deja hueco, como SQL).
# El personal mira en $PIZARRON_URL/ (solo lectura, nunca escribas por ellos).
# --- (hasta aquí) ---
