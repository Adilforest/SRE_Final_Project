# Используем официальный образ Go
FROM golang:1.21 as builder

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway

# Используем минимальный образ для финального контейнера
FROM alpine:latest

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /app/api-gateway .

# Экспонируем порт
EXPOSE 8080

# Команда для запуска приложения
CMD ["./api-gateway"]