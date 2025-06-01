#!/bin/bash

# Script para build y deploy del notification service

echo "🐳 Building Notification Service Docker Image..."

# Build de la imagen
docker-compose build notification-service

if [ $? -eq 0 ]; then
    echo "✅ Build completado exitosamente"
    
    echo "🚀 Iniciando servicios..."
    
    # Levantar todos los servicios
    docker-compose up -d
    
    if [ $? -eq 0 ]; then
        echo "✅ Servicios iniciados exitosamente"
        echo ""
        echo "📊 Estado de los servicios:"
        docker-compose ps
        echo ""
        echo "🔗 URLs disponibles:"
        echo "   - API Health Check: http://localhost:8282/api/v1/health"
        echo "   - API Notifications: http://localhost:8282/api/v1/notifications"
        echo "   - API Queue Status: http://localhost:8282/api/v1/queue/status"
        echo "   - Métricas: http://localhost:9090"
        echo ""
        echo "📝 Para ver logs:"
        echo "   docker-compose logs -f notification-service"
        echo ""
        echo "🛑 Para detener:"
        echo "   docker-compose down"
    else
        echo "❌ Error iniciando servicios"
        exit 1
    fi
else
    echo "❌ Error en el build"
    exit 1
fi 