# Этап 1: Сборка приложения
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum для скачивания зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Собираем приложение.
# -o /app/server - выходной файл
# -ldflags="-s -w" - уменьшает размер бинарного файла
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server -ldflags="-s -w" ./main.go

# Этап 2: Создание минимального финального образа
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только собранный бинарный файл из этапа сборки
COPY --from=builder /app/server .

# Указываем, что контейнер будет слушать порт 8080
EXPOSE 8080

# Команда для запуска приложения
CMD ["./server"]
