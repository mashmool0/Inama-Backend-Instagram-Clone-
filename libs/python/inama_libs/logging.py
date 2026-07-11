"""Structured JSON logging — same shape as the Go services' logs."""

import json
import logging
import os
import sys


class JsonFormatter(logging.Formatter):
    def __init__(self, service: str):
        super().__init__()
        self.service = service

    def format(self, record: logging.LogRecord) -> str:
        payload = {
            "level": record.levelname.lower(),
            "service": self.service,
            "msg": record.getMessage(),
            "time": self.formatTime(record, "%Y-%m-%dT%H:%M:%S%z"),
        }
        # trace_id is reserved for OpenTelemetry later; attach it via
        # logger.info("...", extra={"trace_id": tid}) when available.
        trace_id = getattr(record, "trace_id", None)
        if trace_id:
            payload["trace_id"] = trace_id
        if record.exc_info:
            payload["error"] = self.formatException(record.exc_info)
        return json.dumps(payload)


def new(service: str) -> logging.Logger:
    """Return a JSON logger for the service. LOG_LEVEL controls verbosity."""
    logger = logging.getLogger(service)
    level = os.getenv("LOG_LEVEL", "info").upper()
    logger.setLevel(getattr(logging, level, logging.INFO))

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JsonFormatter(service))
    logger.handlers = [handler]
    logger.propagate = False
    return logger
