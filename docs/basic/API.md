# Especificación API REST - "El Campeón Web"

## 1. Información General

- **Base URL**: `http://localhost:8080/` (desarrollo) o `https://api.elcampeon.com/` (producción)
- **Content-Type**: `application/json`
- **Autenticación**: JWT Bearer Token en header `Authorization`
- **Versionado**: No versionado inicialmente (v1 implícito)
- **Rate Limiting**: A implementar en futuro

## 2. Autenticación

### 2.1 POST /auth/register

Registra un nuevo usuario en el sistema.

**Request:**
```json
{
  "email": "usuario@example.com",
  "first_name": "Juan",
  "last_name": "Pérez",
  "password": "SecurePassword123!",
  "phone": "+5491123456789",
  "address": "Calle Principal 123",
  "city": "Buenos Aires",
  "postal_code": "1425",
  "country": "Argentina"
}
```

**Response:** `201 Created`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "usuario@example.com",
    "first_name": "Juan",
    "last_name": "Pérez",
    "phone": "+5491123456789",
    "address": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina",
    "role": "USER",
    "is_active": true,
    "is_bulk_buyer": false,
    "created_at": "2024-04-27T10:30:00Z"
  },
  "expires_in": 86400
}
```

**Errores:**
- `400 Bad Request` - Email ya registrado
- `400 Bad Request` - Contraseña muy corta (< 8 caracteres)

---

### 2.2 POST /auth/login

Inicia sesión con email y contraseña.

**Request:**
```json
{
  "email": "usuario@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "usuario@example.com",
    "first_name": "Juan",
    "role": "USER",
    "is_active": true,
    "is_bulk_buyer": false
  },
  "expires_in": 86400
}
```

**Errores:**
- `401 Unauthorized` - Credenciales inválidas
- `401 Unauthorized` - Usuario inactivo

---

### 2.3 POST /auth/refresh

Renova el access token usando el refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { /* ... */ },
  "expires_in": 86400
}
```

**Errores:**
- `401 Unauthorized` - Token inválido o expirado

---

## 3. Productos

### 3.1 GET /api/products

Lista todos los productos activos con paginación.

**Query Parameters:**
- `limit` (int, default=20) - Cantidad de resultados
- `offset` (int, default=0) - Offset para paginación

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": 1,
      "sku": "LIB-001",
      "name": "Libro de Ficción",
      "description": "Un libro fascinante...",
      "category": "Libros",
      "price_retail": 250.00,
      "price_wholesale": 200.00,
      "stock": 50,
      "min_bulk_quantity": 10,
      "image_url": "https://cdn.example.com/libro-1.jpg",
      "is_active": true,
      "created_at": "2024-04-20T15:30:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

---

### 3.2 GET /api/products/:id

Obtiene los detalles de un producto específico.

**Response:** `200 OK`
```json
{
  "id": 1,
  "sku": "LIB-001",
  "name": "Libro de Ficción",
  "description": "Un libro fascinante...",
  "category": "Libros",
  "price_retail": 250.00,
  "price_wholesale": 200.00,
  "stock": 50,
  "min_bulk_quantity": 10,
  "image_url": "https://cdn.example.com/libro-1.jpg",
  "is_active": true,
  "created_at": "2024-04-20T15:30:00Z"
}
```

**Errores:**
- `404 Not Found` - Producto no existe

---

### 3.3 GET /api/products/category/:category

Lista productos por categoría.

**Query Parameters:**
- `limit` (int, default=20)
- `offset` (int, default=0)

**Response:** `200 OK` - Mismo formato que lista de productos

---

### 3.4 POST /api/products (ADMIN)

Crea un nuevo producto.

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Request:**
```json
{
  "sku": "LIB-002",
  "name": "Juguete Educativo",
  "description": "Juguete para niños",
  "category": "Juguetes",
  "price_retail": 150.00,
  "price_wholesale": 120.00,
  "stock": 100,
  "min_bulk_quantity": 5,
  "image_url": "https://cdn.example.com/juguete.jpg"
}
```

**Response:** `201 Created`
```json
{
  "id": 2,
  "sku": "LIB-002",
  "name": "Juguete Educativo",
  "category": "Juguetes",
  "price_retail": 150.00,
  "price_wholesale": 120.00,
  "stock": 100,
  "min_bulk_quantity": 5,
  "is_active": true,
  "created_at": "2024-04-27T10:00:00Z"
}
```

