# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/cinema-backend \
    ./cmd/main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 cinema \
    && adduser -S -D -H -u 10001 -G cinema cinema

WORKDIR /app

COPY --from=builder --chown=cinema:cinema /out/cinema-backend /app/cinema-backend

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/app/cinema-backend"]
