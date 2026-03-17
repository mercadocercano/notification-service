# -----------------------------
# Etapa de compilación (builder)
# -----------------------------
FROM golang:1.24-alpine AS builder

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
# Etapa final (runtime)
# -----------------------------
FROM alpine:latest

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
