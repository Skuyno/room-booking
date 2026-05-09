FROM golang:1.25.6 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/api

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/bin/server /app/server

EXPOSE 8080

CMD ["/app/server"]
