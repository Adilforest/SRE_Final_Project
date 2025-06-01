# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

# Release stage
FROM alpine:3.19

WORKDIR /app

# Копируем бинарник из build stage
COPY --from=builder /app/order-service .

# Копируем .env если нужно (опционально)
COPY .env .env

# Открываем порт gRPC
EXPOSE 50053

# Запуск сервиса
CMD ["./order-service"]