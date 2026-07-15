"""Auth service entrypoint — an async gRPC server.

Builds the startup-scoped dependencies (config, JWT keys, password hasher),
registers the AuthService handler, and serves. The Gateway is the only caller.
"""

import asyncio
import os

import grpc
from auth import auth_pb2_grpc

from app.config import load
from app.handlers.auth_handler import AuthHandler
from app.services.jwt_service import JWTService
from app.services.password import PasswordService


async def serve() -> None:
    settings = load()
    handler = AuthHandler(
        settings=settings,
        jwt=JWTService(settings),          # loads the RS256 keys at startup
        passwords=PasswordService(),
    )

    server = grpc.aio.server()
    auth_pb2_grpc.add_AuthServiceServicer_to_server(handler, server)

    port = os.environ.get("GRPC_PORT", "50051")
    server.add_insecure_port(f"0.0.0.0:{port}")
    await server.start()
    print(f"auth grpc server listening on :{port}", flush=True)
    await server.wait_for_termination()


if __name__ == "__main__":
    asyncio.run(serve())
