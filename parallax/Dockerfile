FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies if needed
RUN apk add --no-cache gcc musl-dev

# Now that the build context is the NEXOR root directory,
# we can copy everything so go.mod's replace directive works correctly.
COPY . .

# Change to the parallax directory to build
WORKDIR /app/parallax
RUN go env -w GOSUMDB=off
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o parallax-core .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/parallax/parallax-core .

CMD ["./parallax-core"]
