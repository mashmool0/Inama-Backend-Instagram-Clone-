"""Async SQLAlchemy setup: the engine, the session factory, and the ORM Base.

Repositories (Step 2) get an AsyncSession from SessionLocal. Nothing here talks
to the database at import time except creating the engine object (lazy — no
connection is opened until first use).
"""

from collections.abc import AsyncIterator

from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import DeclarativeBase

from app.config import load

settings = load()

# echo=False: set True locally to see the SQL SQLAlchemy emits.
engine = create_async_engine(settings.database_url, echo=False, future=True)

# expire_on_commit=False keeps ORM objects usable after commit (handy in async).
SessionLocal = async_sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)


class Base(DeclarativeBase):
    """Base class all ORM models inherit from. Its .metadata is what Alembic
    compares the database against."""


async def get_session() -> AsyncIterator[AsyncSession]:
    """Yield a session scoped to one unit of work. Commit/rollback is the
    caller's responsibility (the service layer decides transaction boundaries)."""
    async with SessionLocal() as session:
        yield session
