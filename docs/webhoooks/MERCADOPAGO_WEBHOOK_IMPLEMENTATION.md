# Implementación de Webhooks de MercadoPago

## Descripción General

Esta implementación proporciona un manejo completo y seguro de webhooks de MercadoPago, incluyendo:

1. **Validación de Firma**: Verifica la autenticidad del webhook usando el header `X-Signature`
2. **Obtención de Detalles**: Consulta la API de MercadoPago para obtener información completa del pago
3. **Actualización de Estado**: Actualiza los estados del pago y la orden en la base de datos
4. **Manejo de Errores**: Gestiona diferentes tipos de fallos de manera robusta

## Flujo del Webhook

```
1. MercadoPago envía webhook a /webhooks/mercadopago
   ├─ POST data: { "type": "payment", "action": "payment.created", "data": { "id": "1346501131" } }
   └─ Header: { "X-Signature": "ts=1778367737,v1=..." }

2. El handler recibe el webhook
   ├─ Extrae X-Signature header
   └─ Llama a PaymentService.ProcessMercadopagoWebhook()

3. El servicio valida la firma
   ├─ Parsea X-Signature (ts y v1)
   ├─ Construye el string a firmar: data_id.timestamp
   ├─ Calcula HMAC-SHA256 con access_token
   └─ Compara con la firma recibida

4. Obtiene detalles del pago desde MP API
   ├─ Realiza un GET a /v1/payments/{payment_id}
   ├─ Incluye Bearer token en Authorization header
   └─ Parsea la respuesta JSON

5. Busca el pago local
   ├─ Usa el mercadopago_payment_id de la respuesta
   └─ Valida que los montos coincidan

6. Mapea estado de MP al estado local
   ├─ approved → APPROVED
   ├─ rejected → REJECTED
   ├─ refunded → REFUNDED
   ├─ charged_back → REJECTED
   └─ pending, in_process, in_mediation → PENDING

7. Actualiza la orden según el estado del pago
   ├─ Si APPROVED → CONFIRMED
   ├─ Si REJECTED → CANCELLED
   └─ Si REFUNDED → CANCELLED

8. Retorna 200 OK al webhook
```

## Cambios Realizados

### 1. Modelo de Datos (`internal/models/payment.go`)
- Expandido `MercadopagoWebhookRequest` con todos los campos necesarios
- Agregado `MercadopagoPaymentDetailsResponse` para la respuesta de la API de MP

### 2. Cliente de MercadoPago (`internal/services/payment/mercadopago_client.go`)
- Expandida la interfaz `MercadopagoClient` con `GetPaymentDetails()`
- Implementado método para consultar detalles del pago desde MP API
- Agrega autenticación con Bearer token

### 3. Validador de Webhooks (`internal/services/payment/webhook_validator.go`)
- Nueva clase `WebhookValidator` para validar firmas
- Implementa validación de HMAC-SHA256 en tiempo constante
- Parsea header `X-Signature` correctamente

### 4. Servicio de Pagos (`internal/services/payment/payment_service.go`)
- Actualizado constructor para incluir `WebhookValidator`
- Reemplazado `ProcessMercadopagoWebhook()` con implementación completa
- Agregada función `mapMercadopagoStatusToLocalStatus()` para mapeo de estados

### 5. Repositorio de Pagos (`internal/repositories/payment_repository.go`)
- Agregado método `FindByMercadopagoPaymentID()` para buscar pagos por ID de MP

### 6. Handler de Pagos (`internal/handlers/payment_handler.go`)
- Actualizado `MercadopagoWebhook()` para pasar `X-Signature` al servicio
- Validación de presencia del header

## Configuración Requerida

Asegúrate de tener estas variables de entorno configuradas:

```bash
MERCADOPAGO_ACCESS_TOKEN=your_mp_access_token
MERCADOPAGO_PUBLIC_KEY=your_mp_public_key
```

## Prueba del Webhook

### Ejemplo de Webhook Recibido

```http
POST /webhooks/mercadopago?data.id=1346501131&type=payment HTTP/1.1
Host: localhost:8080
X-Signature: ts=1778367737,v1=4d57042cf9734e2b92dc5c336f294a405f45bd0a6b3635af1b431ece26a51f59
Content-Type: application/json

{
  "action": "payment.created",
  "api_version": "v1",
  "data": {
    "id": "1346501131"
  },
  "date_created": "2026-05-09T23:02:16Z",
  "id": 131777241053,
  "live_mode": false,
  "type": "payment",
  "user_id": "129629411"
}
```

### cURL para Prueba Manual

```bash
curl -X POST http://localhost:8080/webhooks/mercadopago \
  -H "Content-Type: application/json" \
  -H "X-Signature: ts=1778367737,v1=4d57042cf9734e2b92dc5c336f294a405f45bd0a6b3635af1b431ece26a51f59" \
  -d '{
    "action": "payment.created",
    "api_version": "v1",
    "data": {
      "id": "1346501131"
    },
    "date_created": "2026-05-09T23:02:16Z",
    "id": 131777241053,
    "live_mode": false,
    "type": "payment",
    "user_id": "129629411"
  }'
```

## Estados de Pago Soportados

### Estados de MercadoPago
- `approved` - Pago aprobado
- `rejected` - Pago rechazado
- `refunded` - Pago reembolsado
- `charged_back` - Contracargo
- `pending` - Pendiente
- `cancelled` - Cancelado
- `in_process` - En proceso
- `in_mediation` - En mediación

### Estados Locales
- `PENDING` - En espera de confirmación
- `APPROVED` - Aprobado
- `REJECTED` - Rechazado
- `CANCELLED` - Cancelado
- `REFUNDED` - Reembolsado

## Seguridad

1. **Validación de Firma**: Usa HMAC-SHA256 para verificar autenticidad
2. **Tiempo Constante**: Comparación de strings en tiempo constante para evitar timing attacks
3. **Validación de Monto**: Verifica que el monto de MP coincida con el pago local
4. **API Authentication**: Usa Bearer token para consultas a MP API

## Manejo de Errores

La implementación gestiona:
- Webhook con firma inválida (401 Unauthorized)
- Datos JSON malformados (400 Bad Request)
- Pago no encontrado en BD (500 Internal Server Error)
- Monto inconsistente (500 Internal Server Error)
- Actualizaciones fallidas en BD (500 Internal Server Error)

## Monitoreo y Logging

Se implementan logs para debugging:
- Intentos de validación de firma
- Consultas a MP API
- Cambios de estado
- Errores durante el procesamiento

## Próximos Pasos

Para mejorar aún más la implementación:

1. **Reintentos**: Implementar reintentos para fallos transitorios
2. **Queue de Webhooks**: Usar una cola para procesamiento asíncrono
3. **Persistencia de Webhooks**: Guardar eventos de webhook para auditoría
4. **Notificaciones**: Enviar email al cliente cuando hay cambios de estado
5. **Métricas**: Implementar métricas de éxito/error de webhooks

