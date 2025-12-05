FROM golang:1.25.5-bullseye as builder

ENV buildDeps=' \
        build-essential \
        libsqlite3-dev \
        gcc \
        musl-dev \
    '

RUN apt-get update \
    && apt-get install -y $buildDeps --no-install-recommends

WORKDIR /build

# CGO_ENABLED=1 due to go-sqlite :(
ENV CGO_ENABLED=1

# Load dependencies
COPY go.mod go.sum ./
RUN go mod download

# COPY the source code as the last step
COPY . .

RUN go install -ldflags='-s -w -extldflags "-static"' cmd/github-release-monitor-bot/main.go

FROM alpine:3.23

COPY --from=builder /go/bin/main /usr/local/bin/main-app

ENTRYPOINT [ "main-app" ]
