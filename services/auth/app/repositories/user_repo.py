"""Database access for the users table.

DB access ONLY — no password hashing, no business rules, no commits. The
service layer owns transaction boundaries.
"""

import uuid

from sqlalchemy import or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import User


class UserRepository:
    def __init__(self, session: AsyncSession):
        self._session = session

    async def create(self, email: str, username: str, password_hash: str) -> User:
        user = User(email=email, username=username, password_hash=password_hash)
        self._session.add(user)
        await self._session.flush()  # populate id, without committing
        return user

    async def get_by_email(self, email: str) -> User | None:
        return await self._session.scalar(select(User).where(User.email == email))

    async def get_by_username(self, username: str) -> User | None:
        return await self._session.scalar(select(User).where(User.username == username))

    async def get_by_identifier(self, identifier: str) -> User | None:
        """Resolve a login identifier that may be either an email or a username."""
        return await self._session.scalar(
            select(User).where(or_(User.email == identifier, User.username == identifier))
        )

    async def get_by_id(self, user_id: uuid.UUID) -> User | None:
        return await self._session.get(User, user_id)

    async def update_username(self, user: User, username: str) -> User:
        user.username = username
        await self._session.flush()
        return user
