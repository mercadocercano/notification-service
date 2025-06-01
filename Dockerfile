# Build stage
FROM golang:1.24-alpine AS builder

# Instalar dependencias de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod and sum files primero para aprovechar cache de Docker
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copiar el código fuente
COPY . .

# Build the application con flags de optimización
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o notification-service \
    ./src/main.go

# Final stage
FROM alpine:latest

# Instalar dependencias runtime
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Crear usuario no-root para seguridad
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copiar el binario desde builder stage
COPY --from=builder /app/notification-service .

# Copiar archivos de configuración
COPY --from=builder /app/config ./config

# Copiar templates de email
COPY --from=builder /app/templates ./templates

# Crear directorio para logs
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8282

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8282/api/v1/health || exit 1

# Variables de entorno por defecto
ENV SERVER_PORT=8282
ENV SERVER_MODE=release

# Run the application
CMD ["./notification-service"] 