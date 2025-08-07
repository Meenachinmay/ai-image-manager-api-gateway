# syntax=docker/dockerfile:1
# ─────────────────────────────────────────────
# Stage 1: builder – download modules & compile
FROM golang:1.23.5-alpine AS builder
LABEL stage=builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/server ./cmd/server

# ─────────────────────────────────────────────
FROM golang:1.23.5-alpine AS runner
WORKDIR /app

# Add ca-certificates for HTTPS calls and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /out/server ./server

# No need to copy config directory as configuration is loaded from environment variables
# If you have static files or templates to copy, add them here like:
# COPY --from=builder /src/templates ./templates

EXPOSE 8080

ENV APP_ENVIRONMENT=local
CMD ["./server"]