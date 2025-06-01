# Colección Postman - Notification Service API v2

Esta colección contiene todos los endpoints disponibles en el microservicio de notificaciones, incluyendo las nuevas funcionalidades de migraciones y el sistema de actions.

## Estructura de la Colección

### 1. Health & Status
- **Health Check**: Verifica que el servicio esté funcionando correctamente

### 2. Database Migrations
- **Get Migration Status**: Consulta el estado del sistema de migraciones
- **Run Database Migrations**: Ejecuta las migraciones pendientes

### 3. Email Notifications
- **Send Welcome Email**: Envía email de bienvenida usando action `WELCOME`
- **Send Email Verification**: Envía email de verificación usando action `EMAIL_VERIFICATION`
- **Send Password Reset**: Envía email de restablecimiento usando action `PASSWORD_RESET`
- **Send Order Confirmation**: Envía confirmación de pedido usando action `ORDER_CONFIRMATION`
- **Send Shipping Notification**: Envía notificación de envío usando action `SHIPPING_NOTIFICATION`
- **Send Async Email Notification**: Envía notificación de forma asíncrona

### 4. Notification Status
- **Get Notification by ID**: Consulta el estado de una notificación específica

### 5. Error Testing
- **Invalid Email Format**: Prueba validación de email
- **Missing Required Fields**: Prueba campos requeridos faltantes
- **Invalid Action**: Prueba actions no válidas
- **Template Not Found**: Prueba template no encontrado

## Variables de Entorno

La colección utiliza las siguientes variables:

- `base_url`: URL base del servicio (default: `http://localhost:8282`)
- `notification_id`: ID de notificación (se establece automáticamente)

## Nuevo Formato de Actions

La API ahora utiliza **actions** semánticas en lugar de `template_id`:

### Formato Anterior (Obsoleto)
```json
{
  "type": "email",
  "template_id": "welcome_email",
  "recipient": "user@email.com",
  "data": {...}
}
```

### Formato Actual
```json
{
  "type": "email",
  "action": "WELCOME",
  "recipient": "user@email.com",
  "data": {...}
}
```

### Actions Disponibles

- `WELCOME`: Email de bienvenida para nuevos usuarios
- `EMAIL_VERIFICATION`: Verificación de dirección de email
- `PASSWORD_RESET`: Restablecimiento de contraseña
- `ORDER_CONFIRMATION`: Confirmación de pedido
- `SHIPPING_NOTIFICATION`: Notificación de envío
- `PAYMENT_REMINDER`: Recordatorio de pago (ejemplo para testing de errores)

## Flujo de Trabajo Recomendado

1. **Verificar Salud del Servicio**:
   ```
   GET /health
   ```

2. **Ejecutar Migraciones** (primera vez):
   ```
   POST /api/v1/migrations
   ```

3. **Enviar Notificaciones**:
   ```
   POST /api/v1/notifications
   ```

4. **Consultar Estado** (si es necesario):
   ```
   GET /api/v1/notifications/{id}
   ```

## Tests Automáticos

Cada request incluye tests automáticos que verifican:

- **Códigos de status HTTP** apropiados
- **Estructura de respuesta** válida
- **Campos requeridos** en la respuesta
- **Tiempo de respuesta** menor a 30 segundos
- **Formato JSON** válido

## Configuración de PostgreSQL

Antes de usar la colección, asegúrate de que PostgreSQL esté ejecutándose:

```bash
# Usando Docker Compose
docker-compose up -d

# O manualmente
psql -h localhost -p 5432 -U notification_user -d notification_db
```

## Ejemplos de Uso

### Envío de Email de Bienvenida
```json
{
  "type": "email",
  "action": "WELCOME",
  "recipient": "usuario@ejemplo.com",
  "data": {
    "name": "Juan Pérez",
    "email": "usuario@ejemplo.com",
    "company": "Mi Empresa"
  },
  "async": false
}
```

### Verificación de Email
```json
{
  "type": "email",
  "action": "EMAIL_VERIFICATION",
  "recipient": "usuario@ejemplo.com",
  "data": {
    "name": "María García",
    "token": "ABC123XYZ789",
    "verification_link": "https://miapp.com/verify?token=ABC123XYZ789",
    "expiry_time": "24 horas"
  },
  "async": false
}
```

### Restablecimiento de Contraseña
```json
{
  "type": "email",
  "action": "PASSWORD_RESET",
  "recipient": "usuario@ejemplo.com",
  "data": {
    "name": "Carlos López",
    "resetLink": "https://miapp.com/reset-password?token=xyz789",
    "expiry_time": "1 hora"
  },
  "async": false
}
```

## Notas Técnicas

- El sistema ahora utiliza **PostgreSQL** para persistencia
- Las notificaciones se almacenan con datos en formato **JSONB**
- El sistema de **migraciones** es inteligente y detecta tablas existentes
- Logging estructurado con **Zap** para debugging
- Validación robusta en múltiples capas
- Soporte para modo **asíncrono** (cola de trabajos)

## Troubleshooting

### Error de Conexión a BD
Si obtienes errores relacionados con la base de datos:
1. Verifica que PostgreSQL esté ejecutándose
2. Ejecuta las migraciones: `POST /api/v1/migrations`
3. Verifica las variables de entorno de conexión

### Template No Encontrado
Si recibes errores de template no encontrado:
1. Verifica que usas una action válida
2. Asegúrate de que las migraciones se ejecutaron correctamente
3. Revisa que los archivos de template existan en `templates/`

### Variables de Template
Los templates utilizan Go templates con datos en formato JSON. Ejemplo:
- Template: `{{.name}}`
- Data: `{"name": "Juan Pérez"}` 