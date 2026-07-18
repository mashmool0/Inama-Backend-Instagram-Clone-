"""Shared RabbitMQ naming (see proto/EVENTS.md).

One topic exchange for the whole system; publishers use the event name as the
routing key; each consumer binds its own queue to the keys it cares about.
"""

EXCHANGE_NAME = "inama.events"
