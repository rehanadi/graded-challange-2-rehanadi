# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required dependencies for postgres driver
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o server main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk add --no-cache libc6-compat postgresql-client

# Copy the binary and required files from builder
COPY --from=builder /app/server .
COPY --from=builder /app/.env .
COPY --from=builder /app/ddl.sql .

# Create directory for cloud sql proxy
RUN mkdir /cloudsql

EXPOSE ${SERVER_PORT}

CMD ["./server"]