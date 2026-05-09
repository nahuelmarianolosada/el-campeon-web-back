#!/bin/bash
# Ejemplos de Llamadas a API - Sistema de Checkout Mejorado
#
# Este archivo contiene ejemplos de llamadas a los nuevos endpoints
# usando curl para testing manual
#
# Uso: bash API_EXAMPLES.sh

# Variables de configuración
API_URL="http://localhost:8080"
TOKEN="tu_jwt_token_aqui"
CONTENT_TYPE="Content-Type: application/json"

# Colores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== EJEMPLOS DE API - SISTEMA DE CHECKOUT MEJORADO =====${NC}\n"

# ==============================================================================
# EJEMPLO 1: Crear Orden - Envío a Domicilio
# ==============================================================================
echo -e "${GREEN}1. Crear Orden - Envío a Domicilio${NC}"
echo "POST /api/orders"
echo ""

curl -X POST "$API_URL/api/orders" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "shipping_address": {
      "street": "Calle Principal 123",
      "city": "Buenos Aires",
      "postal_code": "1425",
      "country": "Argentina"
    },
    "delivery_method": "shipping",
    "notes": "Por favor dejar en portería"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 2: Crear Orden - Retiro en Librería
# ==============================================================================
echo -e "${GREEN}2. Crear Orden - Retiro en Librería${NC}"
echo "POST /api/orders"
echo ""

curl -X POST "$API_URL/api/orders" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "shipping_address": {
      "street": "Güemes 901, San Salvador de Jujuy, Jujuy, Argentina",
      "city": "San Salvador de Jujuy",
      "postal_code": "4600",
      "country": "Argentina"
    },
    "delivery_method": "pickup-libreria",
    "notes": ""
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 3: Crear Orden - Retiro en Juguetería
# ==============================================================================
echo -e "${GREEN}3. Crear Orden - Retiro en Juguetería${NC}"
echo "POST /api/orders"
echo ""

curl -X POST "$API_URL/api/orders" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "shipping_address": {
      "street": "Güemes 1045, San Salvador de Jujuy, Jujuy, Argentina",
      "city": "San Salvador de Jujuy",
      "postal_code": "4600",
      "country": "Argentina"
    },
    "delivery_method": "pickup-jugueteria",
    "notes": "Retiro después de las 18:00"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 4: Crear Pago - Tarjeta de Crédito
# ==============================================================================
echo -e "${GREEN}4. Crear Pago - Tarjeta de Crédito (MP_CARD)${NC}"
echo "POST /api/payments"
echo ""

curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "MP_CARD"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 5: Crear Pago - Tarjetas Guardadas
# ==============================================================================
echo -e "${GREEN}5. Crear Pago - Tarjetas Guardadas (MP_SAVED)${NC}"
echo "POST /api/payments"
echo ""

curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "MP_SAVED"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 6: Crear Pago - 12 Cuotas
# ==============================================================================
echo -e "${GREEN}6. Crear Pago - Hasta 12 Cuotas (MP_INSTALLMENTS)${NC}"
echo "POST /api/payments"
echo ""

curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "MP_INSTALLMENTS"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 7: Crear Pago - Efectivo
# ==============================================================================
echo -e "${GREEN}7. Crear Pago - Efectivo (CASH)${NC}"
echo "POST /api/payments"
echo ""

curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "CASH"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 8: Obtener Pago por Orden
# ==============================================================================
echo -e "${GREEN}8. Obtener Pago por ID de Orden${NC}"
echo "GET /api/payments/order/:orderId"
echo ""

curl -X GET "$API_URL/api/payments/order/1" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 9: Obtener Mi Pago
# ==============================================================================
echo -e "${GREEN}9. Obtener Mis Pagos${NC}"
echo "GET /api/payments/my?limit=20&offset=0"
echo ""

curl -X GET "$API_URL/api/payments/my?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 10: Obtener Orden por ID
# ==============================================================================
echo -e "${GREEN}10. Obtener Orden por ID${NC}"
echo "GET /api/orders/:id"
echo ""

