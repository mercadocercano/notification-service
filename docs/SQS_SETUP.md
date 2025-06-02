# Configuración de AWS SQS para Notification Service

Este documento explica cómo configurar AWS SQS (Simple Queue Service) para procesar notificaciones de forma asíncrona en el notification service.

## Prerrequisitos

1. **Cuenta de AWS**: Necesitas una cuenta de AWS activa
2. **AWS CLI configurado** (opcional pero recomendado)
3. **Credenciales de AWS**: Access Key ID y Secret Access Key

## Paso 1: Crear una Cola SQS en AWS

### Opción A: Usando AWS Console

1. Accede a la [Consola de AWS SQS](https://console.aws.amazon.com/sqs/)
2. Haz clic en "Create queue"
3. Configura la cola:
   - **Type**: Standard Queue
   - **Name**: `notification-queue` (o el nombre que prefieras)
   - **Visibility timeout**: 30 seconds (tiempo para procesar un mensaje)
   - **Message retention period**: 14 days
   - **Receive message wait time**: 20 seconds (para long polling)
4. Haz clic en "Create queue"
5. Copia la **Queue URL** que aparece en los detalles

### Opción B: Usando AWS CLI

```bash
# Crear la cola
aws sqs create-queue --queue-name notification-queue --region us-east-1

# Obtener la URL de la cola
aws sqs get-queue-url --queue-name notification-queue --region us-east-1
```

## Paso 2: Configurar Credenciales de AWS

### Opción A: Variables de Entorno

```bash
export AWS_ACCESS_KEY_ID=tu-access-key-id
export AWS_SECRET_ACCESS_KEY=tu-secret-access-key
export AWS_REGION=us-east-1
```

### Opción B: AWS CLI

```bash
aws configure
```

### Opción C: IAM Roles (Recomendado para producción)

Si ejecutas en EC2, usa IAM roles en lugar de credenciales hardcodeadas.

## Paso 3: Configurar el Notification Service

### 1. Actualizar el archivo `.env`

```env
# Configuración de AWS SQS
SQS_ENABLED=true
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue
SQS_REGION=us-east-1

# Credenciales de AWS (si no usas AWS CLI o IAM roles)
AWS_ACCESS_KEY_ID=tu-access-key-id
AWS_SECRET_ACCESS_KEY=tu-secret-access-key
AWS_REGION=us-east-1
```

### 2. Reiniciar el servicio

```bash
go run src/main.go
```

Deberías ver en los logs:
```
Initializing SQS queue with URL: https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue, Region: us-east-1
SQS queue initialized successfully
SQS worker started successfully
```

## Paso 4: Probar la Funcionalidad

### 1. Enviar una notificación asíncrona

```bash
curl --location 'http://localhost:8282/api/v1/notifications' \
--header 'Content-Type: application/json' \
--data-raw '{
    "type": "email",
    "action": "WELCOME",
    "recipient": "usuario@ejemplo.com",
    "data": {
        "name": "Juan Pérez",
        "company": "Mi Empresa",
        "welcome_link": "https://miempresa.com/dashboard"
    },
    "async": true
}'
```

### 2. Verificar el estado de la cola

```bash
curl --location 'http://localhost:8282/api/v1/queue/status'
```

Respuesta esperada:
```json
{
    "queue_size": 0,
    "worker_running": true
}
```

## Formato de Mensaje en SQS

El servicio almacena en SQS el formato de request original (sin campos internos), facilitando el debugging y testing manual:

```json
{
    "type": "email",
    "action": "WELCOME",
    "recipient": "usuario@ejemplo.com",
    "data": {
        "name": "Juan Pérez",
        "company": "Mi Empresa",
        "welcome_link": "https://miempresa.com/dashboard"
    }
}
```

**Nota**: El campo `async` no se incluye en el mensaje SQS ya que todo lo que está en cola es inherentemente asíncrono.

### Envío Manual desde Consola AWS

Puedes enviar mensajes manualmente desde la consola de AWS SQS para testing:

#### Ejemplo WELCOME:
```json
{
    "type": "email",
    "action": "WELCOME",
    "recipient": "test@ejemplo.com",
    "data": {
        "name": "Usuario Test",
        "company": "Mi Empresa",
        "welcome_link": "https://miempresa.com/dashboard"
    }
}
```

#### Ejemplo EMAIL_VERIFICATION:
```json
{
    "type": "email",
    "action": "EMAIL_VERIFICATION",
    "recipient": "test@ejemplo.com",
    "data": {
        "name": "Usuario Test",
        "company": "Mi Empresa",
        "token": "ABC123XYZ789",
        "verification_link": "https://miempresa.com/verify?token=ABC123XYZ789"
    }
}
```

#### Ejemplo PASSWORD_RESET:
```json
{
    "type": "email",
    "action": "PASSWORD_RESET",
    "recipient": "test@ejemplo.com",
    "data": {
        "name": "Usuario Test",
        "reset_token": "DEF456UVW012",
        "reset_link": "https://miempresa.com/reset?token=DEF456UVW012"
    }
}
```

## Flujo de Procesamiento

1. **Cliente envía notificación** con `"async": true`
2. **API encola mensaje** en SQS con estado `queued` (formato request limpio)
3. **Worker SQS escucha** continuamente la cola
4. **Worker procesa mensaje**:
   - Recibe request de SQS
   - Genera nuevo ID interno para tracking
   - Procesa y envía email usando Resend
   - Actualiza estado interno a `sent` o `failed`
   - Elimina mensaje de SQS

## Monitoreo y Troubleshooting

### Logs importantes

- `SQS queue initialized successfully`: SQS configurado correctamente
- `SQS worker started successfully`: Worker iniciado
- `Message sent to SQS successfully`: Mensaje encolado
- `Message dequeued from SQS successfully`: Mensaje procesado
- `Notification sent successfully`: Email enviado

### Problemas comunes

1. **"Could not initialize SQS queue"**
   - Verifica las credenciales de AWS
   - Verifica que la cola existe
   - Verifica la región

2. **"Failed to send message to SQS"**
   - Verifica permisos IAM
   - Verifica conectividad a internet

3. **"Worker not running"**
   - Verifica que `SQS_ENABLED=true`
   - Revisa los logs de inicio

### Comandos útiles de AWS CLI

```bash
# Ver mensajes en la cola
aws sqs receive-message --queue-url https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue

# Ver atributos de la cola
aws sqs get-queue-attributes --queue-url https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue --attribute-names All

# Purgar la cola (eliminar todos los mensajes)
aws sqs purge-queue --queue-url https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue
```

## Permisos IAM Necesarios

Para que el servicio funcione correctamente, necesitas los siguientes permisos:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:SendMessage",
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage",
                "sqs:GetQueueAttributes"
            ],
            "Resource": "arn:aws:sqs:us-east-1:123456789012:notification-queue"
        }
    ]
}
```

## Configuración para Desarrollo Local

Para desarrollo local, puedes usar [LocalStack](https://localstack.cloud/) para simular SQS:

```bash
# Instalar LocalStack
pip install localstack

# Iniciar LocalStack
localstack start

# Crear cola local
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name notification-queue

# Configurar .env para LocalStack
SQS_QUEUE_URL=http://localhost:4566/000000000000/notification-queue
SQS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
```

## Configuración para Producción

### Recomendaciones

1. **Usar IAM Roles** en lugar de credenciales hardcodeadas
2. **Configurar Dead Letter Queue** para mensajes fallidos
3. **Habilitar CloudWatch monitoring**
4. **Configurar Auto Scaling** para el worker si es necesario
5. **Usar KMS encryption** para datos sensibles

### Ejemplo de configuración avanzada

```bash
# Crear Dead Letter Queue
aws sqs create-queue --queue-name notification-dlq

# Configurar redrive policy
aws sqs set-queue-attributes --queue-url https://sqs.us-east-1.amazonaws.com/123456789012/notification-queue --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:123456789012:notification-dlq\",\"maxReceiveCount\":3}"}'
``` 