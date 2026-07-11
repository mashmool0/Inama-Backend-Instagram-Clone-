"""Shared conventions for the Python (FastAPI) services: auth and search.

Mirror of libs/go so both languages behave the same:
  - logging  : structured JSON logs
  - config   : env-var configuration helpers
  - identity : the x-user-id gRPC metadata convention
"""
