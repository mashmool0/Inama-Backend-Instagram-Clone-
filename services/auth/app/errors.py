"""Domain errors for Auth.

The service layer raises these; the gRPC handler layer maps each to a gRPC
status code. Keeping them separate means business logic never imports gRPC.
"""


class AuthError(Exception):
    """Base class for all Auth domain errors."""


class EmailAlreadyRegistered(AuthError):
    pass


class UsernameAlreadyTaken(AuthError):
    pass


class InvalidUsername(AuthError):
    pass


class InvalidRegistration(AuthError):
    pass


class UserNotFound(AuthError):
    pass


class InvalidCredentials(AuthError):
    pass


class InvalidToken(AuthError):
    pass
