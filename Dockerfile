# Stage 0: Build the React frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
ENV VITE_API_URL=""
RUN npm run build

# Stage 1: Build the Go binary (with embedded frontend)
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

WORKDIR /build

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and frontend dist for embed
COPY . .
COPY --from=frontend-builder /web/dist ./web/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o /api-server ./cmd/

# Stage 2: S6-overlay runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl

# Install S6-overlay v3
ARG S6_OVERLAY_VERSION=3.2.0.2
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz /tmp
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-x86_64.tar.xz /tmp
RUN tar -C / -Jxpf /tmp/s6-overlay-noarch.tar.xz && \
    tar -C / -Jxpf /tmp/s6-overlay-x86_64.tar.xz && \
    rm /tmp/s6-overlay-*.tar.xz

# S6 legacy service definitions (for legacy-services mode)
COPY docker/services.d/ /etc/services.d/

# Copy binary from builder
COPY --from=builder /api-server /usr/local/bin/api-server

# Cont-init initialization scripts
COPY docker/cont-init.d/ /etc/cont-init.d/

RUN mkdir -p /data

ENV RUSTDESK_API_SERVER_ADDR=0.0.0.0:21114
ENV RUSTDESK_API_DATABASE_PATH=/data/api.db
ENV RUSTDESK_API_LOG_PATH=/data/api.log
ENV S6_BUNDLE=default

EXPOSE 21114
EXPOSE 21115
EXPOSE 21116
EXPOSE 21117

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:21114/api/health || exit 1

ENTRYPOINT ["/init"]
