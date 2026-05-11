# Base de Datos - "El Campeón Web"

## 1. Diagrama Entidad-Relación (ERD)

```
┌─────────────────┐
��     users       │
├─────────────────┤
│ id (PK)         │
│ email (UQ)      │
│ first_name      │
│ last_name       │
│ password        │
│ phone           │
│ address         │
│ city            │
│ postal_code     │
│ country         │
│ role            │
│ is_active       │
│ is_bulk_buyer   │
│ created_at      │
│ updated_at      │
│ deleted_at      │
└─────────────────┘
        │
        │ (1:N)
        ├─────────────────────┐
        │                     │
        ▼                     ▼
┌─────────────────┐   ┌─────────────────┐
│     carts       │   │    payments     │
├─────────────────┤   ├─────────────────┤
│ id (PK)         │   │ id (PK)         │
│ user_id (FK,UQ) │   │ transaction_id   │
│ created_at      │   │ order_id (FK)   │
│ updated_at      │   │ user_id (FK)    │
│ deleted_at      │   │ amount          │
└─────────────────┘   │ currency        │
        │             │ status          │
        │ (1:N)       │ payment_method  │
        ▼             │ mp_preference_id│
┌──────────────────┐  │ mp_payment_id   │
│   cart_items     │  │ mp_data (JSON)  │
├──────────────────┤  │ approved_at     │
│ id (PK)          │  │ rejected_reason │
│ cart_id (FK)     │  │ created_at      │
│ product_id (FK)  │  │ updated_at      │
│ quantity         │  │ deleted_at      │
│ price            │  └─────────────────┘
│ created_at       │
│ updated_at       │
└──────────────────┘
        │
        └─────────┐
                  │
┌─────────────────┴──────────────────┐
│                                    │
├─────────────────┐   ┌─────────────┴────────┐
│    products     │   │    orders           │
├─────────────────┤   ├─────────────────────┤
│ id (PK)         │   │ id (PK)             │
│ sku (UQ)        │   │ order_number (UQ)   │
│ name            │   │ user_id (FK)        │
│ description     │   │ status              │
│ category        │   │ subtotal            │
│ price_retail    │   │ tax                 │
│ price_wholesale │   │ total               │
│ stock           │   │ shipping_address    │
│ min_bulk_qty    │   │ notes               │
│ image_url       │   │ created_at          │
│ is_active       │   │ updated_at          │
│ created_at      │   │ deleted_at          │
│ updated_at      │   └─────────────────────┘
│ deleted_at      │           │
└─────────────────┘           │ (1:N)
                              ▼
                    ┌─────────────────────┐
                    │   order_items       │
                    ├─────────────────────┤
                    │ id (PK)             │
                    │ order_id (FK)       │
                    │ product_id (FK)     │
                    │ quantity            │
                    │ price               │
                    │ created_at          │
                    └─────────────────────┘
```

## 2. Definición de Tablas

### 2.1 Tabla: users

