"""Database access for the users table.

Rules for this layer:
  - DB access ONLY — no password hashing, no token logic, no business rules.
  - No commits. Repositories `flush` when they need a generated id, but the
    SERVICE layer decides when to commit/rollback (transaction boundaries).
"""

import uuid

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import User


class UserRepository:
    def __init__(self, session: AsyncSession):
        self._session = session

    async def create(self, phone: str, password_hash: str) -> User:
        """Insert a new (unverified) user. Flush to populate the generated id
        and defaults, but do not commit."""
        user = User(phone=phone, password_hash=password_hash)
        self._session.add(user)
        await self._session.flush()
        return user

    async def get_by_phone(self, phone: str) -> User | None:
        """Used by login and by the register flow's duplicate check."""
        return await self._session.scalar(select(User).where(User.phone == phone))

    async def get_by_id(self, user_id: uuid.UUID) -> User | None:
        return await self._session.get(User, user_id)

    async def mark_verified(self, user_id: uuid.UUID) -> None:
        """Flip is_verified to true after a successful OTP verification."""
        await self._session.execute(
            update(User).where(User.id == user_id).values(is_verified=True)
        )

    async def update_password(self, user_id: uuid.UUID, password_hash: str) -> None:
        """Set a new password hash (used by the reset-password flow)."""
        await self._session.execute(
            update(User).where(User.id == user_id).values(password_hash=password_hash)
        )
