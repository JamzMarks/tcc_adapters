import pika
import json
import os

AMQP_URL = os.getenv(
    "AMQP_URL",
    "amqp://user:pass@host.docker.internal:5672/"
)

QUEUE_NAME = os.getenv("QUEUE_NAME", "my_queue")

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
        raise e
