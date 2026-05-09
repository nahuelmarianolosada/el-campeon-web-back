# Integración del Checkout Page Mejorado

## Descripción

Este documento describe cómo integrar el checkout page React mejorado que el usuario proporcionó con el backend actualizado.

## Cambios en el Backend (Ya Realizados)

El backend ha sido actualizado para soportar:

1. ✅ Múltiples métodos de entrega (shipping, pickup-libreria, pickup-jugueteria)
2. ✅ Múltiples métodos de pago (MP_SAVED, MP_INSTALLMENTS, MP_CARD, CASH)
3. ✅ Cambio de MercadoPago Orders a Preferences
4. ✅ Migraciones SQL para nuevos campos
5. ✅ Tests actualizados

## Pasos para Integrar el Checkout Page

### 1. Copiar el Componente React

Guarda el componente checkout page en tu proyecto Next.js/React:

```bash
# Ruta recomendada en tu proyecto Next.js
src/app/(authenticated)/checkout/page.tsx
```

O si usas estructura diferente:
```bash
# Alternativa
pages/checkout.tsx
# o
app/checkout/page.tsx
```

### 2. Asegurar Dependencias Instaladas

El componente usa estas librerías (probablemente ya las tienes):

```bash
npm install next react lucide-react
# o si usas tu librería de UI
npm install @/components/ui/...
```

### 3. Importar Helpers de API

Actualiza el archivo `lib/api.ts` o copia el contenido de `CHECKOUT_API_HELPERS.ts`:

```typescript
// En tu archivo lib/api.ts (o crea uno nuevo)
export async function createOrder(token: string, payload: CreateOrderPayload): Promise<OrderResponse> {
  // ... implementación
}

export async function createPayment(token: string, payload: CreatePaymentPayload): Promise<PaymentResponse> {
  // ... implementación
}
```

### 4. Configurar Variables de Entorno

En tu `.env.local` o `.env.development`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 5. Verificar Contextos

El componente usa:
- `useCart()` - Contexto del carrito
- `useAuth()` - Contexto de autenticación

Asegúrate de que estos contextos estén disponibles.

### 6. Probar Localmente

```bash
npm run dev
# Visita http://localhost:3000/checkout
```

## Flujo de Checkout (Actualizado)

```
┌─────────────────────────────────────────────────────────┐
│              USUARIO EN CHECKOUT PAGE                  │
└─────────────────────────────────────────────────────────┘
                          ↓
        ┌──────────────────────────────────────┐
        │  1. SELECCIONAR MÉTODO DE ENTREGA    │
        │     • Envío a domicilio              │
        │     • Retiro en Librería             │
        │     • Retiro en Juguetería           │
        └──────────────────────────────────────┘
                          ↓
        ┌──────────────────────────────────────┐
        │  2. INGRESAR DIRECCIÓN               │
        │     (si seleccionó envío)            │
        │     o confirmar ubicación            │
        │     (si seleccionó retiro)           │
        └──────────────────────────────────────┘
                          ↓
        ┌──────────────────────────────────────┐
        │  3. SELECCIONAR MÉTODO DE PAGO       │
        │     • MP_SAVED (tarjetas guardadas)  │
        │     • MP_INSTALLMENTS (12 cuotas)    │
        │     • MP_CARD (tarjeta nueva)        │
        │     • CASH (Pago Fácil/Rapipago)     │
        └──────────────────────────────────────┘
                          ↓
        ┌──────────────────────────────────────┐
        │  4. CONFIRMAR PEDIDO                 │
        │     POST /api/orders (crear orden)   │
        │     POST /api/payments (crear pago)  │
        └──────────────────────────────────────┘
                          ↓
                    ¿Método de Pago?
                      /        \
                     /          \
            MercadoPago          Efectivo
            (MP_*)               (CASH)
             /  |  \
            /   |   \
        Redirect  Mostrar
        a MP      Código de Pago
```

## Cambios en el Request Body

### Antes (Sin Estas Mejoras)

```json
{
  "order_id": 1,
  "amount": 100.0
}
```

### Ahora (Con Nuevas Mejoras)

```json
// Para crear la orden:
{
  "shipping_address": {
    "street": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina"
  },
  "delivery_method": "shipping",  // ← NUEVO
  "notes": "Por favor dejar en portería"
}

// Para crear el pago:
{
  "order_id": 1,
  "amount": 1210.00,
  "payment_method": "MP_CARD"  // ← NUEVO
}
```

## Mapeo de Métodos de Pago

El componente React mapea así los métodos:

| Componente | Backend | Descripción |
|------------|---------|-------------|
| `mp-saved` | `MP_SAVED` | Tarjetas guardadas o saldo |
| `mp-installments` | `MP_INSTALLMENTS` | Hasta 12 pagos |
| `mp-card` | `MP_CARD` | Débito o Crédito |
| `cash` | `CASH` | Efectivo (PagoFácil/Rapipago) |

