FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o doctor-service ./cmd/doctor-service/main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/doctor-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./doctor-service"]