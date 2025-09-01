# Payment platform

## Objetivo
Crear una plataforma que permita a los usuarios el pago de servicios con gestion de balance incluida

## Diseño e implementación

- [Diseño](design.md)
- [Implementación](implementation.md)

## Instalación y Ejecución

### Prerrequisitos
- Docker
- Docker Compose

### Comandos Principales

```bash
# Construir sin cache
docker compose build --no-cache

# Ejecutar aplicación
docker compose up -d

# Detener servicios
docker compose down
```

### Acceso a Servicios

- **Grafana**: http://localhost:3000 (admin/admin)
- **RabbitMQ Management**: http://localhost:15672 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Kafka UI**: http://localhost:8080
- **Payment Wallet API**: http://localhost:5555


---
**Autor**: [Emilio Nicolas Caccia Campaner]  
**Versión**: 1.0.0  
**Fecha**: Septiembre 2025





