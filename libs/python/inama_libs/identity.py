"""The x-user-id gRPC metadata convention for Python services (auth, search).

Same idea as libs/go/identity: the Gateway verifies the JWT and sends the user
id in the ``x-user-id`` metadata; the service reads it from the call context.
Python's grpc passes a ``context`` object into every servicer method, so plain
helpers are cleaner here than an interceptor.
"""

import grpc

USER_ID_KEY = "x-user-id"


def user_id_from_context(context: grpc.ServicerContext) -> str | None:
    """Return the authenticated user id from the call metadata, or None."""
    for key, value in context.invocation_metadata():
        if key == USER_ID_KEY and value:
            return value
    return None


def require_user_id(context: grpc.ServicerContext) -> str:
    """Return the user id or abort the call with UNAUTHENTICATED."""
    user_id = user_id_from_context(context)
    if not user_id:
        context.abort(grpc.StatusCode.UNAUTHENTICATED, "missing authenticated user")
    return user_id


def with_user_id(user_id: str) -> list[tuple[str, str]]:
    """Metadata to attach when THIS service calls another as a given user."""
    return [(USER_ID_KEY, user_id)]