**Errores:**
- `401 Unauthorized` - Token no proporcionado
- `403 Forbidden` - No es admin
- `400 Bad Request` - SKU duplicado

---

### 3.5 PUT /api/products/:id (ADMIN)

Actualiza un producto.

**Request:**
```json
{
  "stock": 80,
  "price_retail": 275.00
}
```

**Response:** `200 OK` - Producto actualizado

---

### 3.6 DELETE /api/products/:id (ADMIN)

Elimina un producto (soft delete).

**Response:** `204 No Content`

---

## 4. Carrito de Compras

### 4.1 GET /api/cart

Obtiene el carrito del usuario autenticado.

**Headers:**
```
Authorization: Bearer <user_token>
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 1,
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "product": {
        "id": 1,
        "sku": "LIB-001",
        "name": "Libro de Ficción",
        "price_retail": 250.00,
        "price_wholesale": 200.00,
        "stock": 50
      },
      "quantity": 2,
      "price": 250.00,
      "subtotal": 500.00
    }
  ],
  "total": 500.00
}
```

---

### 4.2 POST /api/cart/items

Agrega un producto al carrito.

**Request:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

**Response:** `201 Created`
```json
{
  "message": "item added to cart"
}
```

**Errores:**
- `400 Bad Request` - Cantidad mayor al stock disponible
- `404 Not Found` - Producto no existe

---

### 4.3 PUT /api/cart/items/:itemId

Actualiza la cantidad de un item en el carrito.

**Request:**
```json
{
  "quantity": 5
}
```

**Response:** `200 OK`
```json
{
  "message": "cart item updated"
}
```

---

### 4.4 DELETE /api/cart/items/:itemId

Elimina un item del carrito.

**Response:** `204 No Content`

---

### 4.5 DELETE /api/cart

Vacía todo el carrito del usuario.

**Response:** `204 No Content`

---

### 4.6 GET /api/cart/total

Obtiene el total del carrito.

**Response:** `200 OK`
```json
{
  "total": 500.00
}
```

---

## 5. Órdenes

### 5.1 POST /api/orders

Crea una nueva orden a partir del carrito.

**Request:**
```json
{
  "shipping_address": {
    "street": "Calle Principal 123",
    "city": "Buenos Aires",
    "postal_code": "1425",
    "country": "Argentina"
  },
  "notes": "Enviar entre 9-17hs"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "order_number": "ORD-20240427-123456",
  "user_id": 1,
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "product": { /* ... */ },
      "quantity": 2,
      "price": 250.00,
      "subtotal": 500.00
    }
  ],
  "status": "PENDING",
  "subtotal": 500.00,
  "tax": 105.00,
  "total": 605.00,
  "shipping_address": { /* ... */ },
  "notes": "Enviar entre 9-17hs",
  "created_at": "2024-04-27T10:30:00Z"
}
```

**Errores:**
- `400 Bad Request` - Carrito vacío
- `400 Bad Request` - Dirección inválida

---

### 5.2 GET /api/orders/my

Lista las órdenes del usuario.

**Query Parameters:**
- `limit` (int, default=20)
- `offset` (int, default=0)

**Response:** `200 OK`
```json
{
  "data": [ /* Array de órdenes */ ],
  "limit": 20,
  "offset": 0
}
```

---

### 5.3 GET /api/orders/:id

Obtiene detalles de una orden específica.

**Response:** `200 OK` - Detalles de la orden

---

### 5.4 PUT /api/orders/:id/status (ADMIN)

Actualiza el estado de una orden.

**Request:**
```json
{
  "status": "SHIPPED"
}
```

**Response:** `200 OK` - Orden actualizada

**Estados válidos:** `PENDING`, `CONFIRMED`, `SHIPPED`, `DELIVERED`, `CANCELLED`

---

### 5.5 GET /api/orders (ADMIN)

Lista todas las órdenes (solo admin).

**Query Parameters:**
- `limit` (int, default=20)
- `offset` (int, default=0)

---

## 6. Pagos

### 6.1 POST /api/payments

Crea un nuevo pago para una orden.

