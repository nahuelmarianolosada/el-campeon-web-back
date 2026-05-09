-- Migración: Agregar métodos de entrega y métodos de pago
-- Fecha: 2026-05-08
-- Descripción: Agrega soporte para múltiples métodos de entrega y pago

-- Agregar columna delivery_method a la tabla orders
ALTER TABLE orders
ADD COLUMN delivery_method ENUM('shipping','pickup-libreria','pickup-jugueteria') DEFAULT 'shipping' AFTER shipping_address;

-- Actualizar el ENUM de payment_method en la tabla payments
alter table payments
    drop column payment_method;

alter table payments
    add payment_method ENUM('MP_SAVED','MP_INSTALLMENTS','MP_CARD','CASH') NOT NULL DEFAULT 'MP_CARD';

-- Crear índice para delivery_method para mejorar búsquedas
ALTER TABLE orders
ADD KEY idx_delivery_method (delivery_method);

