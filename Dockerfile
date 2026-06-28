# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

WORKDIR /build

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o /api-server ./cmd/

# Stage 2: Minimal runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /api-server /usr/local/bin/api-server

RUN mkdir -p /data

ENV RUSTDESK_API_SERVER_ADDR=0.0.0.0:21114
ENV RUSTDESK_API_DATABASE_PATH=/data/api.db
ENV RUSTDESK_API_LOG_PATH=/data/api.log

EXPOSE 21114
EXPOSE 21115
EXPOSE 21116
EXPOSE 21117

VOLUME ["/data"]

ENTRYPOINT ["api-server"]
CMD ["serve"]
