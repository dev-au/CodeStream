FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o platform .

FROM alpine:latest
RUN apk add --no-cache docker-cli bash

WORKDIR /app
COPY --from=builder /app/platform .

EXPOSE 8000

ENV DOCKER_HOST=unix:///var/run/docker.sock
CMD ["./platform"]


# docker build -t runner-node:latest runner-images/runner-node