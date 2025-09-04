#!/bin/sh
set -e

echo "Esperando a que RabbitMQ esté listo en el puerto 5672..."
until rabbitmqctl status > /dev/null 2>&1; do
  sleep 2
done

echo "Esperando a que la interfaz de gestión de RabbitMQ esté lista..."
until curl -s -u admin:admin123 http://rabbitmq:15672/api/aliveness-test/%2F > /dev/null; do
  sleep 2
done

echo "Configurando permisos para el usuario admin..."
rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"

if [ ! -f /usr/local/bin/rabbitmqadmin ]; then
  echo "Descargando rabbitmqadmin..."
  curl -s -o /usr/local/bin/rabbitmqadmin http://rabbitmq:15672/cli/rabbitmqadmin
  chmod +x /usr/local/bin/rabbitmqadmin
fi

echo "Declarando exchange..."
rabbitmqadmin -u admin -p admin123 -V / declare exchange name=payments_exchange type=direct durable=true 2> /tmp/rabbitmq_error.log
if [ $? -ne 0 ]; then
  echo "Error al declarar el exchange. Verifica /tmp/rabbitmq_error.log"
  cat /tmp/rabbitmq_error.log
  exit 1
fi

echo "Declarando queue..."
rabbitmqadmin -u admin -p admin123 -V / declare queue name=payment_initiated durable=true 2> /tmp/rabbitmq_error.log
if [ $? -ne 0 ]; then
  echo "Error al declarar la queue. Verifica /tmp/rabbitmq_error.log"
  cat /tmp/rabbitmq_error.log
  exit 1
fi

echo "Haciendo binding..."
rabbitmqadmin -u admin -p admin123 -V / declare binding source=payments_exchange destination=payment_initiated routing_key=payment_initiated 2> /tmp/rabbitmq_error.log
if [ $? -ne 0 ]; then
  echo "Error al declarar el binding. Verifica /tmp/rabbitmq_error.log"
  cat /tmp/rabbitmq_error.log
  exit 1
fi

echo "Verificando que la queue existe..."
rabbitmqadmin -u admin -p admin123 -V / list queues | grep payment_initiated

echo "Inicialización completada ✅"