#!/bin/sh
set -e

echo "Executing database migrations..."
migrate -path=./migrations -database postgres://postgres:postgres@postgres:5432/auth_service?sslmode=disable up

echo "Starting auth service"
exec ./auth-service
