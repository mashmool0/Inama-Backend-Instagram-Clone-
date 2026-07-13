"""One-time passwords (OTP) for phone verification.

Codes live in Redis with a TTL, so they expire automatically — no cleanup job,
no DB table. A code is single-use: it's deleted on the first successful check.
"""

import secrets

from redis.asyncio import Redis

from app.config import Settings


class OTPService:
    def __init__(self, redis: Redis, settings: Settings):
        self._redis = redis
        self._ttl = settings.otp_ttl
        self._length = settings.otp_length

    def _key(self, phone: str) -> str:
        return f"otp:{phone}"

    async def generate(self, phone: str) -> str:
        """Create a numeric code, store it with a TTL, and return it (the caller
        sends it via SMS)."""
        code = "".join(secrets.choice("0123456789") for _ in range(self._length))
        await self._redis.set(self._key(phone), code, ex=self._ttl)
        return code

    async def verify(self, phone: str, code: str) -> bool:
        stored = await self._redis.get(self._key(phone))
        if stored is None:
            return False  # never issued, or already expired
        if isinstance(stored, bytes):
            stored = stored.decode("utf-8")
        # constant-time compare to avoid leaking info via timing
        if secrets.compare_digest(stored, code):
            await self._redis.delete(self._key(phone))  # one-time use
            return True
        return False
