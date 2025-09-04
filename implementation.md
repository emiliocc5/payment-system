# Implementación

## Estructura

# Estructura del Repositorio - Payment Platform

## Estructura General del Proyecto

```
payment-platform/
├── .gitignore
├── Makefile
├── README.md
├── design.md
├── implementation.md
├── docker-compose.yml
├── grafana-config/
│   └── datasources/
│       └── prometheus.yml
├── prometheus-config/
│   └── prometheus.yml
├── rabbitmq-config/
│   ├── enabled_plugins
│   ├── init_queue.sh
│   └── rabbitmq.conf
└── payment-wallet-service/
    ├── Dockerfile
    ├── go.mod
    ├── go.sum
    ├── cmd/
    │   └── main.go
    ├── configuration/
    │   ├── config.yaml
    │   └── config-docker.yaml
    ├── internal/
    │   ├── adapters/
    │   │   ├── http/
    │   │   │   ├── health.go
    │   │   │   ├── health_test.go
    │   │   │   ├── http.go
    │   │   │   ├── payments.go
    │   │   │   └── server.go
    │   │   ├── pubsub/
    │   │   │   └── rabbit/
    │   │   │       └── rabbit_pub.go
    │   │   └── storage/
    │   │       ├── errors.go
    │   │       └── postgresql/
    │   │           ├── balance.go
    │   │           ├── database.go
    │   │           └── payment.go
    │   └── core/
    │       ├── balance/
    │       │   └── service.go
    │       ├── domain/
    │       │   ├── balance.go
    │       │   ├── errors.go
    │       │   └── payment.go
    │       ├── payments/
    │       │   └── service.go
    │       └── ports/
    │           ├── balance.go
    │           ├── database.go
    │           ├── payments.go
    │           ├── publisher.go
    │           └── subscriber.go
    ├── migrations/
    │   ├── 1_initial_schema.up.sql
    │   └── 1_initial_schema.down.sql
    └── pkg/
        ├── config/
        │   └── config.go
        ├── logger/
        │   └── logger.go
        ├── signals/
        │   ├── posix.go
        │   ├── shutdown.go
        │   └── signal.go
        └── uidgen/
            └── uuid.go
```

## Descripción de Componentes

### Archivos Raíz

- **`.gitignore`**: Configuración de archivos ignorados por Git (binarios Go, archivos de cobertura, .env, etc.)
- **`Makefile`**: Comandos automatizados para Docker Compose (build, up, down, logs, clean, restart)
- **`README.md`**: Documentación principal del proyecto con objetivos y tareas pendientes
- **`design.md`**: Especificación técnica detallada con requisitos, arquitectura y escalabilidad
- **`docker-compose.yml`**: Orquestación de servicios (PostgreSQL, RabbitMQ, Kafka, Prometheus, Grafana)

### Configuraciones de Infraestructura

#### `grafana-config/`
- **`datasources/prometheus.yml`**: Configuración del datasource de Prometheus para Grafana

#### `prometheus-config/`
- **`prometheus.yml`**: Configuración de Prometheus con scraping de métricas de todos los servicios

#### `rabbitmq-config/`
- **`enabled_plugins`**: Plugins habilitados (management, prometheus)
- **`init_queue.sh`**: Script de inicialización de exchanges, queues y bindings
- **`rabbitmq.conf`**: Configuración del broker RabbitMQ

### Payment Wallet Service

#### Directorio Raíz del Servicio
- **`Dockerfile`**: Imagen Docker para el servicio
- **`go.mod`** y **`go.sum`**: Gestión de dependencias de Go

#### `cmd/`
- **`main.go`**: Punto de entrada principal con wiring de dependencias y migraciones

#### `configuration/`
- **`config.yaml`**: Configuración para entorno local
- **`config-docker.yaml`**: Configuración para entorno Docker

#### `internal/adapters/`

##### `http/`
- **`server.go`**: Servidor HTTP principal con configuración y rutas
- **`health.go`** y **`health_test.go`**: Endpoint de health check
- **`http.go`**: Utilidades para respuestas JSON y manejo de errores
- **`payments.go`**: Handler para la creación de pagos

##### `pubsub/rabbit/`
- **`rabbit_pub.go`**: Publisher de RabbitMQ para eventos de pagos iniciados

##### `storage/`
- **`errors.go`**: Errores específicos de la capa de storage
- **`postgresql/`**:
    - **`database.go`**: Conexión y manejo de transacciones de PostgreSQL
    - **`balance.go`**: Repositorio de balance de usuarios
    - **`payment.go`**: Repositorio de pagos

#### `internal/core/`

##### `domain/`
- **`balance.go`**: Entidad de balance de usuario
- **`errors.go`**: Errores de dominio del negocio
- **`payment.go`**: Entidades y DTOs relacionados con pagos

##### `ports/`
- **`balance.go`**: Interfaces para repositorio y servicio de balance
- **`database.go`**: Interface para manejo de transacciones
- **`payments.go`**: Interfaces para repositorio y servicio de pagos
- **`publisher.go`** y **`subscriber.go`**: Interfaces para pub/sub

##### Servicios de Negocio
- **`balance/service.go`**: Lógica de negocio para gestión de balance
- **`payments/service.go`**: Lógica de negocio para procesamiento de pagos

#### `migrations/`
- **`1_initial_schema.up.sql`**: Creación de tablas y datos iniciales
- **`1_initial_schema.down.sql`**: Rollback de migraciones

#### `pkg/` (Utilidades Compartidas)
- **`config/config.go`**: Parser de configuración YAML
- **`logger/logger.go`**: Configuración de logging estructurado
- **`signals/`**: Manejo de señales del sistema para graceful shutdown
- **`uidgen/uuid.go`**: Generador de UUIDs

## Arquitectura del Código

### Patrón de Arquitectura
El proyecto sigue **Clean Architecture** con las siguientes capas:

1. **External Layer** (`adapters/`): HTTP handlers, database repositories, message publishers
2. **Application Layer** (`core/`): Servicios de aplicación y lógica de negocio
3. **Domain Layer** (`domain/`): Entidades de dominio y reglas de negocio
4. **Ports** (`ports/`): Interfaces que definen contratos entre capas

### Infraestructura como Código
- **Docker Compose**: Orquestación completa del stack
- **Monitoring Stack**: Prometheus + Grafana para observabilidad
- **Message Brokers**: RabbitMQ (command queues) + Kafka (event streaming)
- **Database**: PostgreSQL con migraciones automáticas

### Características Técnicas
- **Event-Driven Architecture**: Comunicación asíncrona entre servicios
- **Database Transactions**: Consistencia ACID en operaciones críticas
- **Idempotency**: Prevención de pagos duplicados
- **Graceful Shutdown**: Manejo apropiado de señales del sistema
- **Health Checks**: Endpoints para verificación de estado de servicios

## Pendientes
Debido a que el tiempo para realizar el ejercicio era limitado, prioricé entregar menos funcionalidad pero código limpio 
y escalable, los próximos pasos a implementar serán

1. Código de processor-service para recibir eventos de rabbit y procesar contra pasarela de pagos
2. Código en payment-wallet para procesar eventos de kafka
3. Implementación de métricas de negocio
4. Mayor cobertura de unit testing
5. Linter
