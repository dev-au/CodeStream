FROM golang:1.26.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o github.com/dev-au/CodeStream ./cmd/
FROM alpine:3.18.4
WORKDIR /app

COPY --from=builder /app/github.com/dev-au/CodeStream /app/github.com/dev-au/CodeStream

RUN chmod +x /app/github.com/dev-au/CodeStream

COPY scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

