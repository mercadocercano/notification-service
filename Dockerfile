# -----------------------------
# Etapa de compilación (builder)
# -----------------------------
    FROM golang:1.24-alpine AS builder

    # Instalar dependencias de build
    RUN apk add --no-cache git ca-certificates tzdata
    
    WORKDIR /app
    
    # Copiar ficheros de módulos para aprovechar cache
    COPY go.mod go.sum ./
    
    # Descargar dependencias
    RUN go mod download
    
    # Copiar todo el código fuente
    COPY . .
    
    # Compilar la aplicación con flags de optimización y estática
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o notification-service \
        ./src/main.go
    
    # -----------------------------
    # Etapa final (runtime)
    # -----------------------------
    FROM alpine:latest
    
    # Instalar dependencias de runtime (incluye wget para healthcheck)
    RUN apk --no-cache add \
        ca-certificates \
        tzdata \
        wget \
      && update-ca-certificates
    
    # Crear usuario no-root
    RUN addgroup -g 1001 -S appgroup && \
        adduser -u 1001 -S appuser -G appgroup
    
    WORKDIR /app
    
    # Copiar el binario compilado desde builder
    COPY --from=builder /app/notification-service .
    
    # Copiar archivos de configuración y plantillas
    COPY --from=builder /app/config ./config
    COPY --from=builder /app/templates ./templates
    
    # Crear directorio para logs y asignar propiedad
    RUN mkdir -p /app/logs && \
        chown -R appuser:appgroup /app
    
    # Cambiar a usuario no-root
    USER appuser
    
    # Exponer puerto en el que corre el servicio
    EXPOSE 8282
    
    # Healthcheck: verifica que /health devuelva 200 OK
    HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
        CMD wget --no-verbose --tries=1 --spider http://localhost:8282/health || exit 1
    
    # Variables de entorno por defecto (pueden sobreescribirse desde Coolify)
    ENV SERVER_PORT=8282
    ENV SERVER_MODE=release
    
    # Comando por defecto para arrancar la app
    CMD ["./notification-service"]