# call-service/Dockerfile
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Устанавливаем утилиту для миграций
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

RUN echo 'migrate -path=./migrations -database postgres://postgres:postgres@postgres:5432/call_service?sslmode=disable up'

# Копируем и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o call-service ./main.go

# Создаем минимальный образ
FROM alpine:latest

WORKDIR /app

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Копируем бинарный файл migrate
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Копируем скомпилированное приложение
COPY --from=builder /app/call-service .

# Здоровье приложения
HEALTHCHECK --interval=5s --timeout=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Открываем порт для HTTP
EXPOSE 8080

# Скрипт запуска
COPY --from=builder /app/scripts/start.sh .
RUN chmod +x ./start.sh

CMD ["./start.sh"]
