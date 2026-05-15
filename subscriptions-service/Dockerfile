FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/server ./cmd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/internal/handler/static ./internal/handler/static
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
