# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
# If you have go.sum, uncomment the next line
# COPY go.sum ./

# Download dependencies
RUN go mod download || true

# Copy source code
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o candidate-api .

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/candidate-api .

# Run as non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/candidate-api", "-health-check"] || exit 1

# Run the application
ENTRYPOINT ["/app/candidate-api"]

