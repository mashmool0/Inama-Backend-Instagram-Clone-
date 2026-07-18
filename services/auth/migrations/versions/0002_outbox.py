"""add transactional outbox table

Revision ID: 0002_outbox
Revises: 0001_initial
Create Date: 2026-07-16
"""
from alembic import op
import sqlalchemy as sa

revision = "0002_outbox"
down_revision = "0001_initial"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "outbox",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("event_type", sa.String(length=100), nullable=False),
        sa.Column("payload", sa.JSON(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.text("now()"), nullable=False),
        sa.Column("published_at", sa.DateTime(timezone=True), nullable=True),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index("ix_outbox_published_at", "outbox", ["published_at"])


def downgrade() -> None:
    op.drop_table("outbox")
