-- Script de Inicialización de Base de Datos: El Campeón Web
-- Crea todas las tablas necesarias para el sistema

-- Tabla: users
CREATE TABLE IF NOT EXISTS users (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  address VARCHAR(255),
  city VARCHAR(100),
  postal_code VARCHAR(20),
  country VARCHAR(100),
  role ENUM('USER', 'ADMIN') NOT NULL DEFAULT 'USER',
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  is_bulk_buyer BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,

  KEY idx_role (role),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabla: products
CREATE TABLE IF NOT EXISTS products (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  sku VARCHAR(255) NOT NULL UNIQUE KEY,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(255),
  category VARCHAR(100) NOT NULL,
  price_retail DECIMAL(10, 2) NOT NULL,
  price_wholesale DECIMAL(10, 2) NOT NULL,
  stock INT NOT NULL DEFAULT 0,
  min_bulk_quantity INT NOT NULL DEFAULT 10,
  image_url VARCHAR(500),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  has_variants BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,

  KEY idx_sku (sku),
  KEY idx_category (category),
  KEY idx_is_active (is_active),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabla: product_variants (tipos de variantes como Color, Tamaño, Material, etc.)
CREATE TABLE IF NOT EXISTS product_variants
(
    id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id INT UNSIGNED NOT NULL,
    name       VARCHAR(255)    NOT NULL,
    type       VARCHAR(100)    NOT NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP       NULL,

    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    KEY idx_product_id (product_id),
    KEY idx_deleted_at (deleted_at)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;

-- Tabla: product_variant_values (valores específicos como Rojo, Azul, Grande, Pequeño, etc.)
CREATE TABLE IF NOT EXISTS product_variant_values (
                                                      id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
                                                      variant_id INT UNSIGNED NOT NULL,
                                                      value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    KEY idx_variant_id (variant_id),
    KEY idx_deleted_at (deleted_at)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabla: product_variant_combinations (combinaciones específicas de variantes con su propio stock y SKU)
CREATE TABLE IF NOT EXISTS product_variant_combinations
(
    id                  INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id          INT UNSIGNED NOT NULL,
    sku                 VARCHAR(255)    NOT NULL UNIQUE KEY,
    variant_combination JSON            NOT NULL,
    stock               INT             NOT NULL DEFAULT 0,
    price_adjustment    DECIMAL(10, 2)           DEFAULT 0,
    image_url           VARCHAR(500),
    is_active           BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP       NULL,

    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    KEY idx_product_id (product_id),
    KEY idx_sku (sku),
    KEY idx_is_active (is_active),
    KEY idx_deleted_at (deleted_at)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carts (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL UNIQUE KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  KEY idx_user_id (user_id),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabla: cart_items
CREATE TABLE IF NOT EXISTS cart_items (
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

-- Tabla: orders
CREATE TABLE IF NOT EXISTS orders (
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

-- Tabla: order_items
CREATE TABLE IF NOT EXISTS order_items (
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

-- Tabla: payments
CREATE TABLE IF NOT EXISTS payments (
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

-- Datos de Ejemplo

-- Usuario admin (contraseña hasheada con bcrypt: "admin123")
INSERT INTO users (email, first_name, last_name, password, role, is_active) VALUES
('admin@example.com', 'Admin', 'System', '$2a$12$.qOLnDIwjinV1GI8DivLzugmSOTB3GzxjSR2hFyBhmuDYDNHzneFy', 'ADMIN', TRUE)
ON DUPLICATE KEY UPDATE id=id;

-- Usuario regular (contraseña hasheada: "user1234")
INSERT INTO users (email, first_name, last_name, password, phone, address, city, postal_code, country, role, is_active) VALUES
('usuario@example.com', 'Juan', 'Pérez', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36jbMv/u', '+5491123456789', 'Calle Principal 123', 'Buenos Aires', '1425', 'Argentina', 'USER', TRUE)
ON DUPLICATE KEY UPDATE id=id;

-- Usuario mayorista (contraseña hasheada: "bulk1234")
INSERT INTO users (email, first_name, last_name, password, role, is_active, is_bulk_buyer) VALUES
('mayorista@example.com', 'Carlos', 'García', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36jbMv/u', 'USER', TRUE, TRUE)
ON DUPLICATE KEY UPDATE id=id;

-- Productos de ejemplo
INSERT INTO products (sku, name, description, category, price_retail, price_wholesale, stock, min_bulk_quantity) VALUES
('LIB-001', 'Introducción a Go', 'Aprende Go desde cero', 'Libros', 350.00, 280.00, 50, 5),
('LIB-002', 'Clean Code', 'Código limpio y mantenible', 'Libros', 420.00, 340.00, 30, 5),
('TOY-001', 'Robot Educativo', 'Juguete robotizado educativo', 'Juguetes', 1200.00, 950.00, 20, 2),
('TOY-002', 'Cubo Rubik 3x3', 'Clásico cubo mágico', 'Juguetes', 180.00, 140.00, 100, 10),
('ART-001', 'Colores Madera x24', 'Set de colores naturales', 'Arte', 95.00, 75.00, 200, 20),
('ART-002', 'Lienzo Blanco 50x60', 'Lienzo para pintar', 'Arte', 280.00, 220.00, 45, 5)
ON DUPLICATE KEY UPDATE sku=sku;

-- Crear carritos para usuarios
INSERT INTO carts (user_id) VALUES (2), (3)
ON DUPLICATE KEY UPDATE user_id=user_id;

