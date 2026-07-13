"""Alembic environment — async flavour.

Migrations run through an async engine (matching the service's async stack).
The DB URL and target metadata come from the app itself, so there is a single
source of truth.
"""

import asyncio
from logging.config import fileConfig

from alembic import context
from sqlalchemy import pool
from sqlalchemy.ext.asyncio import async_engine_from_config

from app.config import load
from app.db import Base
import app.models  # noqa: F401 — import so models register on Base.metadata

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# Feed the real DB URL in from app config (env var), not from alembic.ini.
config.set_main_option("sqlalchemy.url", load().database_url)

target_metadata = Base.metadata


def run_migrations_offline() -> None:
    """Generate SQL without a DB connection (alembic upgrade --sql)."""
    context.configure(
        url=load().database_url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )
    with context.begin_transaction():
        context.run_migrations()


def do_run_migrations(connection) -> None:
    context.configure(connection=connection, target_metadata=target_metadata)
    with context.begin_transaction():
        context.run_migrations()


async def run_migrations_online() -> None:
    """Open an async connection and apply migrations."""
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )
    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)
    await connectable.dispose()


if context.is_offline_mode():
    run_migrations_offline()
else:
    asyncio.run(run_migrations_online())
