"""Generate an RS256 keypair for local development.

    python scripts/generate_keys.py

Writes keys/jwt_private.pem and keys/jwt_public.pem (both gitignored — *.pem is
ignored repo-wide). In production the keys come from a secret store / mounted
files, never from the repo. Uses `cryptography`, which ships with pyjwt[crypto].
"""

from pathlib import Path

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa


def main() -> None:
    keys_dir = Path("keys")
    keys_dir.mkdir(exist_ok=True)

    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)

    private_pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )
    public_pem = private_key.public_key().public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo,
    )

    (keys_dir / "jwt_private.pem").write_bytes(private_pem)
    (keys_dir / "jwt_public.pem").write_bytes(public_pem)
    print("wrote keys/jwt_private.pem and keys/jwt_public.pem")


if __name__ == "__main__":
    main()
