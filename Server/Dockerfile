# -------- Builder --------
    FROM golang:1.24 AS builder

    # Рабочая директория
    WORKDIR /app
    
    # Копируем go.mod и go.sum
    COPY go.mod go.sum ./
    
    # Устанавливаем зависимости
    RUN go mod download
    
    # Копируем исходники
    COPY . .
    
    # Устанавливаем goose для миграций
    RUN go install github.com/pressly/goose/v3/cmd/goose@latest
    
    # Собираем сервер (статически линкованный бинарник под Linux)
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server/
    
    # Проверяем, что бинарник собран
    RUN ls -l /app/main
    
    # -------- Runtime --------
    FROM alpine:latest
    
    WORKDIR /app
    
    # Минимальные зависимости для запуска бинарника/goose
    RUN apk add --no-cache libc6-compat
    
    # Копируем бинарник и миграции
    COPY --from=builder /app/main .
    COPY --from=builder /app/migrations /app/migrations/
    
    # Копируем goose
    COPY --from=builder /go/bin/goose /usr/local/bin/goose
    
    # Порт приложения
    EXPOSE 8090
    
    # Запускаем миграции и приложение
    CMD ["sh", "-c", "/usr/local/bin/goose -dir /app/migrations postgres \"$DATABASE_URL\" up && ./main"]