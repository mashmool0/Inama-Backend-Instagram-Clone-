"""Helpers for opaque tokens (refresh tokens, password-reset tokens).

Unlike passwords, these tokens are HIGH-entropy random strings, so they don't
need bcrypt's slowness — a plain SHA-256 is enough to store them safely. We hand
the raw token to the client once and keep only its hash in the database, so a DB
leak can't be replayed.
"""

import hashlib
import secrets


def generate_token(nbytes: int = 32) -> str:
    """A cryptographically-strong, URL-safe random token."""
    return secrets.token_urlsafe(nbytes)


def generate_numeric_code(length: int = 6) -> str:
    """A short numeric code suitable for SMS (e.g. password-reset code)."""
    return "".join(secrets.choice("0123456789") for _ in range(length))


def hash_token(token: str) -> str:
    """SHA-256 hex digest — what we store in the DB (never the raw token)."""
    return hashlib.sha256(token.encode("utf-8")).hexdigest()
