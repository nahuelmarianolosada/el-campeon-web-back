# Guest Checkout Implementation

## Overview
This document describes the guest checkout system implemented for el-campeon-web, allowing users to complete purchases without registration or authentication.

## Architecture

### Core Components

#### 1. Database Models
- **guest_sessions** - Temporary sessions for guest users with:
  - `id` (BIGINT AUTO_INCREMENT PRIMARY KEY)
  - `email` (UNIQUE)
  - `verification_code_hash` (bcrypt hashed)
  - `is_verified` (BOOLEAN)
  - `guest_token_hash` (SHA256 hash of token + IP)
  - `user_id` (FK to anonymous user)
  - Rate limiting fields

- **users** - Extended with:
  - `is_anonymous` (BOOLEAN) - Flags guest users
  - `email_verified` (BOOLEAN)
  - `guest_session_id` (FK to guest_sessions)

- **orders** - Extended with:
  - `user_id` (NULLABLE/POINTER) - NULL for guest orders
  - `guest_email` (VARCHAR) - Email for guest orders

### 2. Services

#### GuestService
Handles guest checkout workflow:

```
VerifyEmailAndSendCode(email, clientIP) error
  └─ Rate limit check (max 3 attempts per IP in 15 min)
  └─ Generate 6-digit code
  └─ Hash with bcrypt
  └─ Create GuestSession
  └─ Send email via Mailgun
  └─ Return 200 OK

ConfirmEmailAndCreateSession(email, code, clientIP) (*GuestSessionResponse, error)
  └─ Validate code (max 3 attempts)
  └─ Create anonymous user with is_anonymous=true
  └─ Generate JWT token type="guest" (2h expiry)
  └─ Hash token with IP: SHA256(token + clientIP)
  └─ Update GuestSession with UserID and token hash
  └─ Return guest_token

ValidateGuestToken(token, clientIP) (*GuestSession, error)
  └─ Parse JWT (validate type="guest")
  └─ Verify hash(token + clientIP) matches DB
  └─ Return GuestSession
```

#### OrderService - CreateGuestOrder
```
CreateGuestOrder(req *CreateGuestOrderRequest) (*OrderResponse, error)
  └─ Validate cart items (SKU from localStorage)
  └─ Resolve SKUs to products/variants
  └─ Validate prices against DB (anti-tampering)
  └─ Create Order with:
    - UserID = nil (guest order)
    - guest_email = req.Email
  └─ Create OrderItems from validated items
  └─ Calculate subtotal, tax, total
  └─ Return OrderResponse
```

#### PaymentService - CreateGuestPayment
```
CreateGuestPayment(ctx, req *CreateGuestPaymentRequest) (*PaymentResponse, error)
  └─ Find order by OrderID
  └─ Validate order is guest (guest_email not empty)
  └─ Validate email matches order.guest_email
  └─ Create Payment with:
    - UserID = nil (guest payment)
    - OrderID = req.OrderID
  └─ For MP payments: ExecuteGuestPayment()
    - Use guest_email as payer
  └─ Return PaymentResponse with mercadopago_preference_id
```

### 3. Security Mechanisms

#### Rate Limiting
- IP-based: Max 3 verification attempts per IP in 15 minutes
- Returns HTTP 429 if exceeded

#### Email Verification
- 6-digit code sent via Mailgun
- Valid for 10 minutes
- Hashed with bcrypt in DB
- Max 3 confirmation attempts

#### Session Token Security
- JWT type="guest" with 2-hour expiry
- Token hash = SHA256(token_string + client_ip)
- Only hash is stored in DB
- Validates token hasn't been used from different IP

#### Guest User Identity
- Anonymous user created with `is_anonymous=true`
- No password (empty string)
- Associated with guest_session

#### Price Validation
- Frontend passes prices with cart items
- Backend validates each item price matches DB:
  - For products: PriceRetail
  - For variants: PriceRetail + PriceAdjustment
- Prevents price tampering

### 4. API Endpoints

#### Guest Routes (Public - No Auth Required)

**POST /api/guest/verify-email**
```json
Request:
{
  "email": "user@example.com"
}

Response (200 OK):
{
  "message": "Verification code sent to your email",
  "expires_in_seconds": 600
}

Error (429):
{
  "error": "too many verification attempts from this IP. Try again in 15 minutes"
}
```

**POST /api/guest/confirm-email**
```json
Request:
{
  "email": "user@example.com",
  "verification_code": "123456"
}

Response (200 OK):
{
  "guest_token": "eyJhbGc...",
  "email": "user@example.com",
  "expires_at": "2026-05-21T14:30:00Z"
}
```

**POST /api/orders/guest** (Public - No Auth Required)
```json
Request:
{
  "email": "user@example.com",
  "first_name": "Juan",
  "last_name": "Pérez",
  "phone": "+5491123456789",
  "shipping_address": {
    "street": "Calle 123",
    "city": "Buenos Aires",
    "zipcode": "1400"
  },
  "delivery_method": "shipping",
  "cart_items": [
    {
      "sku": "PEN-RED-THIN",
      "quantity": 2,
      "price": 150.00
    }
  ],
  "notes": "Dejar en recepción"
}

Response (201 Created):
{
  "id": 1,
  "order_number": "ORD-20260521-123456",
  "user_id": 0,
  "guest_email": "user@example.com",
  "items": [...],
  "status": "PENDING",
  "subtotal": 300.00,
  "tax": 63.00,
  "total": 363.00,
  "delivery_method": "shipping",
  "created_at": "2026-05-21T12:30:00Z"
}
```

