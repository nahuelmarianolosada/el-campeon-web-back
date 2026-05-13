-- Migración: Agregar soporte de combinaciones de variantes al carrito
-- Descripción: Agrega la columna product_variant_combination_id a cart_items para permitir
-- que los items del carrito hagan referencia a combinaciones específicas de variantes

ALTER TABLE cart_items
ADD COLUMN product_variant_combination_id INT UNSIGNED NULL AFTER product_id;

-- Agregar foreign key
ALTER TABLE cart_items
ADD CONSTRAINT fk_cart_items_variant_combination
FOREIGN KEY (product_variant_combination_id)
REFERENCES product_variant_combinations(id)
ON DELETE SET NULL;

-- Agregar índice para mejorar queries
ALTER TABLE cart_items
ADD KEY idx_variant_combination_id (product_variant_combination_id);

