"""Test fixtures.

Runs against a throwaway `auth_test_db` on the local Postgres (kept separate
from the real auth_db so tests never touch running-system data). Environment is
set up BEFORE importing the app so the engine binds to the test DB.
"""

import asyncio
import os
import tempfile
from pathlib import Path

import psycopg
import pytest_asyncio
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa

TEST_DB = "auth_test_db"
# Override with TEST_PG_BASE="user:pw@host:port" to point at a different Postgres.
_PG_BASE = os.environ.get("TEST_PG_BASE", "inama:inama@localhost:5432")
_ADMIN_DSN = f"postgresql://{_PG_BASE}/postgres"

# 1) Point the app at the test DB, and generate a throwaway RS256 keypair —
#    both must happen before any `app.*` import reads config.
os.environ["DATABASE_URL"] = f"postgresql+psycopg://{_PG_BASE}/{TEST_DB}"

_key_dir = Path(tempfile.mkdtemp())
_private = rsa.generate_private_key(public_exponent=65537, key_size=2048)
(_key_dir / "jwt_private.pem").write_bytes(
    _private.private_bytes(
        serialization.Encoding.PEM,
        serialization.PrivateFormat.PKCS8,
        serialization.NoEncryption(),
    )
)
(_key_dir / "jwt_public.pem").write_bytes(
    _private.public_key().public_bytes(
        serialization.Encoding.PEM, serialization.PublicFormat.SubjectPublicKeyInfo
    )
)
os.environ["JWT_PRIVATE_KEY_PATH"] = str(_key_dir / "jwt_private.pem")
os.environ["JWT_PUBLIC_KEY_PATH"] = str(_key_dir / "jwt_public.pem")


# 2) Create the test database if it doesn't exist (sync, at import time).
def _ensure_test_db() -> None:
    with psycopg.connect(_ADMIN_DSN, autocommit=True) as conn:
        exists = conn.execute(
            "SELECT 1 FROM pg_database WHERE datname = %s", (TEST_DB,)
        ).fetchone()
        if not exists:
            conn.execute(f"CREATE DATABASE {TEST_DB}")


_ensure_test_db()

# 3) Now it's safe to import the app (engine binds to the test DB).
from sqlalchemy import text  # noqa: E402

import app.models  # noqa: E402,F401  (register models on Base.metadata)
from app.db import Base, SessionLocal, engine  # noqa: E402


# 4) Create the schema once, up front.
async def _create_schema() -> None:
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


asyncio.run(_create_schema())


@pytest_asyncio.fixture(autouse=True)
async def _clean_tables():
    """Empty the tables before every test for isolation."""
    async with engine.begin() as conn:
        await conn.execute(
            text("TRUNCATE users, refresh_tokens, outbox RESTART IDENTITY CASCADE")
        )
    yield


@pytest_asyncio.fixture
async def session():
    async with SessionLocal() as s:
        yield s
