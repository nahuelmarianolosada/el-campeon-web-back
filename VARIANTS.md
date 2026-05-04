# Product Variants Feature Documentation

## Overview

The Product Variants feature allows you to define multiple variations of a single product, such as different colors, sizes, materials, or other attributes. Each variant combination can have its own SKU, stock level, price adjustment, and image.

### Examples

1. **Pens**: Different colors (Red, Blue, Black) and stroke widths (0.5mm, 1mm)
2. **Books**: Format variants (Hardcover, Paperback)
3. **Clothing**: Size variants (S, M, L, XL) and color variants (Red, Blue, Green)
4. **Products in General**: Any configurable attributes

## Architecture

### Models

#### ProductVariant
Represents a type of variant attribute (e.g., "Color", "Size", "Material").
- **Fields**:
  - `ID`: Unique identifier
  - `ProductID`: Reference to the parent product
  - `Name`: Display name (e.g., "Color")
  - `Type`: System identifier (e.g., "color")
  - `CreatedAt`, `UpdatedAt`, `DeletedAt`: Timestamps

#### ProductVariantValue
Represents a specific value for a variant (e.g., "Red" for Color).
- **Fields**:
  - `ID`: Unique identifier
  - `VariantID`: Reference to the parent variant
  - `Value`: The actual value (e.g., "Red", "Large")
  - `CreatedAt`, `UpdatedAt`, `DeletedAt`: Timestamps

#### ProductVariantCombination
Represents a specific combination of variant values with its own SKU and inventory.
- **Fields**:
  - `ID`: Unique identifier
  - `ProductID`: Reference to the parent product
  - `SKU`: Unique SKU for this combination (e.g., "PEN-RED-THIN")
  - `VariantCombination`: JSON map of variant selections (e.g., `{"Color": "Red", "Width": "0.5mm"}`)
  - `Stock`: Current inventory count for this combination
  - `PriceAdjustment`: Additional price on top of base product price
  - `ImageURL`: Optional specific image for this combination
  - `IsActive`: Whether this combination is available for purchase
  - `CreatedAt`, `UpdatedAt`, `DeletedAt`: Timestamps

### Database Schema

Three new tables are created:

