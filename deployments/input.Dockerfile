FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -ldflags="-w -s" -o input cmd/input/main.go

FROM alpine:3.20.3
WORKDIR /app
COPY --from=builder /app/input .
ENTRYPOINT ["/app/input"]