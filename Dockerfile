# -----------------------------
# Etapa de compilación (builder)
# -----------------------------
FROM golang:1.25-alpine AS builder

# Instalar dependencias de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

# Copiar go.mod y go.sum para cache
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar la aplicación de forma estática
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o notification-service \
    ./src/main.go

# -----------------------------
# Etapa de desarrollo (development)
# -----------------------------
FROM golang:1.25-alpine AS development

RUN apk add --no-cache ca-certificates tzdata curl git \
    && cp /usr/share/zoneinfo/UTC /etc/localtime \
    && echo "UTC" > /etc/timezone

RUN addgroup -g 1001 -S appgroup && \
    adduser -S -D -h /app -s /bin/sh -G appgroup -u 1001 appuser

RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

# Copy go mod files first (for better caching)
COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download

# Copy source code
COPY --chown=appuser:appgroup . .

# Create directories and fix permissions for Air + Go mod cache
RUN mkdir -p tmp logs /go/pkg/mod && \
    chmod -R 777 /go/pkg && \
    chown -R appuser:appgroup /app tmp logs

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8282/health || exit 1

EXPOSE 8282

CMD sh -c 'if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi && air -c .air.toml'

# -----------------------------
# Etapa final (runtime/production)
# -----------------------------
FROM alpine:latest AS production

# Instalar dependencias de runtime, incluyendo wget y curl
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
  && update-ca-certificates

# Crear usuario no-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copiar el binario compilado y recursos
COPY --from=builder /app/notification-service .
COPY --from=builder /app/config ./config
COPY --from=builder /app/templates ./templates

# Crear carpeta de logs y ajustar permisos
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Cambiar a usuario no-root
USER appuser

# Exponer el puerto que utiliza el servicio
EXPOSE 8282

# Variables de entorno por defecto (puedes sobreescribirlas en Coolify)
ENV SERVER_PORT=8282
ENV SERVER_MODE=release

# Comando de arranque
CMD ["./notification-service"]
