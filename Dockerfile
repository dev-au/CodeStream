FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk add --no-cache docker-cli bash
WORKDIR /root/
COPY --from=builder /app/main .
COPY .env .
COPY templates/ templates/
ENV DOCKER_HOST=unix:///var/run/docker.sock
CMD ["./main"]


# docker build -t runner-node:latest runner-images/runner-node