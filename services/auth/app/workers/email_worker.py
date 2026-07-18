"""Fake welcome-email worker.

Consumes user.registered from RabbitMQ and "sends" a welcome email — a small
delay to mimic talking to a real email provider, then a log line. We have no
real email provider; this demonstrates the async producer/consumer flow.

Run:  python -m app.workers.email_worker
"""

import asyncio
import json
import os

import aio_pika

from app.messaging import EXCHANGE_NAME

QUEUE_NAME = "auth.welcome_email"
ROUTING_KEY = "user.registered"
FAKE_SEND_SECONDS = 0.5  # pretend we're calling an email provider


async def run() -> None:
    rabbit_url = os.environ.get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
    connection = await aio_pika.connect_robust(rabbit_url)
    channel = await connection.channel()
    await channel.set_qos(prefetch_count=10)

    exchange = await channel.declare_exchange(
        EXCHANGE_NAME, aio_pika.ExchangeType.TOPIC, durable=True
    )
    queue = await channel.declare_queue(QUEUE_NAME, durable=True)
    await queue.bind(exchange, routing_key=ROUTING_KEY)
    print(f"email worker started, waiting for {ROUTING_KEY}", flush=True)

    async with queue.iterator() as iterator:
        async for message in iterator:
            async with message.process():  # ack on success, requeue on error
                event = json.loads(message.body)
                data = event["data"]
                await asyncio.sleep(FAKE_SEND_SECONDS)
                print(
                    f"📧 welcome email sent to {data['email']} (@{data['username']})",
                    flush=True,
                )


if __name__ == "__main__":
    asyncio.run(run())
