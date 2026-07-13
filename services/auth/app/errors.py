"""Domain errors for Auth.

The service layer raises these; the gRPC handler layer (Step 4) maps each to a
gRPC status code. Keeping them separate means business logic never imports gRPC.
"""


class AuthError(Exception):
    """Base class for all Auth domain errors."""


class PhoneAlreadyRegistered(AuthError):
    pass


class InvalidOTP(AuthError):
    pass


class InvalidCredentials(AuthError):
    pass


class AccountNotVerified(AuthError):
    pass


class InvalidToken(AuthError):
    pass


class UserNotFound(AuthError):
    pass
