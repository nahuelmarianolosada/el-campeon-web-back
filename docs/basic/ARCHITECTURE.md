# Arquitectura de "El Campeón Web"

## 1. Descripción General

"El Campeón Web" es un sistema backend monolítico para la venta de productos de librería y juguetería. Implementado en Go con arquitectura de capas, proporciona una plataforma e-commerce completa con gestión de productos, carrito persistente, órdenes y pagos a través de MercadoPago.

## 2. Principios Arquitectónicos

- **Monolito Modular**: Un único servicio cohesivo pero internamente desacoplado por capas
- **Separación de Responsabilidades**: Handlers → Services → Repositories → Models
- **Stateless**: Sesiones basadas en JWT, sin estado del servidor
- **Escalable**: Preparado para migración futura a microservicios
- **Seguro**: Autenticación JWT, autorización por roles, validación de inputs

## 3. Estructura de Carpetas

```
el-campeon-web/
├── cmd/
│   └── main.go                  # Punto de entrada de la aplicación
├── internal/
│   ├── config/
│   │   └── config.go           # Carga de configuración desde env vars
│   ├── database/
│   │   └── setup.go            # Inicialización de BD y migrations
│   ├── middleware/
│   │   └── auth.go             # Middleware de JWT y autorización
│   ├── models/
│   │   ├── user.go             # Modelos de Usuario y Auth
│   │   ├── product.go          # Modelos de Producto
│   │   ├── cart.go             # Modelos de Carrito
│   │   ├── order.go            # Modelos de Orden
│   │   └── payment.go          # Modelos de Pago
│   ├── repositories/
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── cart_repository.go
│   │   ├── order_repository.go
│   │   └── payment_repository.go
│   ├── services/
│   │   ├── user_service.go
│   │   ├── product_service.go
│   │   ├── cart_service.go
│   │   ├── order_service.go
│   │   └── payment_service.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── product_handler.go
│   │   ├── cart_handler.go
│   │   ├── order_handler.go
│   │   ├── payment_handler.go
│   │   └── routes.go           # Definición de rutas
│   └── utils/
│       ├── jwt.go              # Generación y validación de JWT
│       └── password.go         # Hashing de contraseñas
├── migrations/                 # Scripts SQL para BD
├── scripts/                    # Scripts utilitarios
├── dockerfiles/
│   └── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
├── README.md
└── ARCHITECTURE.md            # Este archivo
```

## 4. Capas de la Arquitectura

### 4.1 HTTP Layer (Handlers)
- Recibe solicitudes HTTP
- Valida parámetros y body del request
- Delega lógica de negocio a Services
- Retorna respuestas JSON estructuradas

Archivos:
- `auth_handler.go` - Registro, login, refresh token
- `product_handler.go` - CRUD de productos (admin), listado público
- `cart_handler.go` - Gestión del carrito de compras
- `order_handler.go` - Creación y seguimiento de órdenes
- `payment_handler.go` - Procesamiento de pagos y webhooks

### 4.2 Business Logic Layer (Services)
- Implementa reglas de negocio
- Orquesta operaciones complejas
- Validación de lógica
- Determinación de precios (mayorista/minorista)
- Transacciones y consistencia

Servicios:
- `UserService` - Autenticación y gestión de usuarios
- `ProductService` - CRUD y búsqueda de productos
- `CartService` - Gestión persistente del carrito
- `OrderService` - Creación de órdenes desde carrito
- `PaymentService` - Procesamiento de pagos

### 4.3 Data Access Layer (Repositories)
- Abstracción de acceso a BD
- Operaciones CRUD sobre entidades
- Consultas complejas
- Cierto nivel de lógica de persistencia

Repositorios:
- `UserRepository` - CRUD de usuarios
- `ProductRepository` - CRUD y búsqueda de productos
- `CartRepository` - Gestión del carrito
- `OrderRepository` - CRUD de órdenes
- `PaymentRepository` - CRUD de pagos

### 4.4 Models Layer
- Definición de estructuras de datos
- Etiquetas GORM para ORM
- Request/Response structs para API
- Validación con tags `binding`

### 4.5 Middleware Layer
- Autenticación con JWT
- Autorización por roles (USER, ADMIN)
- CORS
- Logging (extensible)

### 4.6 Utils Layer
- Funciones transversales
- JWT: Generación y validación
- Password: Hashing con bcrypt
- Constants y helpers

## 5. Flujo de una Solicitud

