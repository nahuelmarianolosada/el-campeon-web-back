# Ejemplos de Requests - El Campeón Web

Colección de ejemplos de requests usando `curl` para todos los endpoints de la API.

## Tabla de Contenidos

1. [Autenticación](#autenticación)
2. [Productos](#productos)
3. [Carrito](#carrito)
4. [Órdenes](#órdenes)
5. [Pagos](#pagos)
6. [Variables Útiles](#variables-útiles)

---

## Autenticación

### Registrar Usuario

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "first_name": "Juan",
    "last_name": "Pérez",
    "password": "SecurePassword123!",
    "phone": "+5491123456789",
    "address": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina"
  }'
```

**Respuesta:** `201 Created`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 2,
    "email": "usuario@example.com",
    "first_name": "Juan",
    "last_name": "Pérez",
    "role": "USER",
    "is_active": true,
    "is_bulk_buyer": false
  },
  "expires_in": 86400
}
```

---

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "password": "SecurePassword123!"
  }'
```

**Respuesta:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 2,
    "email": "usuario@example.com",
    "first_name": "Juan"
  },
  "expires_in": 86400
}
```

---

### Renovar Token

```bash
# Guardar en variable
REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

---

## Productos

### Listar Todos los Productos

```bash
curl http://localhost:8080/api/products
```

Con parámetros:
```bash
curl "http://localhost:8080/api/products?limit=10&offset=0"
```

**Respuesta:**
```json
{
  "data": [
    {
      "id": 1,
      "sku": "LIB-001",
      "name": "Introducción a Go",
      "description": "Aprende Go desde cero",
      "category": "Libros",
      "price_retail": 350.00,
      "price_wholesale": 280.00,
      "stock": 50,
      "min_bulk_quantity": 5,
      "image_url": null,
      "is_active": true,
      "created_at": "2024-04-20T15:30:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

---

### Obtener Producto por ID

```bash
curl http://localhost:8080/api/products/1
```

---

### Obtener Producto por SKU

```bash
curl "http://localhost:8080/api/products/sku?sku=LIB-001"
```

---

### Listar por Categoría

```bash
curl "http://localhost:8080/api/products/category/Libros?limit=10"
```

---

### Crear Producto (ADMIN)

```bash
# Guardar token de admin
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "NEW-001",
    "name": "Nuevo Producto",
    "description": "Descripción del producto",
    "category": "Juguetes",
    "price_retail": 500.00,
    "price_wholesale": 400.00,
    "stock": 30,
    "min_bulk_quantity": 3,
    "image_url": "https://cdn.example.com/image.jpg"
  }'
```

---

### Actualizar Producto (ADMIN)

```bash
curl -X PUT http://localhost:8080/api/products/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stock": 45,
    "price_retail": 375.00
  }'
```

---

### Eliminar Producto (ADMIN)

```bash
curl -X DELETE http://localhost:8080/api/products/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## Carrito

### Obtener Carrito

```bash
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl http://localhost:8080/api/cart \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Respuesta:**
```json
{
  "id": 1,
  "user_id": 2,
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "product": {
        "id": 1,
        "sku": "LIB-001",
        "name": "Introducción a Go",
        "price_retail": 350.00
      },
      "quantity": 2,
      "price": 350.00,
      "subtotal": 700.00
    }
  ],
  "total": 700.00
}
```

---

### Agregar al Carrito

```bash
curl -X POST http://localhost:8080/api/cart/items \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

**Respuesta:** `201 Created`
```json
{
  "message": "item added to cart"
}
```

---

### Actualizar Cantidad en Carrito

```bash
curl -X PUT http://localhost:8080/api/cart/items/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 5
  }'
```

---

### Eliminar del Carrito

```bash
curl -X DELETE http://localhost:8080/api/cart/items/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Vaciar Carrito

```bash
curl -X DELETE http://localhost:8080/api/cart \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Obtener Total del Carrito

```bash
curl http://localhost:8080/api/cart/total \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Respuesta:**
```json
{
  "total": 700.00
}
```

---

## Órdenes

### Crear Orden

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipping_address": {
      "street": "Calle Principal 123",
      "city": "Buenos Aires",
      "postal_code": "1425",
      "country": "Argentina"
    },
    "notes": "Enviar antes de las 17hs"
  }'
```

**Respuesta:** `201 Created`
```json
{
  "id": 1,
  "order_number": "ORD-20240427-123456",
  "user_id": 2,
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "product": { /* ... */ },
      "quantity": 2,
      "price": 350.00,
      "subtotal": 700.00
    }
  ],
  "status": "PENDING",
  "subtotal": 700.00,
  "tax": 147.00,
  "total": 847.00,
  "shipping_address": { /* ... */ },
  "notes": "Enviar antes de las 17hs",
  "created_at": "2024-04-27T10:30:00Z"
}
```

---

### Obtener Mis Órdenes

```bash
curl "http://localhost:8080/api/orders/my?limit=10&offset=0" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Obtener Orden por ID

```bash
curl http://localhost:8080/api/orders/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Listar Todas las Órdenes (ADMIN)

```bash
curl "http://localhost:8080/api/orders?limit=20&offset=0" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

### Actualizar Estado de Orden (ADMIN)

```bash
curl -X PUT http://localhost:8080/api/orders/1/status \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "CONFIRMED"
  }'
```

Estados válidos: `PENDING`, `CONFIRMED`, `SHIPPED`, `DELIVERED`, `CANCELLED`

---

## Pagos

### Crear Pago

```bash
curl -X POST http://localhost:8080/api/payments \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "amount": 847.00
  }'
```

**Respuesta:** `201 Created`
```json
{
  "id": 1,
  "transaction_id": "TXN-1714215000000000001",
  "order_id": 1,
  "user_id": 2,
  "amount": 847.00,
  "currency": "ARS",
  "status": "PENDING",
  "payment_method": "MERCADOPAGO",
  "mercadopago_preference_id": "123456789",
  "approved_at": null,
  "created_at": "2024-04-27T10:30:00Z"
}
```

---

### Obtener Mis Pagos

```bash
curl "http://localhost:8080/api/payments/my?limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Obtener Pago por ID

```bash
curl http://localhost:8080/api/payments/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Obtener Pago de Orden

```bash
curl http://localhost:8080/api/payments/order/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Actualizar Estado de Pago (ADMIN)

```bash
curl -X PUT http://localhost:8080/api/payments/1/status \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "APPROVED"
  }'
```

Estados válidos: `PENDING`, `APPROVED`, `REJECTED`, `CANCELLED`, `REFUNDED`

---

### Listar Todos los Pagos (ADMIN)

```bash
curl "http://localhost:8080/api/payments?limit=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

## Variables Útiles

### Guardar en Archivo

Crear `requests.sh`:
```bash
#!/bin/bash

# URLs
BASE_URL="http://localhost:8080"

# Tokens (completar después de login)
USER_TOKEN="PUT_YOUR_TOKEN_HERE"
ADMIN_TOKEN="PUT_YOUR_ADMIN_TOKEN_HERE"

# Funciones útiles
function register() {
  curl -X POST $BASE_URL/auth/register \
    -H "Content-Type: application/json" \
    -d "$1"
}

function login() {
  curl -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d "$1"
}

function get_products() {
  curl "$BASE_URL/api/products?limit=${1:-20}&offset=${2:-0}"
}

# Usar: bash requests.sh
```

### Prettify JSON

```bash
# Instalar jq (recomendado)
# macOS: brew install jq
# Ubuntu: sudo apt-get install jq
# Windows: choco install jq

# Usar:
curl http://localhost:8080/api/products | jq .
curl http://localhost:8080/api/products | jq '.data[0]'
```

### Guardar Token en Variable

```bash
# Desde respuesta de login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"usuario@example.com","password":"pass"}' \
  | jq -r '.access_token')

echo $TOKEN
```

### Script Completo de Ejemplo

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "1. Registrando usuario..."
REGISTER=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "first_name": "Demo",
    "last_name": "User",
    "password": "DemoPassword123!"
  }')

TOKEN=$(echo $REGISTER | jq -r '.access_token')
echo "Token: $TOKEN"

echo -e "\n2. Listando productos..."
curl -s $BASE_URL/api/products | jq '.data[] | {id, name, price_retail}'

echo -e "\n3. Agregando al carrito..."
curl -s -X POST $BASE_URL/api/cart/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "quantity": 2}' | jq .

echo -e "\n4. Obteniendo carrito..."
curl -s $BASE_URL/api/cart \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n✅ Ejemplo completado!"
```

---

## Postman Collection

Importar en Postman:

```json
{
  "info": {
    "name": "El Campeón Web API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "url": "{{base_url}}/auth/register",
            "header": [
              {"key": "Content-Type", "value": "application/json"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"email\": \"user@example.com\", \"first_name\": \"John\", \"last_name\": \"Doe\", \"password\": \"Pass123!\"}"
            }
          }
        }
      ]
    }
  ]
}
```

---

Más ejemplos disponibles en `/examples` o escribe a support@example.com

