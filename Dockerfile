# Build stage
FROM golang:1.25.6-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o erebus .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/erebus .

# Copy any data files if needed
COPY --from=builder /app/html/pages/manifest.tmpl ./html/pages/manifest.tmpl
COPY --from=builder /app/manifest ./manifest
COPY --from=builder /app/words_manifesto ./words_manifesto

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./erebus"]
