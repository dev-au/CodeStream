FROM debian:bookworm-slim AS base
WORKDIR /app

# Umumiy paketlar
RUN apt-get update && apt-get install -y \
    curl wget git build-essential time ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# =====================
# Python 3.11
# =====================
RUN apt-get update && apt-get install -y python3 python3-pip \
    && python3 --version && pip3 --version

# =====================
# Node.js 20 (Nodesource’dan)
# =====================
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && node -v && npm -v

# =====================
# Go 1.24.4 (official tarball)
# =====================
RUN curl -OL https://go.dev/dl/go1.24.4.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz \
    && rm go1.24.4.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}" \
    GOCACHE=/go-cache \
    GOMODCACHE=/go-mod-cache

RUN mkdir -p /go-cache /go-mod-cache \
    && go version
RUN go install std
# =====================
# GCC 13.2.0
# =====================
RUN apt-get update && apt-get install -y gcc g++



WORKDIR /app
CMD ["/bin/bash"]
