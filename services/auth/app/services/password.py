"""Password hashing with bcrypt.

bcrypt is deliberately SLOW and salts every hash — the right tool for
low-entropy secrets like human passwords, because it makes brute-forcing a
stolen hash impractical.
"""

import bcrypt


class PasswordService:
    def hash(self, password: str) -> str:
        # bcrypt operates on bytes and ignores anything past 72 bytes.
        hashed = bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt())
        return hashed.decode("utf-8")

    def verify(self, password: str, password_hash: str) -> bool:
        return bcrypt.checkpw(
            password.encode("utf-8"), password_hash.encode("utf-8")
        )
