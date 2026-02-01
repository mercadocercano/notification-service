# Configuración de API Keys - Notification Service

## 📋 Resumen

El `notification-service` utiliza **Resend** como proveedor de emails. Este documento explica cómo obtener y configurar la API key necesaria.

## 🔑 Obtener API Key de Resend

### 1. Crear cuenta en Resend

1. Visita [https://resend.com](https://resend.com)
2. Regístrate con tu email
3. Verifica tu cuenta

### 2. Obtener API Key

1. Ve a [https://resend.com/api-keys](https://resend.com/api-keys)
2. Click en "Create API Key"
3. Dale un nombre (ej: "Desarrollo Local" o "Producción")
4. Selecciona los permisos:
   - ✅ **Send emails** (requerido)
   - ⚪ Full access (opcional, solo si necesitas otras funciones)
5. Click en "Create"
6. **⚠️ IMPORTANTE**: Copia la API key inmediatamente, solo se muestra una vez

### 3. Verificar dominio (Producción)

Para producción, necesitas verificar tu dominio:

1. Ve a [https://resend.com/domains](https://resend.com/domains)
2. Click en "Add Domain"
3. Ingresa tu dominio (ej: `mercadocercano.com`)
4. Sigue las instrucciones para configurar los registros DNS
5. Espera la verificación (puede tomar hasta 48 horas)

**Para desarrollo**: Puedes usar el dominio de prueba `onboarding@resend.dev` sin verificación.

## ⚙️ Configuración en el Proyecto

### Opción 1: Variables de Entorno (Recomendado)

1. **Crear archivo .env** en la raíz del proyecto:

```bash
cd /Users/hornosg/MyProjects/saas-mt
cp .env.example .env
```

2. **Editar .env** y agregar tu API key:

```bash
# Resend (Servicio de Email)
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxxxxxxx
RESEND_FROM_EMAIL=noreply@mercadocercano.com
```

3. **Reiniciar servicios**:

```bash
make dev-restart
```

### Opción 2: Docker Compose Directo

Si trabajas solo con el notification-service:

```bash
cd services/notification-service

# Editar docker-compose.yml y reemplazar:
RESEND_API_KEY: your_resend_api_key_here

# Levantar servicio
docker-compose up -d
```

### Opción 3: Config YAML (No Recomendado)

⚠️ **NO recomendado** - Solo para pruebas locales rápidas:

Editar `services/notification-service/config/config.yaml`:

```yaml
resend:
  api_key: "re_xxxxxxxxxxxxxxxxxxxxxxxxxx"
```

**⚠️ IMPORTANTE**: Nunca commitear API keys en archivos de configuración.

## 🧪 Probar Configuración

### 1. Verificar que el servicio está corriendo

```bash
curl http://localhost:8282/health
```

Respuesta esperada:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-31T..."
}
```

### 2. Enviar email de prueba

```bash
curl -X POST http://localhost:8282/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "template_id": "welcome_email",
    "recipient": "tu-email@ejemplo.com",
    "data": {
      "name": "Usuario Prueba",
      "email": "tu-email@ejemplo.com"
    },
    "async": false
  }'
```

### 3. Verificar logs

```bash
docker logs mc-notification-service
```

Si la API key es inválida, verás:
```
ERROR: Resend API error: Invalid API key
```

Si todo está bien:
```
INFO: Email sent successfully to tu-email@ejemplo.com
```

## 🔒 Seguridad

### Buenas Prácticas

1. ✅ **Nunca commitear** API keys al repositorio
2. ✅ **Usar variables de entorno** en todos los ambientes
3. ✅ **Rotar API keys** periódicamente en producción
4. ✅ **Usar API keys diferentes** para dev/staging/prod
5. ✅ **Limitar permisos** de las API keys (solo "Send emails")
6. ✅ **Monitorear uso** en el dashboard de Resend

### .gitignore

Verificar que `.env` está en `.gitignore`:

```bash
cat /Users/hornosg/MyProjects/saas-mt/.gitignore | grep .env
```

## 📊 Límites de Resend

### Plan Free (Desarrollo)

- ✅ 100 emails/día
- ✅ 1 dominio verificado
- ✅ API completa

### Plan Pro (Producción)

- ✅ 50,000 emails/mes
- ✅ Dominios ilimitados
- ✅ Soporte prioritario
- 💰 $20/mes

Ver precios actualizados: [https://resend.com/pricing](https://resend.com/pricing)

## 🐛 Troubleshooting

### Error: "Invalid API key"

1. Verificar que la API key está correctamente copiada (sin espacios)
2. Verificar que la variable de entorno está siendo leída:
   ```bash
   docker exec mc-notification-service env | grep RESEND
   ```
3. Regenerar la API key en Resend si es necesario

### Error: "Domain not verified"

- En desarrollo: Usar `onboarding@resend.dev` como remitente
- En producción: Verificar dominio en Resend dashboard

### Error: "Rate limit exceeded"

- Plan Free: 100 emails/día
- Solución temporal: Esperar 24 horas o upgrade a plan Pro
- Verificar logs por si hay envíos duplicados

### Emails no llegan

1. Verificar logs del servicio
2. Revisar "Logs" en Resend dashboard
3. Verificar carpeta de spam
4. En producción: Verificar que el dominio está verificado

## 🔗 Referencias

- [Resend Documentation](https://resend.com/docs)
- [Resend API Reference](https://resend.com/docs/api-reference)
- [Resend Go SDK](https://github.com/resendlabs/resend-go)

## 📞 Soporte

- Documentación Resend: [https://resend.com/docs](https://resend.com/docs)
- Soporte Resend: [support@resend.com](mailto:support@resend.com)
- Issues del proyecto: Consultar con el equipo de desarrollo

