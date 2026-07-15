# Build context is the REPO ROOT (see docker-compose.yml) so the image can copy
# the generated proto stubs from proto/gen alongside the service code.
FROM python:3.12-slim
WORKDIR /app

COPY services/auth/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY services/auth/ /app/
COPY proto/gen/ /app/gen/

# /app          → the `app` package + main.py
# /app/gen      → the generated `auth` package (from auth.proto)
ENV PYTHONPATH=/app:/app/gen

EXPOSE 50051

# Dev bootstrap: make ephemeral RS256 keys, apply migrations, then serve.
# In prod, keys come from a secret store and migrations run as a separate step.
CMD ["sh", "-c", "python scripts/generate_keys.py && alembic upgrade head && python main.py"]
