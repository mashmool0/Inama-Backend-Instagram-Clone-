"""Integration test for the async path: outbox row -> relay -> RabbitMQ -> consumer.

Requires RabbitMQ to be reachable on localhost:5672 (skips otherwise).
"""

import json

import aio_pika
import pytest
import pytest_asyncio
from sqlalchemy import select

from app.db import SessionLocal
from app.messaging import EXCHANGE_NAME
from app.models import OutboxEvent
from app.repositories.outbox_repo import OutboxRepository
from app.workers.outbox_relay import publish_pending

RABBIT_URL = "amqp://guest:guest@localhost:5672/"


@pytest_asyncio.fixture
async def rabbit_channel():
    try:
        connection = await aio_pika.connect_robust(RABBIT_URL, timeout=3)
    except Exception:
        pytest.skip("RabbitMQ not reachable on localhost:5672")
    channel = await connection.channel()
    try:
        yield channel
    finally:
        await connection.close()


async def test_relay_publishes_outbox_event_and_marks_it_published(rabbit_channel):
    exchange = await rabbit_channel.declare_exchange(
        EXCHANGE_NAME, aio_pika.ExchangeType.TOPIC, durable=True
    )
    # a private, auto-deleting queue to observe what the relay publishes
    queue = await rabbit_channel.declare_queue("", exclusive=True)
    await queue.bind(exchange, routing_key="user.registered")  # bind BEFORE publishing

    # 1) an unpublished outbox event exists
    async with SessionLocal() as s:
        await OutboxRepository(s).add(
            "user.registered",
            {"user_id": "u1", "email": "r@x.com", "username": "ru"},
        )
        await s.commit()

    # 2) run the relay once
    async with SessionLocal() as s:
        relayed = await publish_pending(s, exchange)
        await s.commit()
    assert relayed == 1

    # 3) the message actually arrived on the broker, with the right payload
    message = await queue.get(timeout=5)
    body = json.loads(message.body)
    await message.ack()
    assert body["event_type"] == "user.registered"
    assert body["data"]["email"] == "r@x.com"
    assert body["data"]["username"] == "ru"

    # 4) the outbox row is now marked published (won't be sent again)
    async with SessionLocal() as s:
        row = (await s.execute(select(OutboxEvent))).scalars().one()
        assert row.published_at is not None
