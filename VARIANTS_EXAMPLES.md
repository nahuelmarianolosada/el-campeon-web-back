# Product Variants - Usage Examples

This file contains practical examples of how to use the Product Variants API.

## Setup: Creating a Pen Product with Color and Width Variants

### Step 1: Create Base Product
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "PEN-BASE",
    "name": "Premium Ballpoint Pen",
    "description": "High-quality ballpoint pen available in multiple colors and widths",
    "category": "Writing Instruments",
    "price_retail": 100.00,
    "price_wholesale": 75.00,
    "stock": 0,
    "min_bulk_quantity": 10,
    "image_url": "https://example.com/pen-base.jpg"
  }'
```

**Response**:
```json
{
  "id": 1,
  "sku": "PEN-BASE",
  "name": "Premium Ballpoint Pen",
  "description": "High-quality ballpoint pen available in multiple colors and widths",
  "category": "Writing Instruments",
  "price_retail": 100.00,
  "price_wholesale": 75.00,
  "stock": 0,
  "min_bulk_quantity": 10,
  "image_url": "https://example.com/pen-base.jpg",
  "is_active": true,
  "has_variants": false,
  "created_at": "2025-01-01T00:00:00Z"
}
```

### Step 2: Create Color Variant
```bash
curl -X POST http://localhost:8080/api/products/1/variants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Color",
    "type": "color",
    "values": ["Red", "Blue", "Black", "Green"]
  }'
```

**Response**:
```json
{
  "id": 1,
  "name": "Color",
  "type": "color",
  "values": [
    {"value": "Red"},
    {"value": "Blue"},
    {"value": "Black"},
    {"value": "Green"}
  ]
}
```

### Step 3: Create Stroke Width Variant
```bash
curl -X POST http://localhost:8080/api/products/1/variants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Stroke Width",
    "type": "width",
    "values": ["0.5mm", "1.0mm", "1.5mm"]
  }'
```

### Step 4: Create Variant Combinations

#### Red pen with 0.5mm stroke
```bash
curl -X POST http://localhost:8080/api/products/1/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "PEN-RED-05",
    "variant_combination": {
      "Color": "Red",
      "Stroke Width": "0.5mm"
    },
    "stock": 200,
    "price_adjustment": 0.00,
    "image_url": "https://example.com/pen-red-05.jpg"
  }'
```

#### Blue pen with 1.0mm stroke (premium pricing)
```bash
curl -X POST http://localhost:8080/api/products/1/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "PEN-BLUE-10",
    "variant_combination": {
      "Color": "Blue",
      "Stroke Width": "1.0mm"
    },
    "stock": 150,
    "price_adjustment": 5.00,
    "image_url": "https://example.com/pen-blue-10.jpg"
  }'
```

#### Black pen with 1.5mm stroke (premium pricing)
```bash
curl -X POST http://localhost:8080/api/products/1/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "PEN-BLACK-15",
    "variant_combination": {
      "Color": "Black",
      "Stroke Width": "1.5mm"
    },
    "stock": 100,
    "price_adjustment": 10.00,
    "image_url": "https://example.com/pen-black-15.jpg"
  }'
