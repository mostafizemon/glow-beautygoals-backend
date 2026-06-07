# Build stage
FROM golang:alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api

# Run stage
FROM alpine:latest

WORKDIR /

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /api-server /api-server

# Accept PORT argument and expose it
ARG PORT=8085
EXPOSE ${PORT}

# Command to run the executable
CMD ["/api-server"]