```sql
CREATE TABLE users (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE KEY,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  address TEXT,
  city VARCHAR(100),
  postal_code VARCHAR(20),
  country VARCHAR(100),
  role ENUM('USER', 'ADMIN') NOT NULL DEFAULT 'USER',
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  is_bulk_buyer BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  KEY idx_email (email),
  KEY idx_role (role),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.2 Tabla: products

```sql
CREATE TABLE products (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  sku VARCHAR(50) NOT NULL UNIQUE KEY,
  name VARCHAR(255) NOT NULL,
  description LONGTEXT,
  category VARCHAR(100) NOT NULL,
  price_retail DECIMAL(10, 2) NOT NULL,
  price_wholesale DECIMAL(10, 2) NOT NULL,
  stock INT NOT NULL DEFAULT 0,
  min_bulk_quantity INT NOT NULL DEFAULT 10,
  image_url VARCHAR(500),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  KEY idx_sku (sku),
  KEY idx_category (category),
  KEY idx_is_active (is_active),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.3 Tabla: carts

```sql
CREATE TABLE carts (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL UNIQUE KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  KEY idx_user_id (user_id),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.4 Tabla: cart_items

```sql
CREATE TABLE cart_items (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  cart_id INT UNSIGNED NOT NULL,
  product_id INT UNSIGNED NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
  KEY idx_cart_id (cart_id),
  KEY idx_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.5 Tabla: orders

```sql
CREATE TABLE orders (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_number VARCHAR(50) NOT NULL UNIQUE KEY,
  user_id INT UNSIGNED NOT NULL,
  status ENUM('PENDING', 'CONFIRMED', 'SHIPPED', 'DELIVERED', 'CANCELLED') NOT NULL DEFAULT 'PENDING',
  subtotal DECIMAL(12, 2) NOT NULL,
  tax DECIMAL(10, 2) NOT NULL DEFAULT 0,
  total DECIMAL(12, 2) NOT NULL,
  shipping_address JSON,
  notes LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
  KEY idx_order_number (order_number),
  KEY idx_user_id (user_id),
  KEY idx_status (status),
  KEY idx_created_at (created_at),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.6 Tabla: order_items

```sql
CREATE TABLE order_items (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id INT UNSIGNED NOT NULL,
  product_id INT UNSIGNED NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
  KEY idx_order_id (order_id),
  KEY idx_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.7 Tabla: payments

```sql
CREATE TABLE payments (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  transaction_id VARCHAR(100) NOT NULL UNIQUE KEY,
  order_id INT UNSIGNED NOT NULL,
  user_id INT UNSIGNED NOT NULL,
  amount DECIMAL(12, 2) NOT NULL,
  currency VARCHAR(3) NOT NULL DEFAULT 'ARS',
  status ENUM('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED', 'REFUNDED') NOT NULL DEFAULT 'PENDING',
  payment_method ENUM('CREDIT_CARD', 'DEBIT_CARD', 'BANK_TRANSFER', 'MERCADOPAGO') NOT NULL DEFAULT 'MERCADOPAGO',
  mercadopago_preference_id VARCHAR(255),
  mercadopago_payment_id VARCHAR(255),
  mercadopago_data JSON,
  approved_at TIMESTAMP NULL,
  rejected_reason LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
  KEY idx_transaction_id (transaction_id),
  KEY idx_order_id (order_id),
  KEY idx_user_id (user_id),
  KEY idx_status (status),
  KEY idx_created_at (created_at),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## 3. Tipos de Datos

### Campos Comunes
- `TIMESTAMP DEFAULT CURRENT_TIMESTAMP` - Para created_at
- `TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` - Para updated_at
- `TIMESTAMP NULL` - Para deleted_at (soft deletes)

### Moneda
- `DECIMAL(12, 2)` - Para totales (hasta 9,999,999.99)
- `DECIMAL(10, 2)` - Para precios e items

### JSON
- `shipping_address` - Dirección de envío como JSON
- `mercadopago_data` - Respuesta completa de MercadoPago

## 4. Índices de Performance

### Índices por Tabla
- **users**: email, role, deleted_at
- **products**: sku, category, is_active, deleted_at
- **carts**: user_id, deleted_at
- **cart_items**: cart_id, product_id
- **orders**: order_number, user_id, status, created_at, deleted_at
- **order_items**: order_id, product_id
- **payments**: transaction_id, order_id, user_id, status, created_at, deleted_at

## 5. Constraints y Relaciones

### Foreign Keys
- `cart.user_id` → `users.id` (CASCADE on DELETE)
- `cart_items.cart_id` → `carts.id` (CASCADE on DELETE)
- `cart_items.product_id` → `products.id` (RESTRICT on DELETE)
- `orders.user_id` → `users.id` (RESTRICT on DELETE)
- `order_items.order_id` → `orders.id` (CASCADE on DELETE)
- `order_items.product_id` → `products.id` (RESTRICT on DELETE)
- `payments.order_id` → `orders.id` (RESTRICT on DELETE)
- `payments.user_id` → `users.id` (RESTRICT on DELETE)

### Unique Constraints
- `users.email` - Un email por usuario
- `products.sku` - Un SKU por producto
- `carts.user_id` - Un carrito por usuario
- `orders.order_number` - Número de orden único
- `payments.transaction_id` - ID de transacción único

## 6. Enums

```sql
-- users.role
ENUM('USER', 'ADMIN')

-- orders.status
ENUM('PENDING', 'CONFIRMED', 'SHIPPED', 'DELIVERED', 'CANCELLED')

-- payments.status
ENUM('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED', 'REFUNDED')

-- payments.payment_method
ENUM('CREDIT_CARD', 'DEBIT_CARD', 'BANK_TRANSFER', 'MERCADOPAGO')
```

## 7. Ejemplo de Query de Reporte

```sql
-- Órdenes con sus items y usuario
SELECT 
  o.id,
  o.order_number,
  u.email,
  GROUP_CONCAT(p.name SEPARATOR ',') as productos,
  SUM(oi.quantity) as cantidad_items,
  o.total,
  o.status,
  o.created_at
FROM orders o
JOIN users u ON o.user_id = u.id
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
WHERE o.deleted_at IS NULL
GROUP BY o.id
ORDER BY o.created_at DESC;

-- Productos con bajo stock
SELECT 
  id,
  sku,
  name,
  stock,
  category
FROM products
WHERE stock < 10 AND is_active = TRUE
ORDER BY stock ASC;

-- Ingresos por día
SELECT 
  DATE(created_at) as fecha,
  COUNT(*) as cantidad_ordenes,
  SUM(total) as ingresos
FROM orders
WHERE deleted_at IS NULL AND status IN ('CONFIRMED', 'SHIPPED', 'DELIVERED')
GROUP BY DATE(created_at)
ORDER BY fecha DESC;
```

## 8. Consideraciones de Productividad

### Performance
- Índices en foreign keys
- Índices en campos de búsqueda frecuente
- Particionamiento futuro de órdenes por fecha
- Caché de productos activos

### Mantenimiento
- Backup diario de base de datos
- Monitoreo de tamaño de tabla
- Limpieza de soft-deletes (archivar después de 90 días)
- Análisis periódico de índices

### Auditoría
- Timestamps en todas las tablas
- Soft deletes para recuperación
- Tablas de auditoría (futuro)
- Logging de cambios en usuarios y órdenes

## 9. Script de Inicialización Completo

Ver `migrations/init.sql` para script completo que puede ejecutarse directamente.

