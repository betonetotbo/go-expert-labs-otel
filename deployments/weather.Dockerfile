FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -ldflags="-w -s" -o weather cmd/weather/main.go

FROM alpine:3.20.3
WORKDIR /app
COPY --from=builder /app/weather .
ENTRYPOINT ["/app/weather"]