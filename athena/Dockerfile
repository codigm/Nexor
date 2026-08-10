# ============================================================
# ATHENA — Fraud Detection Engine
# Multi-stage Go build for a minimal production image
# Multi-stage Rust build for a minimal production image
# ============================================================
FROM rust:alpine AS builder

WORKDIR /app
# Install build dependencies for Rust on Alpine
RUN apk add --no-cache musl-dev gcc

COPY Cargo.toml ./
COPY src ./src

RUN cargo build --release

# Final minimal image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/target/release/athena-core .

EXPOSE 8085
CMD ["./athena-core"]
