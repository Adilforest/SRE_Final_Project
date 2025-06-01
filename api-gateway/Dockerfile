# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/main.go

# Release stage
FROM alpine:3.19

WORKDIR /app

# Копируем бинарник из build stage
COPY --from=builder /app/api-gateway .

# Копируем .env (если нужно)
COPY .env .env

# Открываем порт HTTP (Gin)
EXPOSE 8080

# Запуск сервиса
CMD ["./api-gateway"]