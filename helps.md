## To connect to DB

`docker exec -it postgres-main bash`

`psql -U user -d paymentsystem`

`\l` --> Listar dbs

`\c paymentsystem` --> Conectarse a la db

`\dt` --> Listar las tablas

`select * from payments limit 100` --> Listar los pagos


--------------------------------------------------------------------------------

## To send a manual message from kafka

1. Connect to http://localhost:8080/
2. Go to topics
3. Select payment-events
4. Select produce message

Message body
```
{ 
    "transaction_id": "validTransactionID",
    "status": "ACTIVE",
    "metadata": {}
}
```