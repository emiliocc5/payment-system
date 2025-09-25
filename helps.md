## To connect to DB

`docker exec -it postgres-main bash`

`psql -U user -d paymentsystem`

`\l` --> Listar dbs

`\c paymentsystem` --> Conectarse a la db

`\dt` --> Listar las tablas

`select * from payments limit 100;` --> Listar los pagos


--------------------------------------------------------------------------------

## To send a manual message from kafka

1. Connect to http://localhost:8080/
2. Go to topics
3. Select payment-events
4. Select produce message

Message Key
``
"550e8400-e29b-41d4-a716-446655440000"
``

Message body
```
{ 
    "transaction_id": "2ed244f7-2711-42fb-974b-7ff4e70e7a4c",
    "status": "SUCCESS",
    "metadata": {}
}
```