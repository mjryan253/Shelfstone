# ---- Stage 1: Build the Go application (Unchanged, it's already good) ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the static Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o libreplex main.go


# ---- Stage 2: Create the final, minimal & functional image ----
FROM alpine:3.20

# Install only the necessary dependencies for Calibre's CLI tools on Alpine
# qt5-webkit is the core rendering engine for ebook-convert
# Other packages provide necessary fonts and libraries to avoid conversion errors.
RUN apk add --no-cache \
    qt5-webkit \
    font-noto-cjk \
    font-freefont \
    ttf-dejavu \
    ttf-liberation \
    mesa-gl \
    curl

# Download and install the official Calibre binaries directly
RUN curl -L https://download.calibre-ebook.com/7.15.0/calibre-7.15.0-x86_64.txz | tar -xJ -C /opt && \
    ln -s /opt/calibre/calibre /usr/bin/calibre && \
    ln -s /opt/calibre/ebook-convert /usr/bin/ebook-convert && \
    ln -s /opt/calibre/ebook-meta /usr/bin/ebook-meta

WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/libreplex .

# Copy the frontend directory
COPY frontend ./frontend

# Create non-root user and group
# Alpine uses UID/GID 1000 for the first user by default in some base images,
# but we specify it here to be explicit. Use a high number to avoid collisions.
RUN addgroup -g 1001 -S appgroup && \
    adduser -S -u 1001 -G appgroup -h /app appuser

# Create directories for volumes. Ownership will be managed by docker-compose.
RUN mkdir -p /config /books /cache

USER appuser

EXPOSE 8080

# Healthcheck using curl, which is included in Alpine and more common than wget
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1

ENTRYPOINT ["./libreplex"]