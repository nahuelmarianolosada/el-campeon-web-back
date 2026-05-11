# Product Variants - Quick Start Guide

## What Are Product Variants?

Product variants are different versions of the same product. Examples:
- **Pens**: Different colors and stroke widths
- **Books**: Hardcover, Paperback, E-book
- **Clothing**: Different sizes and colors
- **Electronics**: Different storage capacities

## Key Concepts

### 1. ProductVariant
A **type** of variation. Examples: "Color", "Size", "Material"

### 2. ProductVariantValue
A specific **value** for that variation. Examples: "Red", "Large", "Cotton"

### 3. ProductVariantCombination
A specific **combination** of values with its own:
- SKU (e.g., "SHIRT-RED-LARGE")
- Stock level
- Price adjustment
- Optional image

## Architecture

```
Product (Base)
├── Variant: Color
│   ├── Value: Red
│   ├── Value: Blue
│   └── Value: Green
├── Variant: Size
│   ├── Value: Small
│   ├── Value: Medium
│   └── Value: Large
└── VariantCombinations
    ├── SHIRT-RED-SMALL ($100)
    ├── SHIRT-RED-MEDIUM ($100)
    ├── SHIRT-BLUE-LARGE ($105)
    └── ... (one for each combination)
```

## Quick Setup: 3 Steps

### Step 1: Create Base Product
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SHIRT",
    "name": "Cotton T-Shirt",
    "category": "Clothing",
    "price_retail": 100.00,
    "price_wholesale": 75.00,
    "stock": 0
  }'
# Response: Product with ID = 1
```

### Step 2: Add Variant Types
```bash
# Add Size variant
curl -X POST http://localhost:8080/api/products/1/variants \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Size",
    "type": "size",
    "values": ["Small", "Medium", "Large", "XL"]
  }'

# Add Color variant
curl -X POST http://localhost:8080/api/products/1/variants \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Color",
    "type": "color",
    "values": ["Red", "Blue", "Black"]
  }'
```

### Step 3: Create Combinations with Inventory
```bash
# Red Small T-Shirt
curl -X POST http://localhost:8080/api/products/1/variant-combinations \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SHIRT-RED-S",
    "variant_combination": {
      "Size": "Small",
      "Color": "Red"
    },
    "stock": 50,
    "price_adjustment": 0.00
  }'

# Blue Large T-Shirt (Premium price)
curl -X POST http://localhost:8080/api/products/1/variant-combinations \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SHIRT-BLUE-XL",
    "variant_combination": {
      "Size": "XL",
      "Color": "Blue"
    },
    "stock": 30,
    "price_adjustment": 5.00
  }'
```

## Retrieving Variants

### View All Variants for Product
```bash
curl http://localhost:8080/api/products/1/variants

# Response:
[
  {
    "id": 1,
    "name": "Size",
    "type": "size",
    "values": [
      {"value": "Small"},
      {"value": "Medium"},
      {"value": "Large"},
      {"value": "XL"}
    ]
  },
  {
    "id": 2,
    "name": "Color",
    "type": "color",
    "values": [
      {"value": "Red"},
      {"value": "Blue"},
      {"value": "Black"}
    ]
  }
]
```

### View All Available Combinations
```bash
curl "http://localhost:8080/api/products/1/variant-combinations?limit=20&offset=0"

