-- Migración 005: Sucursales, stock por sucursal y cálculo de envíos por zona.
-- Fecha: 2026-06-08
--
-- Esta migración es idempotente: puede re-ejecutarse sobre cualquier DB
-- (vacía, ya inicializada con init.sql, o con el script ya aplicado parcialmente)
-- sin efectos colaterales.
--
-- Cambios:
--   1. Tabla `branches` (sucursales) + seed con las dos ya existentes.
--   2. Tabla `product_branch_stock` (stock por sucursal) + backfill desde products.stock.
--   3. Tabla `delivery_zones` + seed con zonas de Jujuy.
--   4. Tabla `delivery_rates` (tarifa por zona × sucursal-origen) + tarifas iniciales.
--   5. Tabla `postal_code_zones` (CP → zona) + seed con CPs principales.
--   6. Columnas nuevas en `orders`: origin_branch_id, delivery_zone_id, shipping_cost.
--   7. ENUM `orders.delivery_method` extendido con 'pickup' genérico (los valores
--      'pickup-libreria' / 'pickup-jugueteria' se mantienen por compatibilidad).

-- ============================================================
-- 1. branches
-- ============================================================
CREATE TABLE IF NOT EXISTS branches (
  id              INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code            VARCHAR(32) NOT NULL,
  name            VARCHAR(120) NOT NULL,
  address         VARCHAR(255) NOT NULL,
  lat             DECIMAL(10,7) NULL,
  lng             DECIMAL(10,7) NULL,
  is_pickup_point BOOLEAN NOT NULL DEFAULT TRUE,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMP NULL,
  UNIQUE KEY uq_branches_code (code),
  KEY idx_branches_is_active (is_active),
  KEY idx_branches_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed: dos sucursales actuales. ON DUPLICATE KEY UPDATE conserva idempotencia.
INSERT INTO branches (code, name, address, is_pickup_point, is_active) VALUES
  ('libreria',   'Librería El Campeón',   'Güemes 901, San Salvador de Jujuy, Jujuy, Argentina',  TRUE, TRUE),
  ('jugueteria', 'Juguetería El Campeón', 'Güemes 1045, San Salvador de Jujuy, Jujuy, Argentina', TRUE, TRUE)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  address = VALUES(address),
  is_pickup_point = VALUES(is_pickup_point);

-- ============================================================
-- 2. product_branch_stock
-- ============================================================
CREATE TABLE IF NOT EXISTS product_branch_stock (
  product_id BIGINT UNSIGNED NOT NULL,
  branch_id  INT UNSIGNED NOT NULL,
  stock      INT NOT NULL DEFAULT 0,
  reserved   INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (product_id, branch_id),
  CONSTRAINT fk_pbs_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
  CONSTRAINT fk_pbs_branch  FOREIGN KEY (branch_id)  REFERENCES branches(id) ON DELETE RESTRICT,
  KEY idx_pbs_branch_stock (branch_id, stock),
  CONSTRAINT chk_pbs_stock_nonneg    CHECK (stock >= 0),
  CONSTRAINT chk_pbs_reserved_nonneg CHECK (reserved >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Backfill: asignar todo el stock existente a la sucursal "libreria".
-- INSERT IGNORE evita duplicar si se re-ejecuta.
INSERT IGNORE INTO product_branch_stock (product_id, branch_id, stock, reserved)
SELECT p.id, b.id, COALESCE(p.stock, 0), 0
FROM products p
CROSS JOIN branches b
WHERE b.code = 'libreria'
  AND p.deleted_at IS NULL;

-- ============================================================
-- 3. delivery_zones
-- ============================================================
CREATE TABLE IF NOT EXISTS delivery_zones (
  id             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name           VARCHAR(120) NOT NULL,
  kind           ENUM('provincial','regional','departamental','barrial') NOT NULL,
  parent_zone_id INT UNSIGNED NULL,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at     TIMESTAMP NULL,
  UNIQUE KEY uq_zones_name_kind (name, kind),
  CONSTRAINT fk_zones_parent FOREIGN KEY (parent_zone_id) REFERENCES delivery_zones(id) ON DELETE SET NULL,
  KEY idx_zones_is_active (is_active),
  KEY idx_zones_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO delivery_zones (name, kind, is_active) VALUES
  ('Capital',                       'departamental', TRUE),
  ('Palpalá',                       'departamental', TRUE),
  ('Perico',                        'departamental', TRUE),
  ('Libertador General San Martín', 'departamental', TRUE),
  ('El Carmen',                     'departamental', TRUE),
  ('Resto de Jujuy',                'provincial',    TRUE)
ON DUPLICATE KEY UPDATE
  is_active = VALUES(is_active);

-- ============================================================
-- 4. delivery_rates
-- ============================================================
CREATE TABLE IF NOT EXISTS delivery_rates (
  id                      INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  zone_id                 INT UNSIGNED NOT NULL,
  origin_branch_id        INT UNSIGNED NOT NULL,
  cost                    DECIMAL(12,2) NOT NULL,
  eta_min_days            INT NOT NULL,
  eta_max_days            INT NOT NULL,
  free_shipping_threshold DECIMAL(12,2) NULL,
  is_active               BOOLEAN NOT NULL DEFAULT TRUE,
  created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at              TIMESTAMP NULL,
  UNIQUE KEY uq_rates_zone_branch (zone_id, origin_branch_id),
  CONSTRAINT fk_rates_zone   FOREIGN KEY (zone_id)          REFERENCES delivery_zones(id) ON DELETE CASCADE,
  CONSTRAINT fk_rates_branch FOREIGN KEY (origin_branch_id) REFERENCES branches(id)        ON DELETE CASCADE,
  KEY idx_rates_is_active (is_active),
  KEY idx_rates_deleted_at (deleted_at),
  CONSTRAINT chk_rates_cost_nonneg CHECK (cost >= 0),
  CONSTRAINT chk_rates_eta_min     CHECK (eta_min_days >= 0),
  CONSTRAINT chk_rates_eta_order   CHECK (eta_max_days >= eta_min_days)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed: tarifa inicial por cada zona y cada sucursal origen.
-- Valores referenciales — el admin los ajusta luego desde la UI.
INSERT INTO delivery_rates
  (zone_id, origin_branch_id, cost, eta_min_days, eta_max_days, free_shipping_threshold, is_active)
SELECT
  z.id,
  b.id,
  CASE z.name
    WHEN 'Capital'                       THEN 1500.00
    WHEN 'Palpalá'                       THEN 2000.00
    WHEN 'Perico'                        THEN 2500.00
    WHEN 'El Carmen'                     THEN 2500.00
    WHEN 'Libertador General San Martín' THEN 4000.00
    WHEN 'Resto de Jujuy'                THEN 5500.00
  END AS cost,
  CASE z.name
    WHEN 'Capital'                       THEN 1
    WHEN 'Palpalá'                       THEN 1
    WHEN 'Perico'                        THEN 2
    WHEN 'El Carmen'                     THEN 2
    WHEN 'Libertador General San Martín' THEN 3
    WHEN 'Resto de Jujuy'                THEN 4
  END,
  CASE z.name
    WHEN 'Capital'                       THEN 2
    WHEN 'Palpalá'                       THEN 2
    WHEN 'Perico'                        THEN 3
    WHEN 'El Carmen'                     THEN 3
    WHEN 'Libertador General San Martín' THEN 5
    WHEN 'Resto de Jujuy'                THEN 7
  END,
  50000.00,
  TRUE
FROM delivery_zones z
CROSS JOIN branches b
WHERE z.deleted_at IS NULL AND b.deleted_at IS NULL
ON DUPLICATE KEY UPDATE
  is_active = VALUES(is_active);

-- ============================================================
-- 5. postal_code_zones
-- ============================================================
CREATE TABLE IF NOT EXISTS postal_code_zones (
  postal_code VARCHAR(16) NOT NULL PRIMARY KEY,
  zone_id     INT UNSIGNED NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_pcz_zone FOREIGN KEY (zone_id) REFERENCES delivery_zones(id) ON DELETE RESTRICT,
  KEY idx_pcz_zone (zone_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed: CPs principales de Jujuy (formato CPA — Y4XXX — y formato legacy 4XXX).
INSERT INTO postal_code_zones (postal_code, zone_id)
SELECT cp.postal_code, z.id
FROM (
  SELECT 'Y4600' AS postal_code, 'Capital' AS zone_name                       UNION ALL
  SELECT '4600',                 'Capital'                                    UNION ALL
  SELECT 'Y4612',                'Palpalá'                                    UNION ALL
  SELECT '4612',                 'Palpalá'                                    UNION ALL
  SELECT 'Y4608',                'Perico'                                     UNION ALL
  SELECT '4608',                 'Perico'                                     UNION ALL
  SELECT 'Y4512',                'Libertador General San Martín'              UNION ALL
  SELECT '4512',                 'Libertador General San Martín'              UNION ALL
  SELECT 'Y4603',                'El Carmen'                                  UNION ALL
  SELECT '4603',                 'El Carmen'
) cp
JOIN delivery_zones z ON z.name = cp.zone_name AND z.kind = 'departamental'
ON DUPLICATE KEY UPDATE
  zone_id = VALUES(zone_id);

-- ============================================================
-- 6. Columnas nuevas en orders (idempotente con guarda INFORMATION_SCHEMA)
-- ============================================================

-- 6.1 origin_branch_id
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'origin_branch_id'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE orders
     ADD COLUMN origin_branch_id INT UNSIGNED NULL AFTER delivery_method,
     ADD KEY idx_orders_origin_branch (origin_branch_id),
     ADD CONSTRAINT fk_orders_origin_branch FOREIGN KEY (origin_branch_id) REFERENCES branches(id) ON DELETE RESTRICT',
  'SELECT ''Column orders.origin_branch_id already exists, skipping''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6.2 delivery_zone_id
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'delivery_zone_id'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE orders
     ADD COLUMN delivery_zone_id INT UNSIGNED NULL AFTER origin_branch_id,
     ADD KEY idx_orders_delivery_zone (delivery_zone_id),
     ADD CONSTRAINT fk_orders_delivery_zone FOREIGN KEY (delivery_zone_id) REFERENCES delivery_zones(id) ON DELETE RESTRICT',
  'SELECT ''Column orders.delivery_zone_id already exists, skipping''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6.3 shipping_cost (snapshot al confirmar la orden)
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'shipping_cost'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE orders ADD COLUMN shipping_cost DECIMAL(12,2) NOT NULL DEFAULT 0 AFTER total',
  'SELECT ''Column orders.shipping_cost already exists, skipping''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 7. Extender ENUM delivery_method para soportar 'pickup' genérico
-- ============================================================
-- Aditivo: los valores legacy 'pickup-libreria' y 'pickup-jugueteria' se mantienen
-- para no romper órdenes históricas. El front nuevo envía 'pickup' + origin_branch_id.
SET @needs_update := (
  SELECT CASE WHEN COLUMN_TYPE LIKE '%''pickup''%' THEN 0 ELSE 1 END
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'delivery_method'
);
SET @sql := IF(@needs_update = 1,
  'ALTER TABLE orders MODIFY COLUMN delivery_method ENUM(''shipping'',''pickup'',''pickup-libreria'',''pickup-jugueteria'') NOT NULL DEFAULT ''shipping''',
  'SELECT ''ENUM orders.delivery_method already includes "pickup", skipping''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 8. Backfill: asignar sucursal origen a órdenes existentes
-- ============================================================
-- Regla de inferencia:
--   - Órdenes con delivery_method='pickup-jugueteria' → sucursal 'jugueteria'
--   - Todas las demás → sucursal 'libreria' (por defecto histórico)
UPDATE orders o
LEFT JOIN branches bj ON bj.code = 'jugueteria'
LEFT JOIN branches bl ON bl.code = 'libreria'
SET o.origin_branch_id = CASE
  WHEN o.delivery_method = 'pickup-jugueteria' THEN bj.id
  ELSE bl.id
END
WHERE o.origin_branch_id IS NULL;
