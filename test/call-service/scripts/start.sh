#!/bin/sh
set -e

echo "Executing database migrations..."
migrate -path=./migrations -database postgres://postgres:postgres@postgres:5432/call_service?sslmode=disable up

echo "Starting call service"
exec ./call-service
