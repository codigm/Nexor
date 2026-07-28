# ============================================================
# ATHENA — Fraud Detection Engine
# Multi-stage Go build for a minimal production image
# ============================================================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go module downloads
RUN apk add --no-cache git

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o athena ./main.go

# --- Final minimal image ---
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/athena .

EXPOSE 8085

CMD ["./athena"]
