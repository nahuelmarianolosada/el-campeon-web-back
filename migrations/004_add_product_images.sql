-- Migración: Soporte de múltiples imágenes por producto
-- Descripción: Crea la tabla product_images para permitir varias imágenes por producto
-- y migra las imágenes existentes desde la columna products.image_url

-- Desactivar temporalmente las restricciones de llave foránea para asegurar la compatibilidad de tipos
SET FOREIGN_KEY_CHECKS = 0;

-- Asegurar que products.id sea INT UNSIGNED (base de referencia para todas las FKs de productos)
-- Esto corrige posibles discrepancias si la tabla se creó con versiones antiguas o mediante GORM
ALTER TABLE products MODIFY id INT UNSIGNED AUTO_INCREMENT;

-- Crear la tabla product_images
CREATE TABLE IF NOT EXISTS product_images
(
    id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id    INT UNSIGNED NOT NULL,
    image_url     VARCHAR(500) NOT NULL,
    display_order INT          NOT NULL DEFAULT 0,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP    NULL
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- Agregar índices si no existen
SET @index_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_NAME = 'product_images' AND INDEX_NAME = 'idx_product_id' AND TABLE_SCHEMA = DATABASE());
SET @sql_idx = IF(@index_exists = 0, 'CREATE INDEX idx_product_id ON product_images(product_id)', 'SELECT ''Index idx_product_id already exists''');
PREPARE stmt_idx FROM @sql_idx;
EXECUTE stmt_idx;
DEALLOCATE PREPARE stmt_idx;

SET @index_del_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_NAME = 'product_images' AND INDEX_NAME = 'idx_deleted_at' AND TABLE_SCHEMA = DATABASE());
SET @sql_idx_del = IF(@index_del_exists = 0, 'CREATE INDEX idx_deleted_at ON product_images(deleted_at)', 'SELECT ''Index idx_deleted_at already exists''');
PREPARE stmt_idx_del FROM @sql_idx_del;
EXECUTE stmt_idx_del;
DEALLOCATE PREPARE stmt_idx_del;

-- Agregar llave foránea de forma segura
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE TABLE_NAME = 'product_images' AND CONSTRAINT_NAME = 'fk_product_images_product' AND TABLE_SCHEMA = DATABASE());
SET @sql_fk = IF(@fk_exists = 0,
    'ALTER TABLE product_images ADD CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE',
    'SELECT ''Constraint already exists, skipping'''
);
PREPARE stmt_fk FROM @sql_fk;
EXECUTE stmt_fk;
DEALLOCATE PREPARE stmt_fk;

-- Reactivar restricciones
SET FOREIGN_KEY_CHECKS = 1;

-- Migrar imágenes existentes de products.image_url a la nueva tabla
-- Solo se ejecuta si la columna aún existe (para bases de datos antiguas)
SET @column_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'products' AND COLUMN_NAME = 'image_url' AND TABLE_SCHEMA = DATABASE());

SET @sql_ins = IF(@column_exists > 0,
    'INSERT INTO product_images (product_id, image_url, display_order) SELECT id, image_url, 0 FROM products WHERE image_url IS NOT NULL AND image_url != \'\'',
    'SELECT \'Column image_url does not exist, skipping migration\''
);
PREPARE stmt_ins FROM @sql_ins;
EXECUTE stmt_ins;
DEALLOCATE PREPARE stmt_ins;

-- Eliminar la columna image_url de la tabla products ya que ahora se usan múltiples imágenes
SET @sql_drp = IF(@column_exists > 0,
    'ALTER TABLE products DROP COLUMN image_url',
    'SELECT \'Column image_url does not exist, skipping drop\''
);
PREPARE stmt_drp FROM @sql_drp;
EXECUTE stmt_drp;
DEALLOCATE PREPARE stmt_drp;
