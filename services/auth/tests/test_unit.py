"""Unit tests for the pure security building blocks (no DB)."""

import jwt as pyjwt
import pytest

from app.config import load
from app.services.jwt_service import JWTService
from app.services.password import PasswordService
from app.services.tokens import generate_token, hash_token


# ---------- password ----------

def test_password_hash_is_not_plaintext_and_verifies():
    svc = PasswordService()
    hashed = svc.hash("s3cret-pw")
    assert hashed != "s3cret-pw"          # never stored in the clear
    assert svc.verify("s3cret-pw", hashed) is True
    assert svc.verify("wrong", hashed) is False


def test_password_hash_is_salted():
    svc = PasswordService()
    # same password → different hashes (random salt)
    assert svc.hash("same") != svc.hash("same")


# ---------- tokens ----------

def test_generate_token_is_random():
    assert generate_token() != generate_token()


def test_hash_token_is_deterministic_and_sized():
    h1 = hash_token("abc")
    assert h1 == hash_token("abc")        # deterministic
    assert h1 != "abc"                    # actually hashed
    assert len(h1) == 64                  # sha-256 hex


# ---------- jwt (RS256) ----------

def test_jwt_issue_and_verify_roundtrip():
    svc = JWTService(load())
    token = svc.issue_access("user-123")
    claims = svc.verify_access(token)
    assert claims["sub"] == "user-123"
    assert claims["type"] == "access"
    assert claims["exp"] > claims["iat"]


def test_jwt_tampered_token_is_rejected():
    svc = JWTService(load())
    token = svc.issue_access("user-123")
    tampered = token[:-3] + ("aaa" if not token.endswith("aaa") else "bbb")
    with pytest.raises(pyjwt.PyJWTError):
        svc.verify_access(tampered)
