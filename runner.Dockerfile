FROM debian:bookworm-slim AS base
WORKDIR /app

RUN apt-get update && apt-get install -y \
    curl wget git build-essential time ca-certificates \
    && rm -rf /var/lib/apt/lists/*


RUN apt-get update && apt-get install -y python3\
    && python3 --version

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && node -v && npm -v


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


RUN apt-get update && apt-get install -y gcc g++



WORKDIR /app
CMD ["/bin/bash"]