**POST /api/payments/guest** (Public - No Auth Required)
```json
Request:
{
  "order_id": 1,
  "email": "user@example.com",
  "amount": 363.00,
  "payment_method": "MP_CARD"
}

Response (201 Created):
{
  "id": 1,
  "transaction_id": "TXN-1716274200000000000",
  "order_id": 1,
  "user_id": null,
  "amount": 363.00,
  "currency": "ARS",
  "status": "PENDING",
  "payment_method": "MP_CARD",
  "mercadopago_preference_id": "123456789",
  "created_at": "2026-05-21T12:30:00Z"
}
```

### 5. Mercadopago Webhook Integration

The existing webhook handler continues to work without modifications:
- **POST /webhooks/mercadopago** receives MP notifications
- Webhook validates preference_id in Payment table
- Updates payment status based on MP response
- Can handle both authenticated and guest users
- Uses email from order.guest_email for notifications

### 6. Frontend Flow

```
1. User fills cart with localStorage
   └─ No backend involved yet

2. User checks out
   └─ POST /api/guest/verify-email
   └─ Display code input

3. User enters code
   └─ POST /api/guest/confirm-email
   └─ Store guest_token in sessionStorage

4. User fills shipping + names
   └─ POST /api/orders/guest
   └─ No auth header needed
   └─ Pass cartItems from localStorage
   └─ Get order_id

5. User proceeds to payment
   └─ POST /api/payments/guest
   └─ No auth header needed
   └─ Get mercadopago_preference_id

6. User completes MP flow
   └─ MP redirects to success/failure
   └─ Webhook updates payment status

7. Optional: Link account after
   └─ Not implemented yet
```

### 7. Database Schema Changes

**Migration: 005_guest_checkout_support.sql**

```sql
-- guest_sessions table
CREATE TABLE guest_sessions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL UNIQUE,
  verification_code_hash VARCHAR(255) NOT NULL,
  verification_code_sent_at TIMESTAMP NULL,
  verification_code_attempts INT DEFAULT 0,
  is_verified BOOLEAN DEFAULT FALSE,
  verified_at TIMESTAMP NULL,
  guest_token_hash VARCHAR(255) NULL,
  session_ip_address VARCHAR(45) NULL,
  user_id BIGINT UNSIGNED NULL UNIQUE,
  attempts_from_ip INT DEFAULT 0,
  last_attempt_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- users table modifications
ALTER TABLE users ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- orders table modifications
ALTER TABLE orders MODIFY COLUMN user_id BIGINT UNSIGNED NULL;
ALTER TABLE orders ADD COLUMN guest_email VARCHAR(255) NULL;
```

### 8. Configuration

Add to .env:
```env
MAILGUN_DOMAIN=sandbox-your-domain.mailgun.org
MAILGUN_API_KEY=key-your-api-key
MAILGUN_FROM_EMAIL=noreply@elcampeon.com
```

### 9. Testing Endpoints

```bash
# 1. Request verification code
curl -X POST http://localhost:8080/api/guest/verify-email \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# 2. Confirm code (use code sent to email or check logs)
curl -X POST http://localhost:8080/api/guest/confirm-email \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","verification_code":"123456"}'

# 3. Create guest order
curl -X POST http://localhost:8080/api/orders/guest \
  -H "Content-Type: application/json" \
  -d '{
    "email":"test@example.com",
    "first_name":"Test",
    "last_name":"User",
    "phone":"+5491123456789",
    "shipping_address":{"street":"Calle 123","city":"CABA"},
    "delivery_method":"shipping",
    "cart_items":[{"sku":"TEST-SKU","quantity":1,"price":100}]
  }'

# 4. Create payment
curl -X POST http://localhost:8080/api/payments/guest \
  -H "Content-Type: application/json" \
  -d '{
    "order_id":1,
    "email":"test@example.com",
    "amount":121.00,
    "payment_method":"MP_CARD"
  }'
```

## Files Modified/Created

### Created
- `migrations/005_guest_checkout_support.sql`
- `internal/models/guest.go`
- `internal/repositories/guest_repository.go`
- `internal/services/email/mailgun_service.go`
- `internal/services/email/context.go`
- `internal/services/guest/guest_service.go`
- `internal/handlers/guest_handler.go`
- `internal/middleware/guest.go`

### Modified
- `internal/config/config.go` - Added Mailgun config
- `internal/models/user.go` - Added is_anonymous, email_verified
- `internal/models/order.go` - UserID nullable, added guest_email
- `internal/utils/jwt.go` - Added GenerateGuestToken, ValidateGuestToken
- `internal/services/order/order_service.go` - Added CreateGuestOrder
- `internal/handlers/order_handler.go` - Added CreateGuestOrder handler
- `internal/services/payment/payment_service.go` - Added CreateGuestPayment, ExecuteGuestPayment
- `internal/handlers/payment_handler.go` - Added CreateGuestPayment handler
- `internal/handlers/routes.go` - Registered guest routes and handlers

## Notes

- Guest sessions expire after 7 days
- Verification codes expire after 10 minutes
- Guest users are permanently stored (no auto-cleanup scheduled)
- MercadoPago webhook works transparently for guest and authenticated users
- EmailService has fallback no-op implementation when Mailgun not configured

