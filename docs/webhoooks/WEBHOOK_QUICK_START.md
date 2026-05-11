# Guía Rápida: Webhooks de MercadoPago - Quick Start

## 📌 5 Pasos para Entender la Implementación

### Paso 1: Entender la Validación de Firma (2 min)

**¿Por qué?** Asegurar que el webhook venga realmente de MercadoPago.

**¿Cómo?** 
```go
// 1. Recibir header X-Signature: ts=TIMESTAMP,v1=SIGNATURE_HEX
// 2. Construir string: PAYMENT_ID.TIMESTAMP
// 3. Calcular HMAC-SHA256 con ACCESS_TOKEN
// 4. Comparar con SIGNATURE_HEX (en tiempo constante)

validator := NewWebhookValidator(cfg.MercadopagoPublicKey)
if !validator.ValidateSignature(xSignature, "payment", paymentID, accessToken) {
    return fmt.Errorf("invalid webhook signature")
}
```

### Paso 2: Obtener Detalles del Pago (2 min)

**¿Por qué?** Confirmar que el pago es real consultando la API de MercadoPago.

**¿Cómo?**
```go
// Hace un GET a: https://api.mercadopago.com/v1/payments/{id}
paymentDetails, err := mercadopagoClient.GetPaymentDetails(ctx, paymentID)
// Retorna: {ID, Status, Amount, Currency, ...}
```

### Paso 3: Buscar el Pago Local (1 min)

**¿Por qué?** Encontrar qué pago en nuestra BD corresponde a este webhook.

**¿Cómo?**
```go
// Buscar por mercadopago_payment_id
payment, err := paymentRepo.FindByMercadopagoPaymentID(paymentID)
```

### Paso 4: Validar y Actualizar (2 min)

**¿Por qué?** Asegurar coherencia y actualizar estados.

**¿Cómo?**
```go
// 1. Verificar que montos coincidan
if paymentDetails.TransactionAmount != payment.Amount {
    return fmt.Errorf("amount mismatch")
}

// 2. Mapear estado de MP a local
status := mapMercadopagoStatusToLocalStatus(paymentDetails.Status)

// 3. Actualizar en BD
payment.Status = status
payment.MercadopagoPaymentID = fmt.Sprintf("%d", paymentDetails.ID)
paymentRepo.Update(payment)
```

### Paso 5: Actualizar Orden Relacionada (1 min)

**¿Por qué?** Mantener sincronizado el estado de la orden.

**¿Cómo?**
```go
// Si pago fue aprobado → orden confirmada
if status == "APPROVED" {
    order.Status = "CONFIRMED"
}
// Si fue rechazado → orden cancelada
if status == "REJECTED" {
    order.Status = "CANCELLED"
}
orderRepo.Update(order)
```

---

## 🎯 Flujo Completo en 1 Minuto

```
┌─────────────────────────────────────────────────────────┐
│ MercadoPago envía webhook a /webhooks/mercadopago       │
├─────────────────────────────────────────────────────────┤
│ Handler recibe: {type: "payment", data: {id: "123"}}    │
│                 Header: X-Signature: ts=...,v1=...      │
├─────────────────────────────────────────────────────────┤
│ 1. Validar firma HMAC-SHA256 ✓                          │
├─────────────────────────────────────────────────────────┤
│ 2. Consultar API: GET /v1/payments/123 ✓                │
│    Respuesta: {status: "approved", amount: 100.00}      │
├─────────────────────────────────────────────────────────┤
│ 3. Buscar pago: payments.mercadopago_payment_id = "123" │
├─────────────────────────────────────────────────────────┤
│ 4. Validar montos coinciden ✓                           │
├─────────────────────────────────────────────────────────┤
│ 5. Actualizar:                                           │
│    - payments.status = "APPROVED"                       │
│    - orders.status = "CONFIRMED"                        │
├─────────────────────────────────────────────────────────┤
│ Retornar HTTP 200 OK                                    │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 Archivos Clave

| Archivo | Propósito |
|---------|-----------|
| `webhook_validator.go` | Validar firmas HMAC-SHA256 |
| `mercadopago_client.go` | Llamar API de MercadoPago |
| `payment_service.go` | Orquestar el flujo |
| `payment_handler.go` | Endpoint HTTP |
| `payment_repository.go` | Acceso a BD |

---

## 🚀 Para Empezar

### 1. Configurar Credenciales
```bash
export MERCADOPAGO_ACCESS_TOKEN="APP_USR_XXXX..."
export MERCADOPAGO_PUBLIC_KEY="APP_USR_YYYY..."
```

### 2. Configurar Webhook en MercadoPago
```
URL: https://tunegocio.com/webhooks/mercadopago
Tipo: POST
Eventos: payment (aprobado, rechazado, refundado)
```

### 3. Probar Localmente
```bash
# En una terminal
go run cmd/main.go

