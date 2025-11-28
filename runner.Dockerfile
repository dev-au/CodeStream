FROM debian:bookworm-slim AS base

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget git build-essential ca-certificates gcc g++ python3 \
    time \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get update && apt-get install -y --no-install-recommends nodejs \
    && npm cache clean --force \
    && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION=1.24.4
RUN curl -fsSL https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz

ENV PATH="/usr/local/go/bin:${PATH}" \
    GOCACHE=/go-cache \
    GOMODCACHE=/go-mod-cache

RUN mkdir -p /go-cache /go-mod-cache && go version
RUN go install std


FROM base AS deps
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


FROM base AS final
WORKDIR /app

COPY --from=deps /go-mod-cache /go-mod-cache
COPY --from=deps /go-cache /go-cache

COPY . .

RUN go mod tidy

CMD ["/bin/bash"]
