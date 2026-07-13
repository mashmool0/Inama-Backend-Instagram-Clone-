import asyncio
import uuid
from datetime import datetime, timedelta, timezone

from app.db import SessionLocal, engine
from app.repositories.user_repo import UserRepository
from app.repositories.token_repo import TokenRepository


async def main():
    phone = "0912" + uuid.uuid4().hex[:7]

    # ۱. ساخت کاربر (یه session)
    async with SessionLocal() as s:
        user = await UserRepository(s).create(phone=phone, password_hash="fake_hash")
        await s.commit()
        user_id = user.id                    # id رو همین‌جا نگه دار
        print("✓ created:", user_id, phone, "| verified:", user.is_verified)

    # ۲. session تازه: پیدا کردن + تأیید
    async with SessionLocal() as s:
        users = UserRepository(s)
        found = await users.get_by_phone(phone)
        print("✓ get_by_phone matches:", found.id == user_id)
        await users.mark_verified(user_id)
        await s.commit()

    # ۳. session تازه: مطمئن شو verified شده
    async with SessionLocal() as s:
        refreshed = await UserRepository(s).get_by_id(user_id)
        print("✓ is_verified now:", refreshed.is_verified)

    # ۴. توکن: افزودن و باطل کردن
    async with SessionLocal() as s:
        tokens = TokenRepository(s)
        await tokens.add_refresh(
            user_id=user_id,
            token_hash="rt_hash_1",
            expires_at=datetime.now(timezone.utc) + timedelta(days=7),
        )
        await s.commit()
        await tokens.revoke_refresh("rt_hash_1")
        await s.commit()

    # ۵. session تازه: مطمئن شو باطل شده
    async with SessionLocal() as s:
        got = await TokenRepository(s).get_refresh("rt_hash_1")
        print("✓ refresh revoked:", got.revoked)

    await engine.dispose()
    print("\n🎉 همه چیز کار کرد")


asyncio.run(main())
