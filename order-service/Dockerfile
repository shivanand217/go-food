FROM golang:1.24-alpine AS builder

WORKDIR /app
# Order service requires CGO for confluent-kafka-go
RUN apk add --no-cache gcc libc-dev pkgconf
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o main .

FROM alpine:latest  
RUN apk add --no-cache libc6-compat
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .env* ./
CMD ["./main"]
