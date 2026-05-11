# Guía de Integración: Nuevo Sistema de Checkout Mejorado

## Descripción General

Se ha mejorado el sistema de pagos del backend para soportar múltiples métodos de entrega y pago. Este documento describe cómo integrar el frontend con estas nuevas capacidades.

## Cambios en el Backend

### 1. Nuevos Campos en el Modelo Order

- **`delivery_method`** (string): Método de entrega seleccionado
  - Valores: `shipping`, `pickup-libreria`, `pickup-jugueteria`
  - Campo requerido en `CreateOrderRequest`

### 2. Nuevos Campos en el Modelo Payment

- **`payment_method`** (string): Método de pago seleccionado
  - Valores: `MP_SAVED`, `MP_INSTALLMENTS`, `MP_CARD`, `CASH`
  - Campo requerido en `CreatePaymentRequest`

### 3. Uso de MercadoPago Preference en lugar de Order

Se cambió la integración de MercadoPago de usar **Orders** a usar **Preferences**. Esto permite:
- Mejor control de métodos de pago disponibles
- Estados de pago más granulares
- Mejor manejo de pagos en efectivo (Pago Fácil, Rapipago)

## Migración de Base de Datos

Se ha creado una migración SQL que actualiza los esquemas de las tablas. Ejecuta:

```sql
-- Ubicación: migrations/002_add_delivery_and_payment_methods.sql
-- Ejecutar esta migración en la base de datos
```

## Flujo de Integración Frontend

### Paso 1: Crear la Orden

**Endpoint:** `POST /api/orders`
**Autenticación:** Requerida (Bearer Token)

**Request Body:**
```json
{
  "shipping_address": {
    "street": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina"
  },
  "delivery_method": "shipping",
  "notes": "Instrucciones especiales opcionales"
}
```

**Nota sobre `delivery_method`:**
- `shipping`: Envío a domicilio con la dirección especificada
- `pickup-libreria`: Retiro en Librería El Campeón (Güemes 901)
- `pickup-jugueteria`: Retiro en Juguetería El Campeón (Güemes 1045)

**Response:**
```json
{
  "id": 1,
  "order_number": "ORD-20250508-123456",
  "user_id": 1,
  "status": "PENDING",
  "subtotal": 1000.00,
  "tax": 210.00,
  "total": 1210.00,
  "shipping_address": {
    "street": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina"
  },
  "delivery_method": "shipping",
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "quantity": 2,
      "price": 500.00,
      "subtotal": 1000.00,
      "product": { ... }
    }
  ],
  "created_at": "2025-05-08T...",
  "updated_at": "2025-05-08T..."
}
```

### Paso 2: Crear el Pago

**Endpoint:** `POST /api/payments`
**Autenticación:** Requerida (Bearer Token)

**Request Body:**
```json
{
  "order_id": 1,
  "amount": 1210.00,
  "payment_method": "MP_CARD"
}
```

**Métodos de Pago Disponibles:**

| Valor | Descripción | Notas |
|-------|-------------|-------|
| `MP_SAVED` | Tarjetas guardadas o saldo MercadoPago | Usuario selecciona tarjeta guardada |
| `MP_INSTALLMENTS` | Hasta 12 pagos sin tarjeta | Financiación con MercadoPago |
| `MP_CARD` | Débito o Crédito | Ingresa datos de tarjeta |
| `CASH` | Efectivo (Pago Fácil/Rapipago) | Pago en local, retiro de código |

**Response:**
```json
{
  "id": 1,
  "transaction_id": "TXN-123456789",
  "order_id": 1,
  "user_id": 1,
  "amount": 1210.00,
  "currency": "ARS",
  "status": "PENDING",
  "payment_method": "MP_CARD",
  "mercadopago_preference_id": "123456789",
  "mercadopago_data": {
    "preference": "..."
  },
  "created_at": "2025-05-08T...",
  "updated_at": "2025-05-08T..."
}
```

## Cambios de Comportamiento

### Pagos con MercadoPago (MP_SAVED, MP_INSTALLMENTS, MP_CARD)

