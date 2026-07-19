"""Integration tests for AuthService against a real Postgres (auth_test_db).

Each call opens its own session — exactly like the gRPC handler does per request
(one session per request). That also matches production behaviour where a
revoked token is re-read fresh from the DB, not from a stale in-memory object.
"""

import pytest
from sqlalchemy import select

from app.config import load
from app.db import SessionLocal
from app.errors import (
    EmailAlreadyRegistered,
    InvalidCredentials,
    InvalidToken,
    UsernameAlreadyTaken,
)
from app.models import OutboxEvent
from app.repositories.outbox_repo import OutboxRepository
from app.repositories.token_repo import TokenRepository
from app.repositories.user_repo import UserRepository
from app.services.auth_service import AuthService
from app.services.jwt_service import JWTService
from app.services.password import PasswordService


def _service(session) -> AuthService:
    settings = load()
    return AuthService(
        session=session,
        users=UserRepository(session),
        tokens=TokenRepository(session),
        outbox=OutboxRepository(session),
        passwords=PasswordService(),
        jwt=JWTService(settings),
        settings=settings,
    )


async def _register(email, username, password):
    async with SessionLocal() as s:
        return await _service(s).register(email, username, password)


async def _login(identifier, password):
    async with SessionLocal() as s:
        return await _service(s).login(identifier, password)


async def _refresh(token):
    async with SessionLocal() as s:
        return await _service(s).refresh(token)


# ---------- register ----------

async def test_register_creates_user_and_outbox_event():
    pair = await _register("ali@x.com", "ali", "pw12345")
    assert pair.access_token and pair.refresh_token and pair.expires_in > 0

    async with SessionLocal() as s:
        user = await UserRepository(s).get_by_email("ali@x.com")
        assert user is not None
        assert user.username == "ali"
        assert user.password_hash != "pw12345"  # stored hashed

        events = (await s.execute(select(OutboxEvent))).scalars().all()
        assert len(events) == 1
        assert events[0].event_type == "user.registered"
        assert events[0].payload["email"] == "ali@x.com"
        assert events[0].payload["username"] == "ali"
        assert events[0].published_at is None  # not relayed yet


async def test_register_duplicate_email_rejected():
    await _register("dup@x.com", "user1", "pw")
    with pytest.raises(EmailAlreadyRegistered):
        await _register("dup@x.com", "user2", "pw")


async def test_register_duplicate_username_rejected():
    await _register("a@x.com", "sameuser", "pw")
    with pytest.raises(UsernameAlreadyTaken):
        await _register("b@x.com", "sameuser", "pw")


# ---------- login ----------

async def test_login_by_email_and_by_username():
    await _register("ali@x.com", "ali", "pw12345")
    assert (await _login("ali@x.com", "pw12345")).access_token
    assert (await _login("ali", "pw12345")).access_token


async def test_login_wrong_password_rejected():
    await _register("ali@x.com", "ali", "pw12345")
    with pytest.raises(InvalidCredentials):
        await _login("ali@x.com", "wrong")


async def test_login_unknown_identifier_rejected():
    with pytest.raises(InvalidCredentials):
        await _login("nobody@x.com", "pw")


# ---------- refresh ----------

async def test_refresh_rotates_and_revokes_old_token():
    pair = await _register("ali@x.com", "ali", "pw12345")
    rotated = await _refresh(pair.refresh_token)
    assert rotated.refresh_token != pair.refresh_token  # rotation

    # the old token is now revoked and can't be reused
    with pytest.raises(InvalidToken):
        await _refresh(pair.refresh_token)


async def test_refresh_invalid_token_rejected():
    with pytest.raises(InvalidToken):
        await _refresh("not-a-real-token")
