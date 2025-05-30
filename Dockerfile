FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp ./main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/myapp .
# Папку static и другие файлы добавим позже
EXPOSE 8080 
# Внутренний порт, на котором слушает приложение
CMD ["./myapp"]