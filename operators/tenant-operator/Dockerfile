# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tenant-operator .

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /app/tenant-operator .

# Run as non-root user
USER 65532:65532

ENTRYPOINT ["/tenant-operator"]

