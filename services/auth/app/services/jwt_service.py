"""Access-token issuing with RS256 (asymmetric).

Auth signs access tokens with the PRIVATE key. The API Gateway verifies them
with the PUBLIC key — so the Gateway can check a token but can never forge one.
Access tokens are short-lived and stateless (not stored); the refresh token
(opaque, stored hashed) is what gives revocation.
"""

import time
from pathlib import Path

import jwt

from app.config import Settings


class JWTService:
    ALGORITHM = "RS256"

    def __init__(self, settings: Settings):
        # Keys are PEM files on disk; paths come from config. Generate them for
        # local dev with scripts/generate_keys.py.
        self._private_key = Path(settings.jwt_private_key_path).read_text()
        self._public_key = Path(settings.jwt_public_key_path).read_text()
        self._access_ttl = settings.access_token_ttl

    def issue_access(self, user_id: str) -> str:
        now = int(time.time())
        payload = {
            "sub": str(user_id),             # subject = the authenticated user
            "iat": now,                      # issued-at
            "exp": now + self._access_ttl,   # expiry
            "type": "access",
        }
        return jwt.encode(payload, self._private_key, algorithm=self.ALGORITHM)

    def verify_access(self, token: str) -> dict:
        """Verify signature + expiry and return the claims. The Gateway does the
        real per-request verification in Go; this mirror is for Auth's own tests.
        Raises jwt.PyJWTError on any invalid/expired token."""
        return jwt.decode(token, self._public_key, algorithms=[self.ALGORITHM])
