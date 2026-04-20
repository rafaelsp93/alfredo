# Alfredo — modular monolith
# Runtime secrets such as Google Calendar and Telegram credentials are injected
# via environment variables by Docker Compose or CI, never baked into the image.

FROM golang:1.26-alpine AS builder

ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o alfredo ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/alfredo .

RUN mkdir -p /app/data /app/photos

EXPOSE 8080
CMD ["./alfredo"]
