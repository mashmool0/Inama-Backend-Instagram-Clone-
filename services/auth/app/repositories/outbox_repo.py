"""Database access for the transactional outbox.

`add` is called inside the same session/transaction as the state change (so the
event is atomic with it). The relay worker uses get_unpublished + mark_published.
"""

import uuid

from sqlalchemy import func, select, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import OutboxEvent


class OutboxRepository:
    def __init__(self, session: AsyncSession):
        self._session = session

    async def add(self, event_type: str, payload: dict) -> OutboxEvent:
        event = OutboxEvent(event_type=event_type, payload=payload)
        self._session.add(event)
        await self._session.flush()
        return event

    async def get_unpublished(self, limit: int = 100) -> list[OutboxEvent]:
        result = await self._session.scalars(
            select(OutboxEvent)
            .where(OutboxEvent.published_at.is_(None))
            .order_by(OutboxEvent.created_at)
            .limit(limit)
        )
        return list(result)

    async def mark_published(self, event_id: uuid.UUID) -> None:
        await self._session.execute(
            update(OutboxEvent)
            .where(OutboxEvent.id == event_id)
            .values(published_at=func.now())
        )
