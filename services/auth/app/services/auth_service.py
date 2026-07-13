"""AuthService — orchestrates the auth flows.

This is the only place that decides transaction boundaries (it owns the session
and calls commit). It wires the repositories and the security services together;
each of those stays single-purpose. Handlers (Step 4) call these methods and
translate the results + domain errors to gRPC.

Password reset (RequestPasswordReset/ResetPassword) is intentionally NOT here
yet: the auth_db has no email column, so the email-based reset in the proto
needs a design decision first (add email vs. reset via phone/SMS). See Step 4.
"""

from dataclasses import dataclass
from datetime import datetime, timedelta, timezone

from app.clients import EmailClient, SmsClient
from app.config import Settings
from app.errors import (
    AccountNotVerified,
    InvalidCredentials,
    InvalidOTP,
    InvalidToken,
    PhoneAlreadyRegistered,
    UserNotFound,
)
from app.repositories.token_repo import TokenRepository
from app.repositories.user_repo import UserRepository
from app.services.jwt_service import JWTService
from app.services.otp import OTPService
from app.services.password import PasswordService
from app.services.tokens import generate_token, hash_token


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
        passwords: PasswordService,
        jwt: JWTService,
        otp: OTPService,
        sms: SmsClient,
        email: EmailClient,
        settings: Settings,
    ):
        self._session = session
        self._users = users
        self._tokens = tokens
        self._passwords = passwords
        self._jwt = jwt
        self._otp = otp
        self._sms = sms
        self._email = email
        self._settings = settings

    # ---------- flows ----------

    async def register(self, phone: str, password: str) -> None:
        """Create an unverified account and send an OTP. Registration does NOT
        log the user in — they must verify the OTP first."""
        existing = await self._users.get_by_phone(phone)
        if existing is not None and existing.is_verified:
            raise PhoneAlreadyRegistered(phone)

        password_hash = self._passwords.hash(password)
        if existing is None:
            await self._users.create(phone, password_hash)
        else:
            # Re-registering an unverified number: refresh the stored password.
            await self._users.update_password(existing.id, password_hash)

        code = await self._otp.generate(phone)
        await self._sms.send_otp(phone, code)
        await self._session.commit()

    async def verify_otp(self, phone: str, code: str) -> TokenPair:
        """Verify the OTP, mark the account verified, and issue tokens."""
        if not await self._otp.verify(phone, code):
            raise InvalidOTP()
        user = await self._users.get_by_phone(phone)
        if user is None:
            raise UserNotFound()
        await self._users.mark_verified(user.id)
        pair = await self._issue_tokens(str(user.id))
        await self._session.commit()
        return pair

    async def login(self, phone: str, password: str) -> TokenPair:
        user = await self._users.get_by_phone(phone)
        # Same error whether the phone is unknown or the password is wrong —
        # don't reveal which phones exist.
        if user is None or not self._passwords.verify(password, user.password_hash):
            raise InvalidCredentials()
        if not user.is_verified:
            raise AccountNotVerified()
        pair = await self._issue_tokens(str(user.id))
        await self._session.commit()
        return pair

    async def refresh(self, refresh_token: str) -> TokenPair:
        """Rotate the refresh token: verify it, revoke it, issue a fresh pair.
        Rotation means a stolen-and-used token is detectable and short-lived."""
        token_hash = hash_token(refresh_token)
        stored = await self._tokens.get_refresh(token_hash)
        now = datetime.now(timezone.utc)
        if stored is None or stored.revoked or stored.expires_at <= now:
            raise InvalidToken()
        await self._tokens.revoke_refresh(token_hash)
        pair = await self._issue_tokens(str(stored.user_id))
        await self._session.commit()
        return pair

    # ---------- helpers ----------

    async def _issue_tokens(self, user_id: str) -> TokenPair:
        """Issue a short-lived access JWT + a stored (hashed) opaque refresh
        token. Shared by verify_otp, login, and refresh."""
        access = self._jwt.issue_access(user_id)
        raw_refresh = generate_token()
        expires_at = now_plus(self._settings.refresh_token_ttl)
        await self._tokens.add_refresh(user_id, hash_token(raw_refresh), expires_at)
        return TokenPair(
            access_token=access,
            refresh_token=raw_refresh,
            expires_in=self._settings.access_token_ttl,
        )


def now_plus(seconds: int) -> datetime:
    return datetime.now(timezone.utc) + timedelta(seconds=seconds)
