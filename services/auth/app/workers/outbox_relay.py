"""Outbox relay worker.

Polls the auth_db outbox table for unpublished events and publishes them to
RabbitMQ, then stamps them published. This is the ONLY thing that talks to
RabbitMQ on the write side — the Auth service itself only writes to the outbox
table, so a broker outage never blocks or breaks signup.

Run:  python -m app.workers.outbox_relay
"""

import asyncio
import json
import os

import aio_pika

from app.db import SessionLocal
from app.messaging import EXCHANGE_NAME
from app.repositories.outbox_repo import OutboxRepository

POLL_INTERVAL_SECONDS = 1.0


async def run() -> None:
    rabbit_url = os.environ.get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
    connection = await aio_pika.connect_robust(rabbit_url)
    channel = await connection.channel()
    exchange = await channel.declare_exchange(
        EXCHANGE_NAME, aio_pika.ExchangeType.TOPIC, durable=True
    )
    print("outbox relay started", flush=True)

    while True:
        try:
            async with SessionLocal() as session:
                repo = OutboxRepository(session)
                events = await repo.get_unpublished(limit=100)
                for event in events:
                    body = json.dumps(
                        {"event_type": event.event_type, "data": event.payload}
                    ).encode()
                    await exchange.publish(
                        aio_pika.Message(
                            body=body,
                            delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                        ),
                        routing_key=event.event_type,
                    )
                    await repo.mark_published(event.id)
                    print(f"relayed {event.event_type} ({event.id})", flush=True)
                await session.commit()
        except Exception as exc:  # keep the worker alive across transient errors
            print(f"relay error (will retry): {exc}", flush=True)
        await asyncio.sleep(POLL_INTERVAL_SECONDS)


if __name__ == "__main__":
    asyncio.run(run())