```sql
-- product_variants table
CREATE TABLE product_variants (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id INT UNSIGNED NOT NULL,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(100) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (product_id) REFERENCES products(id)
);

-- product_variant_values table
CREATE TABLE product_variant_values (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  variant_id INT UNSIGNED NOT NULL,
  value VARCHAR(255) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (variant_id) REFERENCES product_variants(id)
);

-- product_variant_combinations table
CREATE TABLE product_variant_combinations (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id INT UNSIGNED NOT NULL,
  sku VARCHAR(255) NOT NULL UNIQUE,
  variant_combination JSON NOT NULL,
  stock INT NOT NULL DEFAULT 0,
  price_adjustment DECIMAL(10, 2) DEFAULT 0,
  image_url VARCHAR(500),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

The `Product` model has been updated to include:
- `HasVariants`: Boolean indicating if the product has variants
- `Variants`: Relationship to ProductVariant records
- `VariantCombinations`: Relationship to ProductVariantCombination records

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Get Product Variants
```
GET /api/products/:productId/variants
```
Returns all available variants for a product.

**Response**:
```json
[
  {
    "id": 1,
    "name": "Color",
    "type": "color",
    "values": [
      { "value": "Red" },
      { "value": "Blue" },
      { "value": "Green" }
    ]
  }
]
```

#### Get Product Variant Combinations
```
GET /api/products/:productId/variant-combinations?limit=20&offset=0
```
Returns all available variant combinations for a product with pagination.

**Response**:
```json
{
  "data": [
    {
      "id": 1,
      "sku": "PEN-RED-THIN",
      "variant_combination": {
        "Color": "Red",
        "Width": "0.5mm"
      },
      "stock": 100,
      "price_adjustment": 10.50,
      "image_url": "https://example.com/pen-red-thin.jpg",
      "final_price": 110.50,
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

#### Get Single Variant
```
GET /api/variants/:variantId
```
Returns a specific variant with all its values.

#### Get Variant Combination by ID
```
GET /api/variant-combinations/:combinationId
```
Returns a specific variant combination.

#### Get Variant Combination by SKU
```
GET /api/variant-combinations/sku?sku=PEN-RED-THIN
```
Retrieves a variant combination using its SKU.

### Admin Endpoints (Requires Authentication + ADMIN Role)

#### Create Product Variant
```
POST /api/products/:productId/variants
```
Creates a new variant type for a product.

**Request**:
```json
{
  "name": "Color",
  "type": "color",
  "values": ["Red", "Blue", "Green"]
}
```

**Response**: `201 Created` with ProductVariantResponse

#### Update Product Variant
```
PUT /api/variants/:variantId
or
PUT /api/products/:productId/variants/:variantId
```
Updates a variant's name and values.

**Request**:
```json
{
  "name": "Product Color",
  "values": ["Red", "Blue", "Green", "Yellow"]
}
```

#### Delete Product Variant
```
DELETE /api/variants/:variantId
or
DELETE /api/products/:productId/variants/:variantId
```
Deletes a variant and all its associated values.

**Response**: `204 No Content`

#### Create Variant Combination
```
POST /api/products/:productId/variant-combinations
```
Creates a new combination of variants.

**Request**:
```json
{
  "sku": "PEN-RED-THIN",
  "variant_combination": {
    "Color": "Red",
    "Width": "0.5mm"
  },
  "stock": 100,
  "price_adjustment": 10.50,
  "image_url": "https://example.com/pen-red-thin.jpg"
}
```

**Response**: `201 Created` with ProductVariantCombinationResponse

#### Update Variant Combination
```
PUT /api/variant-combinations/:combinationId
or
PUT /api/products/:productId/variant-combinations/:combinationId
```
Updates stock, price adjustment, image, or active status.

**Request**:
```json
{
  "stock": 80,
  "price_adjustment": 12.00,
  "image_url": "https://example.com/pen-red-thin-v2.jpg",
  "is_active": true
}
```

#### Delete Variant Combination
```
DELETE /api/variant-combinations/:combinationId
or
DELETE /api/products/:productId/variant-combinations/:combinationId
```
Deletes a variant combination.

**Response**: `204 No Content`

## Workflow Examples

### Example 1: Create a Pen Product with Color and Width Variants

#### Step 1: Create the base product
```
POST /api/products
{
  "sku": "PEN-BASE",
  "name": "Premium Pen",
  "category": "Writing Instruments",
  "price_retail": 100.00,
  "price_wholesale": 80.00,
  "stock": 0,
  "min_bulk_quantity": 10
}
```

#### Step 2: Create variant types
```
POST /api/products/1/variants
{
  "name": "Color",
  "type": "color",
  "values": ["Red", "Blue", "Black"]
}
```

```
POST /api/products/1/variants
{
  "name": "Stroke Width",
  "type": "width",
  "values": ["0.5mm", "1mm", "2mm"]
}
```

#### Step 3: Create variant combinations
```
POST /api/products/1/variant-combinations
{
  "sku": "PEN-RED-THIN",
  "variant_combination": {
    "Color": "Red",
    "Stroke Width": "0.5mm"
  },
  "stock": 100,
  "price_adjustment": 0.00,
  "image_url": "https://example.com/pen-red-thin.jpg"
}
```

```
POST /api/products/1/variant-combinations
{
  "sku": "PEN-BLUE-THICK",
  "variant_combination": {
    "Color": "Blue",
    "Stroke Width": "2mm"
  },
  "stock": 50,
  "price_adjustment": 5.00,
  "image_url": "https://example.com/pen-blue-thick.jpg"
}
```

### Example 2: Create a Book with Format Variants

#### Step 1: Create the base product
```
POST /api/products
{
  "sku": "BOOK-BASE",
  "name": "The Go Programming Language",
  "category": "Books",
  "price_retail": 350.00,
  "price_wholesale": 280.00,
  "stock": 0,
  "min_bulk_quantity": 5
}
```

#### Step 2: Create format variant
```
POST /api/products/2/variants
{
  "name": "Format",
  "type": "format",
  "values": ["Hardcover", "Paperback", "E-book"]
}
```

#### Step 3: Create combinations with price adjustments
```
POST /api/products/2/variant-combinations
{
  "sku": "BOOK-HARDCOVER",
  "variant_combination": {
    "Format": "Hardcover"
  },
  "stock": 100,
  "price_adjustment": 50.00
}
```

```
POST /api/products/2/variant-combinations
{
  "sku": "BOOK-PAPERBACK",
  "variant_combination": {
    "Format": "Paperback"
  },
  "stock": 200,
  "price_adjustment": 0.00
}
```

```
POST /api/products/2/variant-combinations
{
  "sku": "BOOK-EBOOK",
  "variant_combination": {
    "Format": "E-book"
  },
  "stock": 999,
  "price_adjustment": -50.00
}
```

## Integration with Cart and Orders

When adding items to the cart or creating orders, you can reference variant combinations by their SKU:

```
POST /api/cart/items
{
  "product_id": 1,
  "variant_combination_sku": "PEN-RED-THIN",
  "quantity": 5
}
```

Or if using variant combination ID:

```
POST /api/cart/items
{
  "product_id": 1,
  "variant_combination_id": 1,
  "quantity": 5
}
```

**Note**: The cart and order systems will need to be updated to support variant combinations. This is a future enhancement.

## Best Practices

1. **Unique SKUs**: Always use unique SKUs for each variant combination to avoid inventory conflicts.

2. **Price Adjustments**: Use the `price_adjustment` field to account for variations in production cost:
   - Premium materials: Positive adjustment
   - Budget variants: Negative adjustment
   - Standard variants: Zero adjustment

3. **Stock Management**: Track stock separately for each combination to prevent overselling:
   - Adjust stock when orders are created
   - Monitor low stock levels

4. **Images**: Provide specific images for each combination when visually distinct:
   - Color changes
   - Size differences
   - Format variations

5. **Active Status**: Use `is_active` to soft-delete combinations without losing historical data:
   - Discontinue out-of-stock combinations
   - Rotate seasonal variants

## Future Enhancements

1. **Variant Pre-selection**: Allow specifying default variant combinations when viewing products
2. **Variant Filtering**: Add endpoints to filter products by variant values
3. **Bulk Pricing**: Apply different wholesale discounts to variant combinations
4. **Cart Integration**: Full variant selection UI support
5. **Variant Templates**: Reusable variant configurations across products
6. **Inventory Sync**: Automatic stock level adjustments across all combinations

## Migration Notes

If you're migrating an existing database:

1. Run the migration script to create the three new tables
2. Set `has_variants = false` for existing products
3. Create variant entries for products that have variants
4. Update product stock to 0 and use variant combinations for actual inventory

## Error Handling

Common errors and their resolution:

| Error | Cause | Solution |
|-------|-------|----------|
| `product not found` | ProductID doesn't exist | Verify product exists in database |
| `sku already exists` | Duplicate SKU | Use unique SKU for each combination |
| `variant combination not found` | Invalid variant combination ID | Check the combination ID |
| `insufficient stock` | Not enough inventory | Adjust stock level or select different combination |