1. Se crea una **Preferences** en MercadoPago
2. El usuario es redirigido al checkout de MercadoPago
3. El estado del pago es `PENDING` hasta que MercadoPago confirme
4. Se recibe webhook con confirmación de pago

### Pagos en Efectivo (CASH)

1. El pago se crea con estado `PENDING`
2. **No se genera preference de MercadoPago**
3. Se proporciona un código de pago al usuario
4. El usuario realiza el pago en Pago Fácil o Rapipago
5. Admin confirma el pago manualmente o por webhook (futuro)

## Pickup Locations

Cuando el usuario selecciona retiro en tienda:

### Librería El Campeón
- **Dirección:** Güemes 901, San Salvador de Jujuy, Jujuy, Argentina
- **Código Postal:** 4600
- **Delivery Method:** `pickup-libreria`

### Juguetería El Campeón
- **Dirección:** Güemes 1045, San Salvador de Jujuy, Jujuy, Argentina
- **Código Postal:** 4600
- **Delivery Method:** `pickup-jugueteria`

**Cuando se selecciona pickup:**
- La dirección se sobrescribe automáticamente con la dirección del local
- No hay costo de envío
- El cliente debe retirarlo en persona

## Manejo de Estados de Pago

### Estados Válidos
```
PENDING   → APPROVED│REJECTED│CANCELLED
APPROVED  → CANCELLED│REFUNDED
REJECTED  → (terminal)
CANCELLED → (terminal)
REFUNDED  → (terminal)
```

### Transiciones de Orden al Confirmar Pago

Cuando un pago pasa a `APPROVED`:
```
Order Status: PENDING → CONFIRMED
```

## API Adicionales

### Obtener Pago por Orden
```
GET /api/payments/order/:orderId
```

### Obtener Mis Pagos
```
GET /api/payments/my?limit=20&offset=0
```

### Admin: Actualizar Estado de Pago
```
PUT /api/payments/:id/status
{
  "status": "APPROVED"
}
```

## Webhook de MercadoPago

**Endpoint:** `POST /webhooks/mercadopago`

Actualmente es un placeholder. En producción, se implementará para:
1. Verificar firma del webhook
2. Consultar estado en API de MercadoPago
3. Actualizar estado de pago automáticamente
4. Confirmar la orden

## Archivos Modificados

- `internal/models/order.go` - Agregado `delivery_method`
- `internal/models/payment.go` - Actualizado `payment_method` enum
- `internal/services/payment/payment_service.go` - Cambio de Order a Preference
- `internal/services/payment/mercadopago_client.go` - Nueva interfaz para Preference
- `internal/services/order/order_service.go` - Soporte para delivery_method
- `migrations/002_add_delivery_and_payment_methods.sql` - Nueva migración

## Archivos de Migración

Ejecutar en este orden:
1. `migrations/init.sql` - Inicialización de esquema (si es la primera vez)
2. `migrations/002_add_delivery_and_payment_methods.sql` - Nuevas columnas

## Testing

Los tests han sido actualizados para reflejar los cambios:
- `internal/services/payment/payment_service_test.go` - Nuevos tests para payment_method
- Incluye test para CASH y para MP_CARD

## Notas Importantes

1. **Estado de Pago Inicial:** 
   - Para MercadoPago: `PENDING` (espera confirmación de webhook)
   - Para Efectivo: `PENDING` (espera confirmación manual)

2. **Transición a CONFIRMED:**
   - La orden pasa a `CONFIRMED` inmediatamente al crear el pago
   - El pago confirma el estado una vez aprobado

3. **Validación de Monto:**
   - Se valida que `payment.amount == order.total`
   - Esto previene discrepancias

4. **Métodos de Pago:**
   - Backend solo soporta 4 métodos
   - MercadoPago Preference configurará qué métodos específicos se muestran

## Próximos Pasos

1. Implementar webhook de MercadoPago para confirmación automática
2. Agregar métodos de pago adicionales según necesidad
3. Implementar reembolsos y cancelaciones
4. Agregar reportes de pagos por método

## Soporte

Para preguntas sobre la integración, revisar:
- `API.md` - Documentación general de API
- `IMPLEMENTATION_SUMMARY.md` - Resumen de implementación

