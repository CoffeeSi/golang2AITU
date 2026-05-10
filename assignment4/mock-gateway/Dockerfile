FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mock-gateway ./main.go


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/mock-gateway .

CMD ["./mock-gateway"]