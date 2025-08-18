FROM golang:1.22-alpine

ENV GOCACHE=/go-cache \
    GOMODCACHE=/go-mod-cache

RUN mkdir -p /go-cache /go-mod-cache

RUN go install std

WORKDIR /app
