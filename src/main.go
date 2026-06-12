package main

import (
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
    notificationConfig "notification-service/src/notification/infrastructure/config"
    "notification-service/src/shared/config"
    "notification-service/src/shared/logger"
    "notification-service/src/shared/middleware"
)

func main() {
    // Cargar configuración
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Inicializar logger
    if err := logger.InitLogger(); err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }

    // Configuración del router
    router := gin.New()

    // Agregar middlewares básicos
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(tenantmw.TenantValidation(tenantmw.TenantValidationConfig{
        JWTSecret: os.Getenv("JWT_SECRET"),
        ExcludedRoutes: []string{
            "/health",
			"/api/v1/health",
            "/metrics",
        },
    }))

    // Middleware de manejo de errores centralizado
    router.Use(middleware.ErrorHandlerMiddleware())

    // Configuración de CORS
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":  "up",
            "service": "notification-service",
        })
    })

    // API v1 group
    apiV1 := router.Group("/api/v1")

    // Configurar módulo de notificaciones
    notificationConfig.SetupNotificationModule(apiV1, cfg)

    // Iniciar el servidor
    port := fmt.Sprintf(":%d", cfg.Server.Port)
    log.Printf("Starting notification service on port %d...", cfg.Server.Port)
    if err := router.Run(port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
} 