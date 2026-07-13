# proxi-pulse 🌐

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-E6522C?logo=prometheus)](https://prometheus.io)
[![Grafana](https://img.shields.io/badge/Dashboard-Grafana-F46800?logo=grafana)](https://grafana.com)

**Reverse proxy con geolocalización de visitantes, métricas en vivo y dashboard Grafana.**

Sabé de dónde viene cada visita a tus sitios sin meter scripts de tracking.
El proxy captura país, ciudad, latencia y más — automaticamente.

## Stack

| Capa | Tecnología |
|------|-----------|
| Proxy | Go `net/http/httputil.ReverseProxy` |
| GeoIP | MaxMind GeoLite2 + LRU cache |
| Métricas | Prometheus `client_golang` |
| Dashboard | Grafana (pre-provisionado) |
| Admin UI | Go templates + HTMX |
| Almacenamiento | SQLite (modernc.org — pure Go) |
| Infra | Docker + docker-compose |

## Quick Start

```bash
# 1. Clonar
git clone https://github.com/josesojo2828/proxi-pulse.git
cd proxi-pulse

# 2. Descargar GeoLite2-City.mmdb
#    Registrarse en https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
#    y copiar el archivo a ./data/GeoLite2-City.mmdb

# 3. Configurar (opcional — hay defaults)
cp config.yaml.example config.yaml
# editar según necesidad

# 4. Levantar el stack
docker compose up -d

# 5. Abrir URLs:
#    Admin UI:  http://localhost:8080
#    Grafana:   http://localhost:3000 (admin:admin)
#    Prometheus: http://localhost:9090
```

## Admin UI

Interfaz web simple para administrar los sitios:

- **Dashboard** — resumen de stats
- **Sites** — CRUD de dominios y upstreams
- **Visits** — tabla paginada de visitas recientes

## API REST

Endpoint para integraciones:

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /api/sites | Listar sites |
| POST | /api/sites | Crear site `{"domain":"...","upstream":"..."}` |
| GET | /api/sites/{id} | Obtener site |
| PUT | /api/sites/{id} | Actualizar site |
| DELETE | /api/sites/{id} | Eliminar site |
| GET | /api/stats/overview | Stats generales |
| GET | /api/stats/visits?page=&limit=&site_id= | Visitas paginadas |
| GET | /api/stats/countries | Visitas por país |

## Métricas Prometheus

Expuestas en `/metrics`:

| Métrica | Tipo | Labels |
|---------|------|--------|
| proxi_pulse_requests_total | Counter | host, country, status |
| proxi_pulse_request_duration_ms | Histogram | host, country |
| proxi_pulse_active_connections | Gauge | — |
| proxi_pulse_geo_lookup_duration_ms | Histogram | — |
| proxi_pulse_sites_total | Gauge | — |

## Dashboard Grafana

Incluye dashboard pre-provisionado con:
- 🌍 Mapa mundial de tráfico por país
- 📊 Top países y hosts
- 📈 RPS y latencia P50/P95
- 🔢 Conexiones activas
- 📋 Top user agents

## Docker Compose

```
docker compose up -d
┌── proxi-pulse :8080   → Proxy + Admin UI + /metrics
├── prometheus  :9090   → Almacenamiento de métricas
└── grafana     :3000   → Dashboards (provisionados)
```

## Desarrollo

```bash
make build    # Compilar
make run      # Ejecutar local
make test     # Tests
make clean    # Limpiar binarios
```

## Licencia

MIT
