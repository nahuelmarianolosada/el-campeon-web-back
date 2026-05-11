# Guía Técnica: Validación y Procesamiento de Webhooks de MercadoPago

## 🔐 Validación de Firma

### Algoritmo HMAC-SHA256

La firma del webhook se valida siguiendo estos pasos:

1. **Parsear el header X-Signature:**
   ```
   X-Signature: ts=1778367737,v1=4d57042cf9734e2b92dc5c336f294a405f45bd0a6b3635af1b431ece26a51f59
   ```

2. **Construir el string a firmar:**
   ```go
   verifyString := fmt.Sprintf("%s.%s", dataID, timestamp)
   // Resultado: "1346501131.1778367737"
   ```

3. **Calcular HMAC-SHA256:**
   ```go
   h := sha256.New()
   h.Write([]byte(verifyString))
   computedSignature := fmt.Sprintf("%x", h.Sum(nil))
   ```

4. **Comparar en tiempo constante:**
   ```go
   constantTimeCompare(computedSignature, receivedSignature)
   ```

### Protección contra Timing Attacks

La comparación se realiza en tiempo constante para evitar que un atacante pueda adivinar la firma byte por byte basándose en el tiempo de respuesta:

```go
func constantTimeCompare(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    
    result := 0
    for i := 0; i < len(a); i++ {
        result |= int(a[i]) ^ int(b[i])
    }
    
    return result == 0
}
```

## 📡 Integración con API de MercadoPago

### Obtención de Detalles del Pago

Después de validar la firma, se obtienen los detalles completos del pago:

```http
GET /v1/payments/1346501131 HTTP/1.1
Host: api.mercadopago.com
Authorization: Bearer ACCESS_TOKEN
X-Idempotency-Key: webhook-payment-1346501131
```

### Respuesta de la API

```json
{
  "id": 1346501131,
  "status": "approved",
  "status_detail": "accredited",
  "transaction_amount": 100.00,
  "currency_id": "ARS",
  "payment_method": {
    "id": "visa",
    "type": "credit_card"
  },
  "date_created": "2026-05-09T23:02:16Z",
  "date_last_modified": "2026-05-09T23:05:30Z",
  "external_reference": "ORDER-12345"
}
```

## 💾 Actualización de Base de Datos

### Tabla de Pagos (payments)

```sql
UPDATE payments SET
  mercadopago_payment_id = '1346501131',
  status = 'APPROVED',
  approved_at = NOW(),
  mercadopago_data = JSON_OBJECT(
    'payment_details', '{"id":1346501131,"status":"approved",...}',
    'webhook_received', '2026-05-09T23:05:40Z'
  ),
  updated_at = NOW()
WHERE id = <payment_id>;
```

### Tabla de Órdenes (orders)

```sql
UPDATE orders SET
  status = 'CONFIRMED',
  updated_at = NOW()
WHERE id = <order_id>;
```

## 🎯 Manejo de Casos Especiales

### 1. Webhook con firma inválida
- Retorn: `401 Unauthorized`
- Mensaje: `"invalid webhook signature"`
- Acción: No se procesa, se registra intento sospechoso

### 2. Pago no encontrado en BD
- Retorno: `500 Internal Server Error`
- Mensaje: `"error finding payment"`
- Acción: Se registra en logs para auditoría

### 3. Monto inconsistente
- Retorno: `500 Internal Server Error`
- Mensaje: `"payment amount mismatch: expected X.XX, got Y.YY"`
- Acción: Alarma de seguridad, se investiga manualmente

### 4. Webhook duplicado
- El sistema es idempotente
- Si se recibe el mismo webhook dos veces, solo actualiza sin error
- Se mantiene `webhook_received` con la última fecha

## 📊 Estados y Transiciones

### Transiciones Válidas

```
PENDING → APPROVED
PENDING → REJECTED  
PENDING → CANCELLED
PENDING → REFUNDED
APPROVED → REFUNDED
APPROVED → CANCELLED

REJECTED ✗ (no más transiciones)
CANCELLED ✗ (no más transiciones)
REFUNDED ✗ (no más transiciones)
```

### Acciones por Estado

| Estado | Orden | Email Cliente | Email Admin |
|:---|:---|:---|:---|
| PENDING | Pending | - | Confirmación recibida |
| APPROVED | Confirmed | Pago confirmado | Preparar |
| REJECTED | Cancelled | Intente nuevamente | Revisión |
| CANCELLED | Cancelled | Cancelada | Revisión |
| REFUNDED | Cancelled | Reembolso procesado | Auditoría |

## 🚨 Monitoreo y Alertas

### Métricas Recomendadas

```go
// Contador de webhooks procesados
webhook_processed_total{status="success|error"}

// Latencia de procesamiento
webhook_processing_duration_seconds

// Validaciones fallidas
webhook_signature_validation_failed_total

// Montos inconsistentes
webhook_amount_mismatch_total

// Pagos no encontrados
webhook_payment_not_found_total
```

### Logs a Monitorear

```
// Éxito
"Successfully processed webhook for payment ID 1346501131, status: APPROVED"

// Firma inválida
"invalid webhook signature"

// Monto inconsistente
"payment amount mismatch: expected 100.00, got 99.99"

// Error inesperado
"error fetching payment details from mercadopago: ..."
```

## 🔄 Reintentos en Caso de Error

En caso de error temporal (ej: BD no disponible):

1. Retornar códigos 5xx (500, 502, 503)
2. MercadoPago reintentará después de:
   - 1er intento: 5 minutos
   - 2do intento: 15 minutos
   - 3er intento: 30 minutos
   - Máximo 3 reintentos

## 📝 Auditoría y Trazabilidad

Cada webhook se registra en el campo `mercadopago_data`:

```json
{
  "payment_details": {
    "id": 1346501131,
    "status": "approved",
    "status_detail": "accredited",
    "transaction_amount": 100.00,
    "payment_method": {
      "id": "visa",
      "type": "credit_card"
    },
    "date_created": "2026-05-09T23:02:16Z"
  },
  "webhook_received": "2026-05-09T23:05:40Z"
}
```

Esto permite:
- Debugging: ver exactamente qué información recibió MP
- Auditoría: rastrear todo cambio de estado
- Reconciliación: comparar datos locales vs MP

## 🧪 Testing en Sandbox

### Configuración

```bash
# .env para desarrollo
MERCADOPAGO_ACCESS_TOKEN=APP_USR_XXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXX
MERCADOPAGO_PUBLIC_KEY=APP_USR_YYYY-YYYYYYYYYYYYYYYYYYYYYYYYYYYY
ENV=development  # Automáticamente usa sandbox de MP
```

### Simular Webhook Aprobado

```bash
curl -X POST http://localhost:8080/webhooks/mercadopago \
  -H "Content-Type: application/json" \
  -H "X-Signature: ts=1778367737,v1=SIGNATURE" \
  -d '{
    "type": "payment",
    "data": {"id": "PAYMENT_ID_FROM_SANDBOX"}
  }'
```

## 🔗 Referencias

- [Documentación oficial de MercadoPago Webhooks](https://www.mercadopago.com.ar/developers/es/docs/checkout-pro/how-tos/notifications-webhooks)
- [SDK Go de MercadoPago](https://github.com/mercadopago/sdk-go)
- [RFC 6234 - Hash Functions](https://tools.ietf.org/html/rfc6234)

