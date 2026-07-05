#!/usr/bin/env bash
# Smoke test: boot the resume-app server, probe /health and /metrics, then exit.
# Usage: ./backend/scripts/smoke.sh
set -euo pipefail

PORT="${BACKEND_PORT:-18080}"
export BACKEND_PORT="$PORT"
export DB_PATH="${DB_PATH:-/tmp/resume-smoke.db}"
export LLM_PROVIDER="${LLM_PROVIDER:-ollama}"
export OUTPUT_DIR="${OUTPUT_DIR:-/tmp/resume-smoke-out}"
mkdir -p "$OUTPUT_DIR"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin/smoke"
mkdir -p "$ROOT/bin"

echo "Building smoke binary..."
( cd "$ROOT" && go build -o "$BIN" ./cmd )

echo "Starting server on :$PORT..."
"$BIN" &
SERVER_PID=$!
trap 'kill $SERVER_PID 2>/dev/null || true; rm -f "$BIN"' EXIT

echo "Polling /health (timeout 10s)..."
for _ in $(seq 1 100); do
	if curl -sf "http://localhost:$PORT/health" >/tmp/smoke-health.out 2>/dev/null; then
		break
	fi
	sleep 0.1
done

HEALTH_BODY="$(cat /tmp/smoke-health.out 2>/dev/null || true)"
if [[ "$HEALTH_BODY" != '{"status":"ok"}' ]]; then
	echo "FAIL: /health body = '$HEALTH_BODY', want '{\"status\":\"ok\"}'"
	exit 1
fi
echo "OK: /health -> $HEALTH_BODY"

echo "Probing /metrics..."
METRICS_BODY="$(curl -sf "http://localhost:$PORT/metrics" || true)"
if [[ ! "$METRICS_BODY" =~ (^|[[:space:]])'# HELP'[[:space:]] ]]; then
	echo "FAIL: /metrics does not contain Prometheus # HELP line"
	echo "$METRICS_BODY" | head -20
	exit 1
fi
echo "OK: /metrics returns Prometheus format"

echo "Smoke test passed."
