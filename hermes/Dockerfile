# ============================================================
# HERMES — Saga Orchestrator (Go)
# Multi-stage build for minimal production image
# ============================================================
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .
RUN go env -w GOSUMDB=off
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hermes-app ./main.go

# --- Final minimal image ---
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/hermes-app .

CMD ["./hermes-app"]