curl -X GET "$API_URL/api/orders/1" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 11: Admin - Actualizar Estado de Pago
# ==============================================================================
echo -e "${GREEN}11. Admin - Actualizar Estado de Pago${NC}"
echo "PUT /api/payments/:id/status"
echo ""

curl -X PUT "$API_URL/api/payments/1/status" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "status": "APPROVED"
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 12: Admin - Listar Todos los Pagos
# ==============================================================================
echo -e "${GREEN}12. Admin - Listar Todos los Pagos${NC}"
echo "GET /api/payments?limit=20&offset=0"
echo ""

curl -X GET "$API_URL/api/payments?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# EJEMPLO 13: Webhook de MercadoPago (Simulación)
# ==============================================================================
echo -e "${GREEN}13. Webhook de MercadoPago (Sin Autenticación)${NC}"
echo "POST /webhooks/mercadopago"
echo ""

curl -X POST "$API_URL/webhooks/mercadopago" \
  -H "$CONTENT_TYPE" \
  -d '{
    "id": "webhook123",
    "type": "payment",
    "action": "payment.created",
    "data": {
      "id": "mp_payment_123"
    }
  }' \
  -w "\nStatus: %{http_code}\n\n"

# ==============================================================================
# VALIDACIONES Y ERRORES COMUNES
# ==============================================================================
echo -e "${GREEN}=== EJEMPLOS DE ERRORES ESPERADOS ===${NC}\n"

echo -e "${GREEN}Error: Monto No Coincide${NC}"
curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 500.00,
    "payment_method": "MP_CARD"
  }' \
  -w "\nStatus: %{http_code}\n\n"

echo -e "${GREEN}Error: Payment Method Inválido${NC}"
curl -X POST "$API_URL/api/payments" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "INVALID_METHOD"
  }' \
  -w "\nStatus: %{http_code}\n\n"

echo -e "${GREEN}Error: Delivery Method Inválido${NC}"
curl -X POST "$API_URL/api/orders" \
  -H "$CONTENT_TYPE" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "shipping_address": {
      "street": "Calle Principal 123",
      "city": "Buenos Aires",
      "postal_code": "1425",
      "country": "Argentina"
    },
    "delivery_method": "invalid_delivery",
    "notes": ""
  }' \
  -w "\nStatus: %{http_code}\n\n"

echo -e "${BLUE}===== FIN DE EJEMPLOS =====${NC}\n"

# ==============================================================================
# NOTAS
# ==============================================================================
cat << 'EOF'

NOTAS IMPORTANTES:

1. Reemplazar "tu_jwt_token_aqui" con un token válido obtenido en /auth/login

2. Métodos de Entrega válidos:
   - shipping (envío a domicilio)
   - pickup-libreria (retiro en librería)
   - pickup-jugueteria (retiro en juguetería)

3. Métodos de Pago válidos:
   - MP_SAVED (tarjetas guardadas)
   - MP_INSTALLMENTS (hasta 12 cuotas)
   - MP_CARD (débito o crédito)
   - CASH (efectivo)

4. Estados de Pago válidos:
   - PENDING (pendiente)
   - APPROVED (aprobado)
   - REJECTED (rechazado)
   - CANCELLED (cancelado)
   - REFUNDED (reembolsado)

5. Transiciones de Estado de Pago:
   PENDING → APPROVED, REJECTED, CANCELLED
   APPROVED → CANCELLED, REFUNDED
   Otros → No pueden cambiar (estados terminales)

6. Al obtener un pago para MercadoPago:
   - El campo "mercadopago_preference_id" contiene el ID de preference
   - Redirigir a: https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id={ID}

7. Para pagos en efectivo (CASH):
   - No se genera preference de MercadoPago
   - El transaction_id se usa como código de pago
   - Usuario paga en Pago Fácil o Rapipago

EOF

