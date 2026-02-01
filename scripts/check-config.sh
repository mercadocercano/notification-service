#!/bin/bash

# ============================================================================
# Script de Verificación de Configuración - Notification Service
# ============================================================================
# Verifica que todas las variables de entorno necesarias estén configuradas
# ============================================================================

set -e

echo "=========================================="
echo "🔍 Verificando Configuración"
echo "=========================================="
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Contadores
ERRORS=0
WARNINGS=0

# Función para verificar variable de entorno
check_env_var() {
    local var_name=$1
    local required=$2
    local default_value=$3
    
    if [ -z "${!var_name}" ]; then
        if [ "$required" = "true" ]; then
            echo -e "${RED}❌ ERROR: $var_name no está configurada${NC}"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${YELLOW}⚠️  WARNING: $var_name no está configurada (opcional)${NC}"
            if [ ! -z "$default_value" ]; then
                echo "   Valor por defecto: $default_value"
            fi
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        # Ocultar valores sensibles
        if [[ $var_name == *"KEY"* ]] || [[ $var_name == *"PASSWORD"* ]] || [[ $var_name == *"SECRET"* ]]; then
            local masked_value="${!var_name:0:10}..."
            echo -e "${GREEN}✅ $var_name está configurada ($masked_value)${NC}"
        else
            echo -e "${GREEN}✅ $var_name está configurada: ${!var_name}${NC}"
        fi
    fi
}

echo "📦 Variables de Entorno:"
echo ""

# Variables requeridas
echo "🔴 REQUERIDAS:"
check_env_var "RESEND_API_KEY" "true"
echo ""

# Variables importantes (con valores por defecto)
echo "🟡 IMPORTANTES (con valores por defecto):"
check_env_var "RESEND_FROM_EMAIL" "false" "noreply@mercadocercano.com"
check_env_var "SERVER_PORT" "false" "8282"
check_env_var "SERVER_MODE" "false" "development"
echo ""

# Variables de base de datos
echo "🗄️  BASE DE DATOS:"
check_env_var "DATABASE_HOST" "false" "localhost"
check_env_var "DATABASE_PORT" "false" "5432"
check_env_var "DATABASE_USER" "false" "postgres"
check_env_var "DATABASE_PASSWORD" "false" "postgres"
check_env_var "DATABASE_NAME" "false" "notification_db"
echo ""

# Variables opcionales
echo "⚪ OPCIONALES:"
check_env_var "REDIS_HOST" "false" "localhost"
check_env_var "REDIS_PORT" "false" "6379"
check_env_var "SQS_ENABLED" "false" "false"
check_env_var "METRICS_ENABLED" "false" "true"
echo ""

# Verificar archivo .env
echo "=========================================="
echo "📄 Archivos de Configuración:"
echo "=========================================="
echo ""

if [ -f "../../.env" ]; then
    echo -e "${GREEN}✅ Archivo .env encontrado en raíz del proyecto${NC}"
elif [ -f ".env" ]; then
    echo -e "${GREEN}✅ Archivo .env encontrado en directorio del servicio${NC}"
else
    echo -e "${YELLOW}⚠️  No se encontró archivo .env${NC}"
    echo "   Crear desde .env.example:"
    echo "   cd ../.. && cp .env.example .env"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# Verificar config.yaml
if [ -f "config/config.yaml" ]; then
    echo -e "${GREEN}✅ Archivo config/config.yaml encontrado${NC}"
    
    # Verificar si tiene API key hardcodeada
    if grep -q "api_key.*re_" config/config.yaml 2>/dev/null; then
        echo -e "${RED}❌ ERROR: config.yaml contiene API key hardcodeada${NC}"
        echo "   Las API keys deben estar en variables de entorno"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  No se encontró config/config.yaml (se usarán solo variables de entorno)${NC}"
fi
echo ""

# Verificar conectividad a Resend (si está configurado)
if [ ! -z "$RESEND_API_KEY" ] && [ "$RESEND_API_KEY" != "your_resend_api_key_here" ]; then
    echo "=========================================="
    echo "🌐 Verificando Conectividad a Resend"
    echo "=========================================="
    echo ""
    
    # Test simple de API key (solo verifica formato)
    if [[ $RESEND_API_KEY == re_* ]]; then
        echo -e "${GREEN}✅ API key tiene formato válido (re_...)${NC}"
    else
        echo -e "${YELLOW}⚠️  API key no tiene el formato esperado (debería empezar con 're_')${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
    echo ""
fi

# Resumen
echo "=========================================="
echo "📊 Resumen"
echo "=========================================="
echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ Configuración correcta!${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Hay $WARNINGS advertencias (no críticas)${NC}"
    fi
    echo ""
    echo "🚀 Puedes iniciar el servicio:"
    echo "   make dev-start"
    echo "   o"
    echo "   docker-compose up -d"
    exit 0
else
    echo -e "${RED}❌ Hay $ERRORS errores que deben corregirse${NC}"
    echo ""
    echo "📖 Ver guía de configuración:"
    echo "   docs/CONFIGURACION_API_KEYS.md"
    echo ""
    echo "🔧 Pasos recomendados:"
    echo "   1. Crear archivo .env: cp ../../.env.example ../../.env"
    echo "   2. Obtener API key en: https://resend.com/api-keys"
    echo "   3. Editar .env y agregar RESEND_API_KEY=tu_api_key"
    echo "   4. Ejecutar nuevamente: ./scripts/check-config.sh"
    exit 1
fi

