#!/usr/bin/env bash
# register-connector.sh — registers the Debezium outbox connector for user-service.
# Usage: APP_ENV=local ./register-connector.sh [connect-url]
# Default connect URL: http://localhost:8083

set -euo pipefail

APP_ENV="${APP_ENV:-local}"
CONNECT_URL="${1:-http://localhost:8083}"
CONNECTOR_TEMPLATE="$(dirname "$0")/outbox-connector.json"
MAX_RETRIES=30
RETRY_INTERVAL=5

echo "Environment : ${APP_ENV}"
echo "Kafka Connect: ${CONNECT_URL}"

echo "Waiting for Kafka Connect..."
for i in $(seq 1 "$MAX_RETRIES"); do
  if curl -sf "${CONNECT_URL}/connectors" > /dev/null 2>&1; then
    echo "Kafka Connect is ready."
    break
  fi
  if [ "$i" -eq "$MAX_RETRIES" ]; then
    echo "ERROR: Kafka Connect did not become ready after $((MAX_RETRIES * RETRY_INTERVAL))s." >&2
    exit 1
  fi
  echo "  attempt $i/$MAX_RETRIES — retrying in ${RETRY_INTERVAL}s..."
  sleep "$RETRY_INTERVAL"
done

echo "Registering connector (topic: ${APP_ENV}.modami.user.events)..."
CONFIG=$(APP_ENV="$APP_ENV" envsubst < "$CONNECTOR_TEMPLATE")

HTTP_CODE=$(echo "$CONFIG" | curl -s -o /tmp/connect_response.json -w "%{http_code}" \
  -X POST "${CONNECT_URL}/connectors" \
  -H "Content-Type: application/json" \
  --data-binary @-)

case "$HTTP_CODE" in
  201)
    echo "Connector registered successfully (HTTP 201)."
    ;;
  409)
    echo "Connector already exists (HTTP 409) — skipping."
    ;;
  *)
    echo "ERROR: unexpected response HTTP ${HTTP_CODE}:" >&2
    cat /tmp/connect_response.json >&2
    exit 1
    ;;
esac
