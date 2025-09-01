# Diseño

## Requisitos funcionales

- Un usuario debe poder pagar un servicio
- Un usuario debe poder ver su saldo disponible

## Requisitos no funcionales

- Alta disponibilidad con latencia mínima.
- El sistema debe poder ser escalable y eficiente.

## Assumptions

- La autenticación se realizará por algún third-party service, ej Auth0.
- El foco del sistema en una etapa inicial se centra en manejar correctamente la transacción de pago en lugar de aspirar a un alto rendimiento, es decir, la clave del sistema estará en no procesar pagos por duplicado ni dejar transacciones inconsistentes
- Solo se diseña el "pay-in" flow
    - Se deja de lado la conciliación con el proveedor externo
- Se asume un único currency
- Se asume que existe un repositorio de usuarios
- Se asume que existe un repositorio de entidades de pago habilitadas


## Estimaciones

- TPD (transactions per day): 1 millón
- RPS (request per second): **1M/(24h * 3600s) = ~10**.
- Tamaño promedio de un pago: 250 bytes
- Storage por día: **1M * 250 bytes = ~240MB/día**

## Elección de la base de datos

Para un primer diseño, debido a la características de los datos y las relaciones que presentan entre sí, consideraremos **PostgreSQL**.

## Modelo de datos

![datamodel](./diagrams/datamodel.png)

## Diseño inicial

![initial](./diagrams/initial.png)

- Inicialmente se proponen los siguientes componentes
  - Payment-Wallet-Service: 
    - Este servicio es el encargado de gestionar el balance del usuario y las transacciones. Se encarga de congelar el saldo comprometido durante el ciclo de vida de la transacción,
      crear la transacción y publicar un evento. Se encarga tambien de escuchar eventos de resultado de las operaciones para aplicar un estado final a una transacción  e imputar el balance correspondiente.
  - Payment-Processor-Service: 
    - Este servicio es el encargado de procesar los pagos iniciados, comunicarse con las pasarelas de pago externas
      y procesar las respuestas de las mismas. Cuando obtiene respuesta de la pasarela externa, publica un evento
      con el resultado de dicha operación, que será procesado por el servicio payment-wallet
  - Ambos servicios pushearán métricas a Prometheus y serán visualizadas en Grafana, que lo tendrá configurado como datasource
  - Payment-Wallet al crear el pago, publica el evento en una queue de RabbitMQ
  - Payment-Processor recibe el mensaje y procesa contra la pasarela de pago, de forma síncrona se entera del estado del pago y publica en un tópico de kafka
  - Payment-Wallet procesará el mensaje del tópico de Kafka y finalizará la transacción, con el estado correspondiente informado por payment-processor

## Diagrama de flujo de eventos

![flow](./diagrams/flow.png)

## API

- `POST /payments`
    - Crea un nuevo pago
- `POST /health`
    - Retorna el estado del servidor.
  
## Escalabilidad del diseño

A medida que nuestro negocio escala, es necesario implementar medidas para que nuestro sistema pueda soportar la carga.

Una primer medida puede ser agregar instancias de nuestro server y un load balancer para poder balancear la carga que tiene cada una de las mismas. Podría elegirse un algoritmo adaptativo para ajustarse a la carga actual, la capacidad del servidor, etc.

En cuanto a la base de datos, si bien podemos escalarla verticalmente, esto se volvería un límite eventualmente. Por eso, optaremos por escalar horizontalmente mediante la técnica de sharding, particionando la data en diferentes instancias. Para asignar a cuál instancia se guardaría un nuevo dato, podríamos usar consistent hashing para tomar la decisión.

Nuestro diseño inicial quedaría de la siguiente manera:

![intermediate](./diagrams/intermediate.png)

Si quisieramos escalar aún más, llevando el análisis un poco más cerca del mundo real, podríamos sacar las siguientes conclusiones:

- Partir nuestro monolito modular en diferentes microservicios. Esto permitiría usar distintas bases de datos según el caso de uso. Deberíamos primero definir los boundaries de cada dominio. Una opción sería:
    - **Users Service**: se encarga de dar del dominio de usuarios (alta y grafo de followers).
    - **Tweets Service**: se encarga de la ingestión de un nuevo tweet, creación del mismo,  notificar actualización de la cache del timeline y leer tweets.
    - **Timeline Service**: se encarga de la generación del timeline.

- Dado que no estamos frente a un caso donde nos afecte tener consistencia eventual, podríamos elegir una base de datos NoSQL, como un clúster de Cassandra o ScyllaDB, para el almacenamiento de los tweets y así poder aprovechar sus bondades para la escalabilidad.

- Podríamos almacenar el timeline en una cache, por ej Redis, para optimizar la respuesta a las consultas del mismo.

Podemos bocetar un diagrama simplifado del diseño final de la siguiente manera:

![final](./diagrams/final.png)

Una explicación sencilla podría ser:

- **Ingestión de un nuevo tweet**
    - Se crea un nuevo tweet en Cassandra.
    - Se publica un nuevo mensaje notificando el evento en Kafka
    - Se actualiza el timeline en Redis para el usuario correspondiente.

- **Timeline**
    - Se busca primero en Redis si existe el timeline correspondiente o si es válido.
    - Si no se encuentra en Redis, se consulta al servicio de Tweets para obtener los datos de Cassandra y actualizar Redis.
    - Si hay un evento de follow/unfollow, se actualiza el timeline.

- **Follow/Unfollow**
    - Se actualiza PostgreSQL.
    - Se notifica el evento por Kafka.

- **Creación de nuevo usuario**
    - Se inserta un nuevo item en PostgreSQL.
    - Se notifica el evento por Kafka para que se creen un timeline en Redis.