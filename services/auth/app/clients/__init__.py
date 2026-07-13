"""Outbound client contracts (SMS, email).

These are Protocols — the interface AuthService depends on. The concrete mock
implementations (dev) land in Step 6. Depending on the interface keeps the
business logic testable (inject a fake in tests).
"""

from typing import Protocol


class SmsClient(Protocol):
    async def send_otp(self, phone: str, code: str) -> None: ...


class EmailClient(Protocol):
    async def send_reset_link(self, email: str, token: str) -> None: ...
