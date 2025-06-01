# Build stage
FROM golang:1.23 AS builder

# 1. Создаем структуру проекта внутри контейнера
WORKDIR /workspace
RUN mkdir -p services/auth-service && \
    mkdir -p services/product-service

# 2. Копируем workspace файлы
COPY go.work go.work.sum ./

# 3. Копируем модульные файлы
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
COPY services/product-service/go.mod services/product-service/go.sum ./services/product-service/

# 4. Скачиваем зависимости
RUN cd services/product-service && \
    go mod download

# 5. Копируем весь код
COPY . .

# 6. Билдим
RUN cd services/product-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/product-service ./cmd/main.go

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/product-service .
CMD ["/app/product-service"]