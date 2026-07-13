"""Outbound client contracts (SMS).

A Protocol — the interface AuthService depends on. The concrete mock
implementation (dev) lands in Step 6. Depending on the interface keeps the
business logic testable (inject a fake in tests). Everything is phone/SMS
based, so there is no email client.
"""

from typing import Protocol


class SmsClient(Protocol):
    async def send_otp(self, phone: str, code: str) -> None: ...

    async def send_reset_code(self, phone: str, code: str) -> None: ...