# Response:
{
  "data": [
    {
      "id": 1,
      "sku": "SHIRT-RED-S",
      "variant_combination": {"Size": "Small", "Color": "Red"},
      "stock": 50,
      "price_adjustment": 0.00,
      "final_price": 100.00,
      "is_active": true
    },
    {
      "id": 2,
      "sku": "SHIRT-BLUE-XL",
      "variant_combination": {"Size": "XL", "Color": "Blue"},
      "stock": 30,
      "price_adjustment": 5.00,
      "final_price": 105.00,
      "is_active": true
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Get Specific Combination by SKU
```bash
curl "http://localhost:8080/api/variant-combinations/sku?sku=SHIRT-RED-S"

# Response: Single combination object
```

## Managing Inventory

### Update Stock for Combination
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/1 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stock": 75
  }'
```

### Update Price
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/1 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "price_adjustment": 7.50
  }'
```

### Discontinue Combination
```bash
curl -X PUT http://localhost:8080/api/variant-combinations/1 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false
  }'
```

## Price Calculation

`Final Price = Base Product Price + Price Adjustment`

**Example:**
- Base product (SHIRT): $100.00
- Red Small: $100.00 + $0.00 = **$100.00**
- Blue XL: $100.00 + $5.00 = **$105.00**

## Common Use Cases

### Case 1: Book with Formats
```
Book: "Go Programming"
├─ Variant: Format
│  └─ Values: Hardcover, Paperback, E-book
└─ Combinations:
   ├─ BOOK-HC: +$50 = $199.99
   ├─ BOOK-PB: +$0 = $149.99
   └─ BOOK-EB: -$50 = $99.99
```

### Case 2: Pen with Multiple Variants
```
Pen: "Premium Ballpoint"
├─ Variant: Color
│  └─ Values: Red, Blue, Black, Green
├─ Variant: Width
│  └─ Values: 0.5mm, 1.0mm, 1.5mm
└─ Combinations: 4 × 3 = 12 combinations
```

### Case 3: Clothing with Size & Color
```
Shirt: "Cotton T-Shirt"
├─ Variant: Size
│  └─ Values: XS, S, M, L, XL, XXL
├─ Variant: Color
│  └─ Values: Red, Blue, Black, White
└─ Combinations: 6 × 4 = 24 combinations
```

## Response Examples

### Variant Response
```json
{
  "id": 1,
  "name": "Color",
  "type": "color",
  "values": [
    {"value": "Red"},
    {"value": "Blue"},
    {"value": "Black"}
  ]
}
```

### Variant Combination Response
```json
{
  "id": 1,
  "sku": "SHIRT-RED-S",
  "variant_combination": {
    "Size": "Small",
    "Color": "Red"
  },
  "stock": 50,
  "price_adjustment": 0.00,
  "image_url": "https://example.com/shirt-red-s.jpg",
  "final_price": 100.00,
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z"
}
```

## API Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success (GET) |
| 201 | Created (POST) |
| 204 | No Content (DELETE) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (no token) |
| 403 | Forbidden (not admin) |
| 404 | Not Found (resource doesn't exist) |
| 500 | Server Error |

## Error Handling

### Common Errors

```json
{
  "error": "product not found"
}
```

```json
{
  "error": "sku already exists"
}
```

```json
{
  "error": "variant combination not found"
}
```

## Best Practices

✅ **DO:**
- Use unique, descriptive SKUs
- Group related products with variants
- Set realistic stock levels
- Use price adjustments for quality/material differences
- Provide images for different variants

❌ **DON'T:**
- Create multiple products when variants would work
- Leave stock at 0 when variants should be inactive
- Use ambiguous variant names
- Mix different product types in one variant set
- Forget to update inventory after sales

## Next Steps

1. **Read VARIANTS.md** for comprehensive documentation
2. **Review VARIANTS_EXAMPLES.md** for step-by-step examples
3. **Check VARIANTS_IMPLEMENTATION.md** for architecture details
4. **Test the API** using provided curl examples
5. **Integrate with your frontend** to display variants

## Frequently Asked Questions

**Q: Can a product have more than 2 types of variants?**
A: Yes! Create as many variant types as needed for your product.

**Q: What's the difference between Product stock and Combination stock?**
A: Product stock is the sum of all combinations. Combinations track individual variant stock.

**Q: Can I change prices per combination?**
A: Yes, use price_adjustment to add or subtract from the base price.

**Q: How do I discontinue a variant without losing data?**
A: Set is_active to false instead of deleting it.

**Q: Can I have multiple images for the same product?**
A: Yes, each combination can have its own image_url.

---

**Ready to get started?** Check out the examples in VARIANTS_EXAMPLES.md!

