from fastapi import FastAPI

# Phase 0 scaffold: minimal health endpoint so the service builds, runs,
# and can be curled independently. Real search logic (register, OTP, login,
# JWT) comes later — see docs/02-services.md.
app = FastAPI(title="search")


@app.get("/health")
def health():
    return {"status": "ok", "service": "search"}
