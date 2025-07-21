# Build stage
FROM golang:1.21-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download || go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8082

# Run the binary
CMD ["./main"]