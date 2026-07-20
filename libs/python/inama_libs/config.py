"""Environment-variable configuration helpers — mirror of libs/go/config."""

import os
from dataclasses import dataclass


def get(key: str, default: str = "") -> str:
    return os.environ.get(key, default)


def must_get(key: str) -> str:
    """Return the env var or raise — fail fast at startup for values with no
    safe default (e.g. a database DSN)."""
    value = os.environ.get(key)
    if not value:
        raise RuntimeError(f"required env var {key!r} is not set")
    return value


def get_int(key: str, default: int) -> int:
    try:
        return int(os.environ[key])
    except (KeyError, ValueError):
        return default


def get_bool(key: str, default: bool) -> bool:
    value = os.environ.get(key)
    if value is None:
        return default
    return value.strip().lower() in ("1", "true", "yes", "on")


@dataclass
class Base:
    service_name: str
    grpc_port: int
    metrics_port: int
    log_level: str


def load_base(service: str) -> Base:
    return Base(
        service_name=service,
        grpc_port=get_int("GRPC_PORT", 50051),
        metrics_port=get_int("METRICS_PORT", 8000),
        log_level=get("LOG_LEVEL", "info"),
    )
