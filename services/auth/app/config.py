"""Auth service configuration, loaded from environment variables.

Twelve-factor: config comes from the environment, never hardcoded. Sensible
local defaults so `alembic upgrade head` works out of the box against the
docker-compose Postgres.
"""

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    database_url: str
    # JWT (RS256)
    jwt_private_key_path: str
    jwt_public_key_path: str
    access_token_ttl: int   # seconds
    refresh_token_ttl: int  # seconds


def load() -> Settings:
    return Settings(
        database_url=os.environ.get(
            "DATABASE_URL",
            "postgresql+psycopg://inama:inama@localhost:5432/auth_db",
        ),
        jwt_private_key_path=os.environ.get("JWT_PRIVATE_KEY_PATH", "keys/jwt_private.pem"),
        jwt_public_key_path=os.environ.get("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
        access_token_ttl=int(os.environ.get("ACCESS_TOKEN_TTL", "900")),       # 15 min
        refresh_token_ttl=int(os.environ.get("REFRESH_TOKEN_TTL", "604800")),  # 7 days
    )
