## Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-X main.version=${VERSION}" -o webtail .

## Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    addgroup -S webtail && \
    adduser -S webtail -G webtail

WORKDIR /app
COPY --from=builder /build/webtail /app/webtail

# /config holds config.json (mount as read-only)
# /data holds tsnet state (persisted via named volume)
RUN mkdir -p /config /data && chown webtail:webtail /config /data

USER webtail

# Route os.UserConfigDir() → /data so tsnet state lands at /data/webtail/<node>
ENV XDG_CONFIG_HOME=/data

VOLUME ["/config", "/data"]

ENTRYPOINT ["/app/webtail", "--config", "/config/config.json"]
