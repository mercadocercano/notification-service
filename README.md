# Notification Service

Servicio de notificaciones que permite enviar diferentes tipos de notificaciones (email, SMS) utilizando templates. Diseñado para una fácil integración en una arquitectura de microservicios.

## Requisitos Previos

Antes de comenzar, asegúrate de tener instalado:

- **Go 1.21 o superior**
  
  **Instalación en macOS:**
  ```bash
  # Opción 1: Usando Homebrew (recomendado)
  brew install go
  
  # Opción 2: Descargar desde el sitio oficial
  # Visita https://golang.org/dl/ y descarga el instalador para macOS
  ```
  
  **Verificar instalación:**
  ```bash
  go version
  ```

- Git
  ```bash
  # Verificar versión de Git
  git --version
  ```
- **Resend API Key (para envío de emails)** - REQUERIDO
  - Obtén tu API Key en [Resend](https://resend.com)
  - 📖 **Ver guía completa**: [docs/CONFIGURACION_API_KEYS.md](docs/CONFIGURACION_API_KEYS.md)

## Instalación

1. **Clonar el repositorio**
   ```bash
   git clone https://github.com/leopegorin/notification-service.git
   cd notification-service
   ```

2. **Instalar dependencias**
   ```bash
   # Descargar todas las dependencias necesarias
   go mod tidy
   ```

3. **Configurar API Keys**
   ```bash
   # Copiar archivo de ejemplo de variables de entorno
   cd ../..  # Ir a raíz del proyecto
   cp .env.example .env

   # Editar .env y agregar tu RESEND_API_KEY
   nano .env
   # O usar tu editor favorito:
   # code .env
   # vim .env
   ```

   **⚠️ IMPORTANTE**: Debes configurar `RESEND_API_KEY` con una clave válida de Resend.
   
   📖 **Ver guía completa**: [docs/CONFIGURACION_API_KEYS.md](docs/CONFIGURACION_API_KEYS.md)

## Ejecución

### Modo Desarrollo

1. **Iniciar el servicio**
   ```bash
   go run src/main.go
   ```

2. **Verificar que el servicio está funcionando**
   ```bash
   curl http://localhost:8282/health
   ```

### Modo Producción

1. **Compilar el servicio**
   ```bash
   go build -o notification-service src/main.go
   ```

2. **Ejecutar el binario**
   ```bash
   ./notification-service
   ```

### Usando Docker

1. **Construir la imagen**
   ```bash
   docker build -t notification-service .
   ```

2. **Ejecutar el contenedor**
   ```bash
   docker run -p 8282:8282 notification-service
   ```

## Pruebas con Postman

Hemos incluido una colección completa de Postman para facilitar las pruebas del servicio:

### Importar la Colección

1. **Importar archivos en Postman:**
   - `postman/notification-service.postman_collection.json` - Colección principal
   - `postman/notification-service.postman_environment.json` - Variables de entorno

2. **Requests incluidos:**
   - Health Check
   - Envío de notificaciones (Welcome, Verification, Password Reset)
   - Notificaciones asíncronas
   - Consulta de estado de notificaciones
   - Tests de manejo de errores

3. **Tests automáticos incluidos** para verificar respuestas y manejo de errores

📁 **Ver documentación completa**: [postman/README.md](postman/README.md)

## Verificación de la Instalación

Para verificar que todo está funcionando correctamente:

1. **Verificar la estructura del proyecto**
   ```bash
   tree -L 3
   ```

2. **Ejecutar los tests**
   ```bash
   go test ./...
   ```

3. **Probar el envío de una notificación**
   ```bash
   curl -X POST http://localhost:8282/api/v1/notifications \
     -H "Content-Type: application/json" \
     -d '{
       "type": "email",
       "template_id": "welcome_email",
       "recipient": "usuario@ejemplo.com",
       "data": {
         "name": "Juan Pérez",
         "email": "usuario@ejemplo.com"
       },
       "async": false
     }'
   ```

## Estructura del Proyecto

```
notification-service/
├── src/
│   ├── main.go                              # Punto de entrada principal
│   ├── notification/
│   │   ├── domain/                          # Entidades y reglas de negocio
│   │   │   ├── notification.go
│   │   │   └── template.go
│   │   ├── application/                     # Casos de uso y lógica de aplicación
│   │   │   ├── request/                     # DTOs de entrada
│   │   │   ├── response/                    # DTOs de salida
│   │   │   └── usecase/                     # Casos de uso
│   │   ├── infrastructure/                  # Implementaciones técnicas
│   │   │   ├── controller/                  # Controladores HTTP
│   │   │   ├── email/                       # Cliente de email (Resend)
│   │   │   └── config/                      # Configuración del módulo
│   │   └── ports/                           # Interfaces
│   │       ├── input/                       # Puertos de entrada
│   │       └── output/                      # Puertos de salida
│   └── shared/                              # Código compartido
│       ├── config/                          # Configuración global
│       ├── logger/                          # Logger
│       └── metrics/                         # Métricas
├── pkg/                                     # Utilidades reutilizables
│   ├── templates/                           # Procesador de templates
│   └── validator/                           # Validadores
├── postman/                                 # Colección de Postman
├── config/                                  # Archivos de configuración
├── go.mod                                   # Dependencias de Go
├── Dockerfile                               # Imagen Docker
└── README.md                                # Documentación
```

## Desarrollo

### Ejecutar tests

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con cobertura
go test -cover ./...
```

### Linting

```bash
# Instalar golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Ejecutar linting
golangci-lint run
```

## Monitoreo

El servicio expone métricas en el puerto 9090 (configurable) en el endpoint `/metrics`. Las métricas incluyen:

- Total de notificaciones enviadas
- Latencia de envío de notificaciones
- Tasa de éxito/fallo

## Solución de Problemas

### Errores comunes

1. **Error con la dependencia de Resend**
   ```bash
   # Si ves un error como:
   # go: github.com/resend/resend-go/v2@v2.0.0: go.mod has non-.../v2 module path
   
   # Ejecuta:
   go get github.com/resendlabs/resend-go
   go mod tidy
   ```

2. **Error: missing go.sum entry**
```bash
go mod tidy
```

3. **Error: cannot find package**
```bash
go mod download
go mod tidy
```

4. **Error: undefined: Queue**
Asegúrate de que la interfaz Queue esté definida en el paquete correcto.

## Contribuir

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/amazing-feature`)
3. Commit tus cambios (`git commit -m 'Add some amazing feature'`)
4. Push a la rama (`git push origin feature/amazing-feature`)
5. Abre un Pull Request

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.