## Mapeo de Métodos de Entrega

| Componente | Backend | Ubicación | Dirección |
|-----------|---------|----------|-----------|
| `shipping` | `shipping` | Domicilio | La especificada por usuario |
| `pickup-libreria` | `pickup-libreria` | Librería | Güemes 901, San Salvador de Jujuy |
| `pickup-jugueteria` | `pickup-jugueteria` | Juguetería | Güemes 1045, San Salvador de Jujuy |

## Ejemplo de Implementación del Handler

Si necesitas personalizar el `handleSubmit`:

```typescript
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault()
  if (!token || !cart) return

  setError("")
  setIsProcessing(true)

  try {
    // 1. Determinar dirección de envío
    const address = deliveryMethod === "shipping" 
      ? shippingAddress 
      : getPickupAddress(deliveryMethod)

    // 2. Crear orden
    const order = await createOrder(token, {
      shipping_address: address,
      delivery_method: deliveryMethod,
      notes: notes || undefined,
    })

    // 3. Crear pago
    const payment = await createPayment(token, {
      order_id: order.id,
      amount: order.total,
      payment_method: paymentMethod as "MP_SAVED" | "MP_INSTALLMENTS" | "MP_CARD" | "CASH",
    })

    // 4. Procesar pago
    if (["MP_SAVED", "MP_INSTALLMENTS", "MP_CARD"].includes(paymentMethod)) {
      // Redirigir a MercadoPago
      if (payment.mercadopago_preference_id) {
        const url = `https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=${payment.mercadopago_preference_id}`
        window.location.href = url
      }
    } else if (paymentMethod === "CASH") {
      // Mostrar código de pago para efectivo
      await clearCart()
      router.push(`/mis-ordenes/${order.id}?payment_code=${payment.transaction_id}`)
    }
  } catch (err) {
    setError(err instanceof Error ? err.message : "Error al procesar el pedido")
  } finally {
    setIsProcessing(false)
  }
}
```

## Testing Manual

### Test 1: Envío a Domicilio + Tarjeta de Crédito

1. Seleccionar "Envío a domicilio"
2. Llenar dirección
3. Seleccionar "Débito o Crédito"
4. Hacer clic en "Confirmar Pedido"
5. Debería redirigir a MercadoPago

### Test 2: Retiro en Librería + Efectivo

1. Seleccionar "Retiro en Librería El Campeón"
2. Seleccionar "Efectivo"
3. Hacer clic en "Confirmar Pedido"
4. Debería mostrar código de pago (sin redirigir a MP)

### Test 3: Retiro en Juguetería + Tarjetas Guardadas

1. Seleccionar "Retiro en Juguetería El Campeón"
2. Seleccionar "Tarjetas guardadas o saldo"
3. Hacer clic en "Confirmar Pedido"
4. Debería redirigir a MercadoPago

## Troubleshooting

### Error: "payment amount does not match order total"

**Causa:** Discrepancia entre subtotal+impuesto enviado y el calculado en backend

**Solución:** El backend calcula automáticamente 21% de IVA. Asegúrate de aplicar el mismo cálculo en el frontend.

### Error: "invalid order status transition"

**Causa:** Estado de orden está siendo modificado por una transición no válida

**Solución:** Revisar `internal/services/order/status/status_machine_service.go` para las transiciones válidas

### MercadoPago redirige a page de error

**Causa:** El preference ID es inválido o está vencido

**Solución:** 
1. Verificar que el token de MercadoPago en el backend es válido
2. Revisar logs del backend para errores de API

### El pago no se confirma

**Causa:** Webhook de MercadoPago no está implementado

**Solución:** El webhook es un placeholder. En producción, implementar:
- Verificación de firma
- Consulta a API de MercadoPago
- Actualización automática de estado

## Performance Considerations

1. **Validación en Frontend:**
   - Validar monto antes de enviar
   - Validar dirección antes de enviar
   - Validar que el carrito no esté vacío

2. **Error Handling:**
   - Mostrar errores claros al usuario
   - Log de errores en backend para debugging

3. **Loading States:**
   - Mostrar indicador de "Procesando..." durante la llamada
   - Deshabilitar el botón durante el procesamiento

## Seguridad

**Importante:** El usuario NO puede cambiar:
- El monto total (se valida en backend)
- El ID de la orden (se genera en backend)
- El usuario de la orden (se obtiene del JWT token)

Todo esto está validado en el backend.

## Próximos Pasos

1. **Webhook de MercadoPago:** Implementar confirmación automática de pagos
2. **Confirmación por Email:** Enviar email después de crear orden
3. **Rastreo de Pedido:** Página para ver estado del pedido
4. **Múltiples Direcciones:** Permitir guardar direcciones favoritas

## Referencias

- `CHECKOUT_INTEGRATION.md` - Documentación técnica completa
- `CHECKOUT_API_HELPERS.ts` - Funciones helper para API
- `API.md` - Documentación de endpoints

