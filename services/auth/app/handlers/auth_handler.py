"""gRPC handler layer — thin translation between proto and AuthService.

Rules for this layer:
  - No business logic. Unpack the request, call AuthService, pack the response.
  - Open one DB session per call; build the repos + AuthService around it.
  - Map domain errors (app.errors) to gRPC status codes — the ONLY place that
    knows about both worlds.
"""

import grpc
from auth import auth_pb2, auth_pb2_grpc

from app.config import Settings
from app.db import SessionLocal
from app.errors import (
    AuthError,
    EmailAlreadyRegistered,
    InvalidCredentials,
    InvalidToken,
    UsernameAlreadyTaken,
)
from app.repositories.outbox_repo import OutboxRepository
from app.repositories.token_repo import TokenRepository
from app.repositories.user_repo import UserRepository
from app.services.auth_service import AuthService, TokenPair
from app.services.jwt_service import JWTService
from app.services.password import PasswordService

_ERROR_MAP = {
    EmailAlreadyRegistered: (grpc.StatusCode.ALREADY_EXISTS, "email already registered"),
    UsernameAlreadyTaken: (grpc.StatusCode.ALREADY_EXISTS, "username already taken"),
    InvalidCredentials: (grpc.StatusCode.UNAUTHENTICATED, "invalid credentials"),
    InvalidToken: (grpc.StatusCode.UNAUTHENTICATED, "invalid or expired refresh token"),
}


async def _abort(context: grpc.aio.ServicerContext, err: AuthError) -> None:
    code, message = _ERROR_MAP.get(type(err), (grpc.StatusCode.INTERNAL, "internal error"))
    await context.abort(code, message)


def _to_proto(pair: TokenPair) -> "auth_pb2.TokenPair":
    return auth_pb2.TokenPair(
        access_token=pair.access_token,
        refresh_token=pair.refresh_token,
        expires_in=pair.expires_in,
    )


class AuthHandler(auth_pb2_grpc.AuthServiceServicer):
    """Stateless-ish deps (settings, jwt, passwords) are built once at startup;
    a fresh DB session + AuthService is built per request."""

    def __init__(self, settings: Settings, jwt: JWTService, passwords: PasswordService):
        self._settings = settings
        self._jwt = jwt
        self._passwords = passwords

    def _service(self, session) -> AuthService:
        return AuthService(
            session=session,
            users=UserRepository(session),
            tokens=TokenRepository(session),
            outbox=OutboxRepository(session),
            passwords=self._passwords,
            jwt=self._jwt,
            settings=self._settings,
        )

    async def Register(self, request, context):
        async with SessionLocal() as session:
            try:
                pair = await self._service(session).register(
                    request.email, request.username, request.password
                )
            except AuthError as err:
                await _abort(context, err)
        return _to_proto(pair)

    async def Login(self, request, context):
        async with SessionLocal() as session:
            try:
                pair = await self._service(session).login(
                    request.identifier, request.password
                )
            except AuthError as err:
                await _abort(context, err)
        return _to_proto(pair)

    async def RefreshToken(self, request, context):
        async with SessionLocal() as session:
            try:
                pair = await self._service(session).refresh(request.refresh_token)
            except AuthError as err:
                await _abort(context, err)
        return _to_proto(pair)
