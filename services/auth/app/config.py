"""Auth service configuration, loaded from environment variables.

Twelve-factor: config comes from the environment, never hardcoded. Sensible
local defaults are provided so `alembic upgrade head` works out of the box
against the docker-compose Postgres. Only database_url matters for Step 1;
the JWT/OTP settings are consumed from Step 3 onward.
"""

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    database_url: str
    redis_url: str
    # JWT (RS256) — used from Step 3
    jwt_private_key_path: str
    jwt_public_key_path: str
    access_token_ttl: int   # seconds
    refresh_token_ttl: int  # seconds
    # OTP — used from Step 3
    otp_ttl: int            # seconds
    otp_length: int


def load() -> Settings:
    return Settings(
        database_url=os.environ.get(
            "DATABASE_URL",
            "postgresql+psycopg://inama:inama@localhost:5432/auth_db",
        ),
        redis_url=os.environ.get("REDIS_URL", "redis://localhost:6379/0"),
        jwt_private_key_path=os.environ.get("JWT_PRIVATE_KEY_PATH", "keys/jwt_private.pem"),
        jwt_public_key_path=os.environ.get("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
        access_token_ttl=int(os.environ.get("ACCESS_TOKEN_TTL", "900")),       # 15 min
        refresh_token_ttl=int(os.environ.get("REFRESH_TOKEN_TTL", "604800")),  # 7 days
        otp_ttl=int(os.environ.get("OTP_TTL", "120")),                         # 2 min
        otp_length=int(os.environ.get("OTP_LENGTH", "6")),
    )
