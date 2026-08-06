FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies if needed
RUN apk add --no-cache gcc musl-dev

# Now that the build context is the OLYMPUS root directory,
# we can copy everything so go.mod's replace directive works correctly.
COPY . .

# Change to the argus directory to build
WORKDIR /app/argus
RUN go env -w GOSUMDB=off
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o argus-core .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/argus/argus-core .

CMD ["./argus-core"]
