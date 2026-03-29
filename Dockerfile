# Alfredo — modular monolith
# NOTE: Calendar (EventKit) is macOS-only. In Linux containers, calendar
# operations degrade gracefully — they log errors and the pet-care data still saves.

FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o alfredo ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/alfredo .

RUN mkdir -p /app/data /app/photos

EXPOSE 8080
CMD ["./alfredo"]
