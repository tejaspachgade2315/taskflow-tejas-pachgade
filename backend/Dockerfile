FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /bin/api /app/api
COPY --from=builder /app/migrations /app/migrations

ENV APP_PORT=8080
ENV MIGRATIONS_DIR=/app/migrations

USER app
EXPOSE 8080

CMD ["/app/api"]