```

## Retrieving Variant Information

### Get All Available Variants for Product
```bash
curl http://localhost:8080/api/products/1/variants
```

**Response**:
```json
[
  {
    "id": 1,
    "name": "Color",
    "type": "color",
    "values": [
      {"value": "Red"},
      {"value": "Blue"},
      {"value": "Black"},
      {"value": "Green"}
    ]
  },
  {
    "id": 2,
    "name": "Stroke Width",
    "type": "width",
    "values": [
      {"value": "0.5mm"},
      {"value": "1.0mm"},
      {"value": "1.5mm"}
    ]
  }
]
```

### Get All Variant Combinations for Product
```bash
curl "http://localhost:8080/api/products/1/variant-combinations?limit=10&offset=0"
```

**Response**:
```json
{
  "data": [
    {
      "id": 1,
      "sku": "PEN-RED-05",
      "variant_combination": {
        "Color": "Red",
        "Stroke Width": "0.5mm"
      },
      "stock": 200,
      "price_adjustment": 0.00,
      "image_url": "https://example.com/pen-red-05.jpg",
      "final_price": 100.00,
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "sku": "PEN-BLUE-10",
      "variant_combination": {
        "Color": "Blue",
        "Stroke Width": "1.0mm"
      },
      "stock": 150,
      "price_adjustment": 5.00,
      "image_url": "https://example.com/pen-blue-10.jpg",
      "final_price": 105.00,
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 3,
      "sku": "PEN-BLACK-15",
      "variant_combination": {
        "Color": "Black",
        "Stroke Width": "1.5mm"
      },
      "stock": 100,
      "price_adjustment": 10.00,
      "image_url": "https://example.com/pen-black-15.jpg",
      "final_price": 110.00,
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "limit": 10,
  "offset": 0
}
```

### Get Specific Variant Combination by ID
```bash
curl http://localhost:8080/api/variant-combinations/1
```

**Response**:
```json
{
  "id": 1,
  "sku": "PEN-RED-05",
  "variant_combination": {
    "Color": "Red",
    "Stroke Width": "0.5mm"
  },
  "stock": 200,
  "price_adjustment": 0.00,
  "image_url": "https://example.com/pen-red-05.jpg",
  "final_price": 100.00,
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z"
}
```

### Get Variant Combination by SKU
```bash
curl "http://localhost:8080/api/variant-combinations/sku?sku=PEN-BLUE-10"
```

**Response**:
```json
{
  "id": 2,
  "sku": "PEN-BLUE-10",
  "variant_combination": {
    "Color": "Blue",
    "Stroke Width": "1.0mm"
  },
  "stock": 150,
  "price_adjustment": 5.00,
  "image_url": "https://example.com/pen-blue-10.jpg",
  "final_price": 105.00,
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z"
}
```

## Managing Variants

### Update Variant (Add/Remove Values)
```bash
curl -X PUT http://localhost:8080/api/variants/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Pen Color",
    "values": ["Red", "Blue", "Black", "Green", "Yellow", "Purple"]
  }'
```

### Update Variant Combination Stock
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "stock": 150
  }'
```

### Update Variant Combination Price
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/2 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "price_adjustment": 7.50
  }'
```

### Discontinue Variant Combination
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/3 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "is_active": false
  }'
```

### Delete Variant Combination
```bash
curl -X DELETE http://localhost:8080/api/variant-combinations/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response: `204 No Content`

### Delete Entire Variant (All Values and Combinations)
```bash
curl -X DELETE http://localhost:8080/api/variants/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response: `204 No Content`

## Use Case: Book Store with Format Variants

### Create Base Book Product
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "BOOK-GOBASICS",
    "name": "The Go Programming Language Basics",
    "description": "Learn the fundamentals of Go programming",
    "category": "Programming Books",
    "price_retail": 350.00,
    "price_wholesale": 280.00,
    "stock": 0,
    "min_bulk_quantity": 5
  }'
```

### Create Format Variant
```bash
curl -X POST http://localhost:8080/api/products/2/variants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Format",
    "type": "format",
    "values": ["Hardcover", "Paperback", "E-book"]
  }'
```

### Create Hardcover Variant (Premium)
```bash
curl -X POST http://localhost:8080/api/products/2/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "BOOK-GOBASICS-HC",
    "variant_combination": {
      "Format": "Hardcover"
    },
    "stock": 50,
    "price_adjustment": 100.00,
    "image_url": "https://example.com/book-hardcover.jpg"
  }'
```

### Create Paperback Variant
```bash
curl -X POST http://localhost:8080/api/products/2/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "BOOK-GOBASICS-PB",
    "variant_combination": {
      "Format": "Paperback"
    },
    "stock": 100,
    "price_adjustment": 0.00,
    "image_url": "https://example.com/book-paperback.jpg"
  }'
```

### Create E-book Variant (Discount)
```bash
curl -X POST http://localhost:8080/api/products/2/variant-combinations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "sku": "BOOK-GOBASICS-EB",
    "variant_combination": {
      "Format": "E-book"
    },
    "stock": 999,
    "price_adjustment": -50.00,
    "image_url": "https://example.com/book-ebook.jpg"
  }'
```

## Expected Prices

For the book example with base price 350.00:
- Hardcover: 350.00 + 100.00 = 450.00
- Paperback: 350.00 + 0.00 = 350.00
- E-book: 350.00 - 50.00 = 300.00

For the pen example with base price 100.00:
- Red 0.5mm: 100.00 + 0.00 = 100.00
- Blue 1.0mm: 100.00 + 5.00 = 105.00
- Black 1.5mm: 100.00 + 10.00 = 110.00