```
HTTP Request
    ↓
CORS Middleware
    ↓
Auth Middleware (si es protected)
    ↓
Handler (valida input)
    ↓
Service (lógica de negocio)
    ↓
Repository (acceso a BD)
    ↓
GORM/MySQL
    ↓
Response JSON
```

## 6. Flujos Principales

### 6.1 Flujo de Autenticación
1. Usuario se registra con email, contraseña y datos
2. Service valida email único
3. Service hashea contraseña con bcrypt
4. Se crea usuario en BD
5. Se generan tokens JWT (access + refresh)
6. Se retorna AuthResponse con tokens y datos del usuario

### 6.2 Flujo de Compra
1. Usuario explora productos (GET /api/products)
2. Agrega items al carrito (POST /api/cart/items)
3. Service verifica stock y aplica precio correcto
4. Usuario crea orden (POST /api/orders)
5. Service calcula subtotal, tax, total
6. Se vacía el carrito automáticamente
7. Se crea pago (POST /api/payments)
8. Usuario es redirigido a MercadoPago

### 6.3 Determinación de Precios
- Si usuario es `IsBulkBuyer` Y cantidad >= `MinBulkQuantity` → precio mayorista
- Si no → precio minorista

### 6.4 Flujo de Webhook MercadoPago
1. MercadoPago notifica pago completado
2. Se valida la firma del webhook
3. Se consulta estado en API de MercadoPago
4. Se actualiza estado de Payment en BD
5. Si APPROVED, se actualiza Order a CONFIRMED

## 7. Seguridad

### 7.1 Autenticación
- JWT con HS256
- Access token: 24 horas
- Refresh token: 7 días
- Renovación de tokens vía `/auth/refresh`

### 7.2 Autorización
- Middleware de roles (USER, ADMIN)
- Endpoints de admin protegidos
- Validación de propiedad de recursos (ej: mi carrito, mis órdenes)

### 7.3 Validación de Inputs
- Binding en handlers (required, email, min length)
- Tipado fuerte en Go
- Validación de enums (estados de orden, pago, etc)

### 7.4 Hashing de Contraseñas
- Bcrypt con cost factor DefaultCost
- Nunca se almacenan contraseñas en texto plano
- Verificación segura al login

## 8. Modelo de Datos

Ver `DATABASE.md` para esquema detallado.

Tablas principales:
- `users` - Usuarios del sistema
- `products` - Catálogo de productos
- `carts` - Carritos por usuario
- `cart_items` - Items dentro de cada carrito
- `orders` - Órdenes de compra
- `order_items` - Items de cada orden
- `payments` - Pagos y transacciones

## 9. Extensibilidad y Escalabilidad

### 9.1 Para Migrar a Microservicios
1. Separar servicios por dominio (Users, Products, Orders, Payments)
2. Introducir Message Queue (RabbitMQ, Kafka)
3. Cada servicio con su propia BD
4. API Gateway para enrutamiento

### 9.2 Para Escalar
1. Base de datos: Read replicas, sharding
2. Cache: Redis para productos, carritos
3. CDN para imágenes
4. Horizontal scaling de la app (load balancer)
5. Logging centralizado (ELK stack)

### 9.3 Mejoras Futuras
- Soft deletes (ya implementado con gorm.DeletedAt)
- Auditoría de cambios
- Rate limiting
- GraphQL API
- Búsqueda full-text (Elasticsearch)
- Recomendaciones de productos

## 10. Dependencias Principales

- **gin-gonic/gin** - Framework HTTP
- **gorm.io/gorm** - ORM
- **golang-jwt/jwt** - JWT
- **golang.org/x/crypto** - Bcrypt
- **mercadopago/sdk-go** - SDK de MercadoPago (futuro)
- **joho/godotenv** - Variables de entorno

## 11. Variables de Entorno Requeridas

```env
PORT=8080
ENV=development|production
DB_HOST=localhost
DB_PORT=3306
DB_USER=user
DB_PASSWORD=pass
DB_NAME=el_campeon_web
JWT_SECRET_KEY=secret
JWT_REFRESH_SECRET=refresh_secret
JWT_EXPIRY_HOURS=24
MERCADOPAGO_ACCESS_TOKEN=token
MERCADOPAGO_PUBLIC_KEY=key
API_BASE_URL=http://localhost:8080
```

## 12. Consideraciones de Producción

- Usar secretos fuertes para JWT
- HTTPS obligatorio
- SQL para queries complejas
- Índices en BD para performance
- Logging y monitoring
- Rate limiting en endpoints públicos
- Validación de CORS más restrictiva
- Testing de seguridad