**Request:**
```json
{
  "order_id": 1,
  "amount": 605.00
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "transaction_id": "TXN-1714215000000000001",
  "order_id": 1,
  "user_id": 1,
  "amount": 605.00,
  "currency": "ARS",
  "status": "PENDING",
  "payment_method": "MERCADOPAGO",
  "mercadopago_preference_id": "123456789",
  "approved_at": null,
  "created_at": "2024-04-27T10:30:00Z"
}
```

**Errores:**
- `400 Bad Request` - Monto no coincide con total de orden
- `400 Bad Request` - Orden cancelada

---

### 6.2 GET /api/payments/my

Lista los pagos del usuario.

**Response:** `200 OK` - Array de pagos

---

### 6.3 GET /api/payments/:id

Obtiene detalles de un pago.

**Response:** `200 OK` - Detalles del pago

---

### 6.4 GET /api/payments/order/:orderId

Obtiene el pago de una orden específica.

**Response:** `200 OK` - Detalles del pago

---

### 6.5 PUT /api/payments/:id/status (ADMIN)

Actualiza el estado de un pago.

**Request:**
```json
{
  "status": "APPROVED"
}
```

**Response:** `200 OK` - Pago actualizado

**Estados válidos:** `PENDING`, `APPROVED`, `REJECTED`, `CANCELLED`, `REFUNDED`

---

### 6.6 GET /api/payments (ADMIN)

Lista todos los pagos.

---

### 6.7 POST /webhooks/mercadopago

Recibe webhooks de MercadoPago.

**Headers:**
```
X-Signature: <signature>
X-Request-Id: <request_id>
```

**Request:**
```json
{
  "id": "12345",
  "type": "payment",
  "action": "payment.created",
  "data": {
    "id": "9876543"
  }
}
```

**Response:** `200 OK`
```json
{
  "message": "webhook processed"
}
```

---

## 7. Health Check

### 7.1 GET /health

Verifica el estado del servicio.

**Response:** `200 OK`
```json
{
  "status": "ok",
  "service": "el-campeon-web"
}
```

---

## 8. Códigos de Estado HTTP

- `200 OK` - Solicitud exitosa
- `201 Created` - Recurso creado exitosamente
- `204 No Content` - Solicitud exitosa sin contenido en respuesta
- `400 Bad Request` - Validación fallida
- `401 Unauthorized` - Token no proporcionado o inválido
- `403 Forbidden` - No tiene permisos (no es admin)
- `404 Not Found` - Recurso no encontrado
- `500 Internal Server Error` - Error del servidor

---

## 9. Manejo de Errores

**Formato de Error Estándar:**
```json
{
  "error": "Descripción del error"
}
```

**Ejemplos:**

```json
// Email ya registrado
{
  "error": "email already registered"
}

// Credenciales inválidas
{
  "error": "invalid credentials"
}

// Acceso denegado
{
  "error": "admin access required"
}

// Carrito vacío
{
  "error": "cart is empty"
}
```

---

## 10. Rol de Precios

### Determinación Automática de Precio

Al agregar un producto al carrito:
- Si `user.is_bulk_buyer == true` Y `quantity >= product.min_bulk_quantity`
  - Aplica `price_wholesale`
- Si no
  - Aplica `price_retail`

El precio se fija en el momento de agregar al carrito y se persiste en `CartItem.price`.

---

## 11. Cálculo de Impuestos

En la creación de órdenes:
```
subtotal = SUM(item.quantity * item.price)
tax = subtotal * 0.21  // IVA 21% para Argentina
total = subtotal + tax
```

---

## 12. Ejemplos de Flujo Completo

### Flujo de Compra Típico

```
1. POST /auth/login
   → Obtener access_token

2. GET /api/products
   → Explorar catálogo

3. POST /api/cart/items
   → Agregar productos al carrito

4. GET /api/cart
   → Validar carrito

5. POST /api/orders
   → Crear orden (carrito se vacía automáticamente)

6. POST /api/payments
   → Crear pago de la orden

7. Redirigir a MercadoPago preference_url

8. MercadoPago → POST /webhooks/mercadopago
   → Actualizar estado de pago y orden

9. GET /api/payments/:id
   → Verificar estado del pago
```

---

## 13. Consideraciones de Seguridad

- Validar propiedad de recursos (mi carrito, mis órdenes)
- Validar que el amount del pago coincida exactamente
- Nunca exponer información sensible del usuario
- Validar estados de transición de órdenes y pagos
- Rate limiting en endpoints públicos (futuro)
- CORS restringido en producción (futuro)

