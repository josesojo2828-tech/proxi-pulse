# proxi-pulse 🌐

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-E6522C?style=flat&logo=prometheus)
![License](https://img.shields.io/badge/License-MIT-green)

**Un reverse proxy HTTP/HTTPS con geolocalización, métricas en vivo y observabilidad.**

Sabés de dónde viene cada request, cuánto tardó, y adónde fue. Pensado para homelabs,
entornos de producción chicos, y cualquiera que quiera tener visibilidad posta de su
infraestructura sin morir en el intento.

Parte del ecosistema de monitoreo junto con
[net-probe-exporter](https://github.com/tuusuario/net-probe-exporter).

## Características

- **Geo-routing**: identificá país, ciudad e ISP de cada request con MaxMind GeoLite2
- **Métricas Prometheus**: RPS, latencia P50/P95/P99, status codes — todo etiquetado por geolocalización
- **Dashboard Grafana ready**: conectás Prometheus, importás el JSON y tenés un mapa mundial de tráfico
- **Ligero**: Go puro, sin frameworks, `httputil.ReverseProxy` de la stdlib, nada de magia rara
- **Docker**: `docker-compose up` y tenés el proxy + Prometheus + Grafana funcionando

## Stack

| Capa | Tecnología |
|------|-----------|
| Lenguaje | Go 1.25+ |
| Proxy | `net/http/httputil.ReverseProxy` |
| GeoIP | MaxMind GeoLite2 |
| Métricas | prometheus/client_golang |
| Config | YAML |
| Infra | Docker + docker-compose |

## Quick Start

```bash
# TODO: cuando el proyecto esté implementado, acá van los comandos
```

Por ahora está en construcción, bro. Dale una estrella y segui el repo si te interesa.

## Roadmap

- [x] Repo + README (esto)
- [ ] **v1**: Proxy básico + GeoIP + métricas Prometheus + dashboard Grafana
- [ ] **v2**: Logging estructurado + Loki para trazabilidad
- [ ] **v3**: Auto SSL / Let's Encrypt (porque en 2026 no tener HTTPS es pecado)
- [ ] **v4**: Dashboard web propio con mapa de calor en vivo

## Licencia

[MIT](LICENSE) — hacé lo que quieras, si te sirve, joya.
