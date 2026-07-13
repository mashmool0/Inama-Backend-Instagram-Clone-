"""Database access for refresh_tokens and password_reset_tokens.

Both are "temporary, revocable token" tables with the same shape, so one
repository owns them. Only `token_hash` (never the raw token) is stored — the
service layer hashes before calling in. DB access only; no commits here.
"""

import uuid
from datetime import datetime

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import PasswordResetToken, RefreshToken


class TokenRepository:
    def __init__(self, session: AsyncSession):
        self._session = session

    # ---------- refresh tokens ----------

    async def add_refresh(
        self, user_id: uuid.UUID, token_hash: str, expires_at: datetime
    ) -> RefreshToken:
        token = RefreshToken(user_id=user_id, token_hash=token_hash, expires_at=expires_at)
        self._session.add(token)
        await self._session.flush()
        return token

    async def get_refresh(self, token_hash: str) -> RefreshToken | None:
        return await self._session.scalar(
            select(RefreshToken).where(RefreshToken.token_hash == token_hash)
        )

    async def revoke_refresh(self, token_hash: str) -> None:
        """Single-token revoke — used on logout / refresh rotation."""
        await self._session.execute(
            update(RefreshToken)
            .where(RefreshToken.token_hash == token_hash)
            .values(revoked=True)
        )

    async def revoke_all_for_user(self, user_id: uuid.UUID) -> None:
        """Revoke every refresh token for a user — used on password change."""
        await self._session.execute(
            update(RefreshToken)
            .where(RefreshToken.user_id == user_id)
            .values(revoked=True)
        )

    # ---------- password reset tokens ----------

    async def add_reset(
        self, user_id: uuid.UUID, token_hash: str, expires_at: datetime
    ) -> PasswordResetToken:
        token = PasswordResetToken(
            user_id=user_id, token_hash=token_hash, expires_at=expires_at
        )
        self._session.add(token)
        await self._session.flush()
        return token

    async def get_reset(self, token_hash: str) -> PasswordResetToken | None:
        return await self._session.scalar(
            select(PasswordResetToken).where(PasswordResetToken.token_hash == token_hash)
        )

    async def get_reset_for_user(
        self, user_id: uuid.UUID, token_hash: str
    ) -> PasswordResetToken | None:
        """Scope the lookup by user so short reset codes can't collide across
        accounts (the code alone isn't globally unique)."""
        return await self._session.scalar(
            select(PasswordResetToken).where(
                PasswordResetToken.user_id == user_id,
                PasswordResetToken.token_hash == token_hash,
            )
        )

    async def mark_used(self, token_hash: str) -> None:
        """Burn a reset token so it can't be used twice."""
        await self._session.execute(
            update(PasswordResetToken)
            .where(PasswordResetToken.token_hash == token_hash)
            .values(used=True)
        )
