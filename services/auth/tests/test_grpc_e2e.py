"""End-to-end gRPC tests: start the real AuthHandler on an in-process server and
call it through a generated client stub. Proves the full path — proto
serialization, the handler, AuthService, the DB, and error→status mapping.
"""

import grpc
import pytest
import pytest_asyncio
from auth import auth_pb2, auth_pb2_grpc

from app.config import load
from app.handlers.auth_handler import AuthHandler
from app.services.jwt_service import JWTService
from app.services.password import PasswordService


@pytest_asyncio.fixture
async def stub():
    settings = load()
    server = grpc.aio.server()
    auth_pb2_grpc.add_AuthServiceServicer_to_server(
        AuthHandler(settings, JWTService(settings), PasswordService()), server
    )
    port = server.add_insecure_port("127.0.0.1:0")
    await server.start()
    channel = grpc.aio.insecure_channel(f"127.0.0.1:{port}")
    try:
        yield auth_pb2_grpc.AuthServiceStub(channel)
    finally:
        await channel.close()
        await server.stop(grace=None)


async def test_register_then_login_then_refresh(stub):
    reg = await stub.Register(
        auth_pb2.RegisterRequest(email="g@x.com", username="ge", password="pw123")
    )
    assert reg.access_token and reg.refresh_token and reg.expires_in > 0

    login = await stub.Login(auth_pb2.LoginRequest(identifier="ge", password="pw123"))
    assert login.access_token

    refreshed = await stub.RefreshToken(
        auth_pb2.RefreshTokenRequest(refresh_token=reg.refresh_token)
    )
    assert refreshed.refresh_token != reg.refresh_token  # rotated


async def test_duplicate_email_maps_to_already_exists(stub):
    await stub.Register(auth_pb2.RegisterRequest(email="d@x.com", username="d1", password="pw"))
    with pytest.raises(grpc.aio.AioRpcError) as err:
        await stub.Register(auth_pb2.RegisterRequest(email="d@x.com", username="d2", password="pw"))
    assert err.value.code() == grpc.StatusCode.ALREADY_EXISTS


async def test_wrong_password_maps_to_unauthenticated(stub):
    await stub.Register(auth_pb2.RegisterRequest(email="w@x.com", username="w1", password="right"))
    with pytest.raises(grpc.aio.AioRpcError) as err:
        await stub.Login(auth_pb2.LoginRequest(identifier="w@x.com", password="wrong"))
    assert err.value.code() == grpc.StatusCode.UNAUTHENTICATED
