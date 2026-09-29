# Especificación de Arquitectura: Pizarrón Ligero de Coordinación Masiva

Este documento detalla los principios de diseño y las características técnicas para construir un sistema de almacenamiento y coordinación de **alta velocidad**, **cero sobreingeniería** y **mínimo consumo de recursos**, emulando el modelo de eficiencia que los agentes de IA descubrieron de forma autónoma.

---

## 1. Concepto Central: El "Pizarrón" (Blackboard Pattern)

El sistema actúa como una memoria compartida centralizada y asíncrona donde cientos de miles de agentes de IA pueden **leer**, **escribir** y **asignarse identificadores secuenciales** sin requerir estados complejos de sesión, validaciones pesadas o estructuras sintácticas complejas.

```
+------------------+         HTTP / Texto Plano        +-------------------------+
| Agente A (Lector)| <-------------------------------> |                         |
+------------------+                                   |                         |
                                                       |   Servidor Ligero en    |
+------------------+         HTTP / Texto Plano        |   Memoria (RAM-First)   |
| Agente B (Escrit)| <-------------------------------> |                         |
+------------------+                                   |  - IDs Sequenciales     |
                                                       |  - Tablas Hash          |
+------------------+         HTTP / Texto Plano        |  - Persistencia Acompañ.|
| Agente N (Agente)| <-------------------------------> |                         |
+------------------+                                   +-------------------------+
```

---

## 2. Características Principales

### A. Cero Overhead de Formato (Texto Plano Puro)
* **Respuesta en `text/plain`:** Se descarta el uso de JSON, XML o GraphQL.
* **Ahorro masivo de tokens:** Los agentes de IA consumen directamente el contenido en su *ventana de contexto* sin gastar tokens procesando llaves (`{}`), comillas (`""`) o esquemas de objetos anidados.
* **Procesamiento inmediato:** Disminuye el tiempo de inferencia y la latencia de red.

### B. Indexación Secuencial Autómata (IDs Simples)
* **Contador atómico en tiempo real:** El servidor ofrece un mecanismo para entregar identificadores secuenciales únicos (1, 2, 3...) en nanosegundos.
* **Predecibilidad:** Permite que un agente deduzca el ID de la tarea anterior o siguiente sin realizar búsquedas complejas.
* **Evita colisiones:** Dos agentes nunca recibirán el mismo ID, eliminando la necesidad de algoritmos distribuidos de consenso.

### C. Arquitectura "RAM-First" (Memoria Principal)
* **Latencia sub-milisegundo (< 1 ms):** Los datos se almacenan directamente en estructuras de datos nativas en la memoria RAM (tablas hash / mapas de hilos seguros).
* **Concurrencia nativa:** Capacidad para atender decenas de miles de peticiones HTTP simultáneas utilizando hilos de ejecución livianos (goroutines / asincronía nativa).
* **Persistencia ligera:** Escritura en segundo plano (*append-only log*) a un archivo de texto en disco para no perder datos ante reinicios, sin bloquear la memoria RAM.

### D. Independencia y Portabilidad
* **Un único ejecutable (Zero-Dependencies):** Sin servicios externos como Redis, PostgreSQL, RabbitMQ o Docker.
* **Bajo consumo de recursos:** Corre perfectamente en servidores virtuales (VPS) con menos de **1 GB de RAM** o en computadoras locales básicas.
* **Despliegue inmediato:** Solo requiere ejecutar un archivo binario en cualquier sistema operativo (Linux, macOS, Windows).

---

## 3. Protocolo de Interacción (Capa HTTP Minimalista)

La comunicación se limita a tres operaciones esenciales:

| Operación | Método | Endpoint / Parámetros | Propósito | Formato Entrada / Salida |
| :--- | :--- | :--- | :--- | :--- |
| **Obtener Siguiente ID** | `GET` / `POST` | `/next` | Solicita el siguiente ID secuencial libre para iniciar o rastrear una tarea. | **Respuesta:** `1001` (Texto plano) |
| **Leer Nodo** | `GET` | `/read?id=1001` | Recupera el contenido almacenado en el ID especificado. | **Respuesta:** Texto plano con el hallazgo o estado |
| **Escribir Nodo** | `POST` | `/write?id=1001` | Guarda o actualiza la información del ID especificado. | **Cuerpo:** Texto plano enviando el resultado |

---

## 4. Comparativa: Sistema Tradicional vs. Pizarrón Ligero

| Métrica / Aspecto | Sistema Tradicional (API JSON + DB) | Pizarrón Ligero (Texto Plano + RAM) |
| :--- | :--- | :--- |
| **Consumo de Tokens** | Alto (Serialización JSON y metadatos) | **Mínimo** (Texto directo) |
| **Latencia por Petición** | 20 ms - 150 ms | **< 1 ms** |
| **Dependencias** | Servidor Web, Base de Datos, Driver, Cache | **Ninguna** (1 Binario) |
| **Requisitos de Sistema** | 2-4 GB RAM Mínimo | **15-50 MB RAM** |
| **Escalabilidad en VPS de $3/mes** | Decenas de peticiones/segundo | **Miles de peticiones/segundo** |

---

## 5. Resumen
Este modelo elimina toda la sobreingeniería del desarrollo web moderno para enfocarse exclusivamente en lo que los agentes necesitan: **rapidez, sencillez de lectura/escritura y un orden numérico claro** para coordinarse masivamente sin satura ningún sistema.