# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download || true

# Copy source code
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hirer-api .

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/hirer-api .

# Run as non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/hirer-api"]