# En otra terminal
curl -X POST http://localhost:8080/webhooks/mercadopago \
  -H "X-Signature: ts=123456,v1=signature" \
  -H "Content-Type: application/json" \
  -d '{"type":"payment","data":{"id":"12345"}}'
```

---

## ❓ Preguntas Frecuentes

### P: ¿Qué pasa si la firma es inválida?
R: Se retorna 401 Unauthorized y se registra en logs sin procesar nada.

### P: ¿Qué pasa si MercadoPago API no responde?
R: Se retorna 500 Internal Server Error. MercadoPago reintentará en 5 min.

### P: ¿Qué pasa si el pago no existe en nuestra BD?
R: Se retorna 500 Internal Server Error. Auditoría manual requerida.

### P: ¿Qué pasa si recibimos el mismo webhook dos veces?
R: Es seguro, simplemente actualiza con los mismos valores. Es idempotente.

### P: ¿Necesito llamar a la API de MercadoPago?
R: Sí, es esencial para confirmar que el pago es real y obtener todos los detalles.

---

## 📊 Estados Posibles

```
Webhook: "approved" → BD: "APPROVED"     → Orden: "CONFIRMED"
Webhook: "rejected" → BD: "REJECTED"     → Orden: "CANCELLED"
Webhook: "refunded" → BD: "REFUNDED"     → Orden: "CANCELLED"
Webhook: "pending"  → BD: "PENDING"      → Orden: "PENDING"
Webhook: "charged_back" → BD: "REJECTED" → Orden: "CANCELLED"
```

---

## 🐛 Debugging

### Ver logs de webhook
```bash
tail -f logs/webhook.log
```

### Verificar pago en BD
```sql
SELECT id, status, mercadopago_payment_id, mercadopago_data 
FROM payments 
WHERE mercadopago_payment_id = '1346501131';
```

### Verificar orden asociada
```sql
SELECT id, status, updated_at 
FROM orders 
WHERE id = (
    SELECT order_id FROM payments 
    WHERE mercadopago_payment_id = '1346501131'
);
```

---

## 📈 Cuando Algo No Funciona

### Checklist de Debugging

- [ ] ¿Webhook se recibe en el endpoint? (Revisar logs del servidor)
- [ ] ¿Firma es válida? (Checar `X-Signature` header)
- [ ] ¿Payment ID existe en MercadoPago? (Verificar en MP dashboard)
- [ ] ¿Pago existe en BD? (SQL query)
- [ ] ¿Montos coinciden? (Comparar BD vs API response)
- [ ] ¿Base de datos está disponible? (Revisar conexión)
- [ ] ¿Credenciales de MP son correctas? (Probar en curl)

---

## 💡 Tips de Producción

1. **Monitorea excepciones**: Usa APM (DataDog, New Relic)
2. **Alertas**: Si fallan más de 5 webhooks en 1 hora
3. **Reintentos**: MercadoPago reintenta automáticamente
4. **Idempotencia**: El webhook puede llegar 2+ veces
5. **Timeout**: Pone timeout de 30s en API calls
6. **Logs**: Guarda todo para auditoría

---

## 📚 Documentación Completa

- `MERCADOPAGO_WEBHOOK_IMPLEMENTATION.md` - Visión global
- `WEBHOOK_TECHNICAL_GUIDE.md` - Detalles técnicos
- Código: `internal/services/payment/webhook_validator.go`
- Tests: `internal/services/payment/webhook_validator_test.go`

---

## ✅ Checklist de Implementación

- [x] Validación de firma HMAC-SHA256
- [x] Consulta de detalles a MP API
- [x] Búsqueda de pago local
- [x] Validación de montos
- [x] Actualización de pago
- [x] Actualización de orden
- [x] Manejo de errores
- [x] Pruebas unitarias
- [x] Documentación
- [x] Compilación exitosa

