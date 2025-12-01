import pika
import json
import os
from datetime import datetime
import uuid
AMQP_URL = os.getenv(
    "AMQP_URL",
    "amqp://user:pass@host.docker.internal:5672/"
)

ERROR_EXCHANGE = "error.events"
ERROR_QUEUE = "error.handler"

QUEUE_NAME = os.getenv("QUEUE_NAME", "my_queue")

def publish_error(original_message, exc):
    params = pika.URLParameters(AMQP_URL)
    connection = pika.BlockingConnection(params)
    channel = connection.channel()

    # garante exchange e fila de erro
    channel.exchange_declare(exchange=ERROR_EXCHANGE, exchange_type="topic", durable=True)
    channel.queue_declare(queue=ERROR_QUEUE, durable=True)
    channel.queue_bind(queue=ERROR_QUEUE, exchange=ERROR_EXCHANGE, routing_key="error.*")

    # 🔥 Routing Key dinâmico
    rk_suffix = original_message.get("deviceType") or original_message.get("deviceId") or "unknown"
    routing_key = f"error.{rk_suffix}"

    error_payload = {
        "errorId": str(uuid.uuid4()),
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "deviceId": original_message.get("deviceId"),
        "deviceType": original_message.get("deviceType"),
        "context": {
            "operation": "publish",
            "route": "/publish",
        },
        "error": {
            "message": str(exc),
            "type": exc.__class__.__name__,
            "payload": original_message
        }
    }

    channel.basic_publish(
        exchange=ERROR_EXCHANGE,
        routing_key=routing_key,   # ← agora dinâmico
        body=json.dumps(error_payload),
        properties=pika.BasicProperties(delivery_mode=2)
    )

    connection.close()

def publish_message(message: dict):
    try:
        params = pika.URLParameters(AMQP_URL)
        connection = pika.BlockingConnection(params)
        channel = connection.channel()

        channel.queue_declare(queue=QUEUE_NAME, durable=True)

        channel.basic_publish(
            exchange="",
            routing_key=QUEUE_NAME,
            body=json.dumps(message),
            properties=pika.BasicProperties(delivery_mode=2)
        )

        connection.close()

    except Exception as e:
        print("❌ Erro ao publicar no RabbitMQ:", e)
        publish_error(message, e)
        raise e