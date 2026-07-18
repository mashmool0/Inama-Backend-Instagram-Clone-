"""AuthService — orchestrates the auth flows.

The only place that owns transaction boundaries (calls commit). It wires the
repositories and security services together; each of those stays single-purpose.
Handlers translate the results + domain errors to gRPC.

Email + username based, password login, no verification step. Signup logs the
user in immediately and (later, via the Outbox) emits user.registered so the
User service can create a profile.
"""

from dataclasses import dataclass
from datetime import datetime, timedelta, timezone

from app.config import Settings
from app.errors import (
    EmailAlreadyRegistered,
    InvalidCredentials,
    InvalidToken,
    UsernameAlreadyTaken,
)
from app.repositories.outbox_repo import OutboxRepository
from app.repositories.token_repo import TokenRepository
from app.repositories.user_repo import UserRepository
from app.services.jwt_service import JWTService
from app.services.password import PasswordService
from app.services.tokens import generate_token, hash_token

USER_REGISTERED = "user.registered"


@dataclass
class TokenPair:
    access_token: str
    refresh_token: str
    expires_in: int  # access-token lifetime, seconds


class AuthService:
    def __init__(
        self,
        session,
        users: UserRepository,
        tokens: TokenRepository,
        outbox: OutboxRepository,
        passwords: PasswordService,
        jwt: JWTService,
        settings: Settings,
    ):
        self._session = session
        self._users = users
        self._tokens = tokens
        self._outbox = outbox
        self._passwords = passwords
        self._jwt = jwt
        self._settings = settings

    async def register(self, email: str, username: str, password: str) -> TokenPair:
        """Create the account and log the user in. Uniqueness of email and
        username is checked up front (the DB constraints are the final guard)."""
        if await self._users.get_by_email(email) is not None:
            raise EmailAlreadyRegistered()
        if await self._users.get_by_username(username) is not None:
            raise UsernameAlreadyTaken()

        user = await self._users.create(email, username, self._passwords.hash(password))
        # Written in the SAME transaction as the user — atomic. A relay worker
        # publishes it to RabbitMQ; a consumer "sends" the welcome email, and
        # (later) the User service creates the profile.
        await self._outbox.add(
            USER_REGISTERED,
            {"user_id": str(user.id), "email": email, "username": username},
        )
        pair = await self._issue_tokens(str(user.id))
        await self._session.commit()
        return pair

    async def login(self, identifier: str, password: str) -> TokenPair:
        """Log in by username OR email + password."""
        user = await self._users.get_by_identifier(identifier)
        # Same error whether the identifier is unknown or the password is wrong.
        if user is None or not self._passwords.verify(password, user.password_hash):
            raise InvalidCredentials()
        pair = await self._issue_tokens(str(user.id))
        await self._session.commit()
        return pair

    async def refresh(self, refresh_token: str) -> TokenPair:
        """Rotate the refresh token: verify it, revoke it, issue a fresh pair."""
        token_hash = hash_token(refresh_token)
        stored = await self._tokens.get_refresh(token_hash)
        now = datetime.now(timezone.utc)
        if stored is None or stored.revoked or stored.expires_at <= now:
            raise InvalidToken()
        await self._tokens.revoke_refresh(token_hash)
        pair = await self._issue_tokens(str(stored.user_id))
        await self._session.commit()
        return pair

    async def _issue_tokens(self, user_id: str) -> TokenPair:
        """Issue a short-lived access JWT + a stored (hashed) opaque refresh token."""
        access = self._jwt.issue_access(user_id)
        raw_refresh = generate_token()
        expires_at = datetime.now(timezone.utc) + timedelta(
            seconds=self._settings.refresh_token_ttl
        )
        await self._tokens.add_refresh(user_id, hash_token(raw_refresh), expires_at)
        return TokenPair(
            access_token=access,
            refresh_token=raw_refresh,
            expires_in=self._settings.access_token_ttl,
        )
