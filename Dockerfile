FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk --no-cache add curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/bin/migrate

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server -ldflags="-s -w" ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /usr/bin/migrate /usr/bin/migrate

# Copy migrations and entrypoint
COPY migrations /migrations
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./server"]
