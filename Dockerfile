# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# -ldflags="-w -s" strips debug information and symbols, reducing binary size
# CGO_ENABLED=0 ensures the binary is statically linked and doesn't depend on C libraries (important for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o libreplex main.go

# Stage 2: Create the final, minimal image
FROM debian:stable-slim

# Install Calibre and other necessary dependencies
# sudo is needed for the Calibre install script
# wget and ca-certificates are needed to download the Calibre installer
# xz-utils is needed by the Calibre installer
# libopengl0 is a runtime dependency for Calibre's ebook-convert
RUN apt-get update && apt-get install -y \
    sudo \
    wget \
    ca-certificates \
    xz-utils \
    libopengl0 \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Download and install Calibre
# The script installs Calibre to /opt/calibre
RUN wget -nv -O- https://download.calibre-ebook.com/linux-installer.sh | sudo sh /dev/stdin version=7.15.0 installdir=/opt/calibre کتابخانه=/opt/calibre/calibre_books calibre_desktop_installation=n

# Add Calibre to PATH so ebook-meta and ebook-convert can be called directly
ENV PATH="/opt/calibre:${PATH}"

WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/libreplex .

# Copy the frontend directory
COPY frontend ./frontend

# Create non-root user and group for security
RUN groupadd -r appgroup && useradd -r -g appgroup -d /app -s /sbin/nologin -c "Docker App User" appuser
RUN chown -R appuser:appgroup /app && \
    chown -R appuser:appgroup ./frontend
# /config and /books will be mounted as volumes, ownership will be handled by docker-compose or user
# Create directories for volumes if they don't exist and set permissions
# These directories will be owned by root initially, but Docker volumes will overlay them.
# For config, it's good if the app user can write to it if it's not explicitly mounted with other permissions.
RUN mkdir -p /config /books /cache && \
    chown appuser:appgroup /config && \
    chown appuser:appgroup /books && \
    chown appuser:appgroup /cache


USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Define a healthcheck (optional, but good practice)
# This will try to connect to the /api/health endpoint every 30s after an initial 5s delay
# It will try 3 times before marking the container as unhealthy.
# Note: wget needs to be installed. If using a super minimal base without it, this needs adjustment.
# We installed wget earlier, so this should be fine.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
ENTRYPOINT ["./libreplex"]
