# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=build /out/server /app/server

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/health >/dev/null || exit 1

ENTRYPOINT ["/app/server"]
