# 🎯 El Campeón Web - Backend E-commerce

Sistema backend monolítico completo para la venta de productos de librería y juguetería, desarrollado en **Go** con arquitectura de capas, **JWT**, **MySQL** e integración con **MercadoPago**.

## 🚀 Características

✅ **Autenticación y Autorización**
- Registro y login con JWT
- Roles: USER y ADMIN
- Access token (24h) + Refresh token (7 días)
- Middleware de autenticación

✅ **Gestión de Productos**
- CRUD completo (Admin)
- Precios minorista y mayorista
- Categorías y búsqueda
- Control de stock
- Soft deletes

✅ **Carrito de Compras**
- Persistente por usuario
- Cálculo automático de precios
- Agregar, actualizar, eliminar items
- Total con impuestos

✅ **Sistema de Órdenes**
- Creación desde carrito
- Direcciones de envío
- Estados de orden (PENDING, CONFIRMED, SHIPPED, DELIVERED, CANCELLED)
- Historial de compras

✅ **Procesamiento de Pagos**
- Integración con MercadoPago
- Estados de pago
- Webhooks para confirmación
- Transacciones seguras

✅ **Infraestructura**
- Docker + Docker Compose
- MySQL 8.0
- Gin HTTP Framework
- GORM ORM
- Logging y error handling

## 📋 Requisitos

- **Go** 1.21+
- **Docker** y **Docker Compose**
- **MySQL** 8.0 (o usa Docker Compose)
- **Git**

## 🏗️ Estructura del Proyecto

```
el-campeon-web/
├── cmd/
│   └── main.go                    # Punto de entrada
├── internal/
│   ├── config/                    # Configuración
│   ├── database/                  # Inicialización de BD
│   ├── handlers/                  # HTTP handlers
│   ├── middleware/                # Middleware de auth
│   ├── models/                    # Estructuras de datos
│   ├── repositories/              # Acceso a datos
│   ├── services/                  # Lógica de negocio
│   └── utils/                     # Utilidades (JWT, password)
├── migrations/                    # Scripts SQL
├── dockerfiles/
│   └── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
├── ARCHITECTURE.md                # Documentación de arquitectura
├── DATABASE.md                    # Esquema de BD
├── API.md                         # Especificación de endpoints
└── SETUP.md                       # Guía de instalación
```

## 🚀 Inicio Rápido con Docker

### 1. Clonar el repositorio
```bash
git clone <repo-url>
cd el-campeon-web
```

### 2. Configurar variables de entorno
```bash
cp .env.example .env
# Editar .env con tus valores (opcional)
```

### 3. Ejecutar con Docker Compose
```bash
docker-compose up
```

El servicio está disponible en `http://localhost:8080`

## 🛠️ Instalación Local sin Docker

### 1. Instalar dependencias de Go
```bash
go mod download
```

### 2. Configurar MySQL
```bash
# Opción A: Usar Docker solo para MySQL
docker run --name mysql-campeon \
  -e MYSQL_ROOT_PASSWORD=root_pass \
  -e MYSQL_DATABASE=el_campeon_web \
  -e MYSQL_USER=el_campeon_user \
  -e MYSQL_PASSWORD=user_pass \
  -p 3306:3306 \
  -d mysql:8.0

# Opción B: Usar MySQL instalado localmente
# Crear BD manualmente
mysql -u root -p < migrations/init.sql
```

### 3. Configurar variables de entorno
```bash
cp .env.example .env
# Editar .env con tus valores
```

### 4. Ejecutar la aplicación
```bash
go run ./cmd/main.go
```

## 📚 Documentación

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Descripción detallada de la arquitectura
- **[DATABASE.md](./DATABASE.md)** - Esquema de base de datos y relaciones
- **[API.md](./API.md)** - Especificación completa de endpoints REST
- **[SETUP.md](./SETUP.md)** - Guía paso a paso de instalación

## 🧪 Testing

### Ejecutar tests
```bash
go test ./...
```

### Tests específicos
```bash
go test ./internal/services/...
go test ./internal/repositories/...
```

### Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📡 API Endpoints Principales

### Autenticación
```bash
POST /auth/register           # Registrar usuario
POST /auth/login              # Login
POST /auth/refresh            # Renovar token
```

### Productos
```bash
GET  /api/products            # Listar productos
GET  /api/products/:id        # Obtener producto
POST /api/products            # Crear (ADMIN)
PUT  /api/products/:id        # Actualizar (ADMIN)
DELETE /api/products/:id      # Eliminar (ADMIN)
```

### Carrito
```bash
GET  /api/cart                # Obtener carrito
POST /api/cart/items          # Agregar al carrito
PUT  /api/cart/items/:itemId  # Actualizar cantidad
DELETE /api/cart/items/:itemId # Eliminar del carrito
DELETE /api/cart              # Vaciar carrito
GET  /api/cart/total          # Obtener total
```

### Órdenes
```bash
POST /api/orders              # Crear orden
GET  /api/orders/my           # Mis órdenes
GET  /api/orders/:id          # Detalles de orden
PUT  /api/orders/:id/status   # Actualizar estado (ADMIN)
GET  /api/orders              # Listar todas (ADMIN)
```

### Pagos
```bash
POST /api/payments            # Crear pago
GET  /api/payments/my         # Mis pagos
GET  /api/payments/:id        # Detalles de pago
GET  /api/payments/order/:id  # Pago de orden
PUT  /api/payments/:id/status # Actualizar estado (ADMIN)
```

## 🔐 Seguridad

- ✅ **Contraseñas**: Hashing con bcrypt
- ✅ **JWT**: HS256 con secretos fuertes
- ✅ **Autorización**: Middleware de rol (USER/ADMIN)
- ✅ **Validación**: Input validation en todos los handlers
- ✅ **CORS**: Habilitado para desarrollo
- ✅ **HTTPS**: Configurar en producción

### Cambiar Secretos en Producción
```env
JWT_SECRET_KEY=your-strong-secret-key-change-me
JWT_REFRESH_SECRET=your-strong-refresh-secret-change-me
```

## 💰 Modelo de Precios

El sistema soporta dos modelos de precios:

1. **Precio Minorista (Retail)**
   - Aplicado a usuarios normales
   - O cuando cantidad < `min_bulk_quantity`

2. **Precio Mayorista (Wholesale)**
   - Aplicado a usuarios con `is_bulk_buyer = true`
   - Y cantidad >= `min_bulk_quantity`

Ejemplo:
- Producto: Libro
- Precio minorista: $100
- Precio mayorista: $80
- Cantidad mínima: 5 unidades

```
Usuario normal:  $100 c/u (aunque compre 10)
Usuario mayorista (10 unidades): $80 c/u
Usuario mayorista (3 unidades): $100 c/u (< mínimo)
```

## 🔗 Integración MercadoPago

### Configuración
```env
MERCADOPAGO_ACCESS_TOKEN=your_access_token
MERCADOPAGO_PUBLIC_KEY=your_public_key
```

### Flujo de Pago
1. Usuario crea orden
2. Usuario inicia pago (POST /api/payments)
3. Sistema crea preferencia de pago en MP
4. Usuario es redirigido al checkout de MP
5. MP retorna webhook a `/webhooks/mercadopago`
6. Sistema actualiza estado del pago
7. Si aprobado, orden cambia a CONFIRMED

## 📊 Impuestos

El sistema calcula IVA al 21% (Argentina):

```
subtotal = SUM(item.quantity * item.price)
tax = subtotal * 0.21
total = subtotal + tax
```

Personalizable en `OrderService.CreateOrder()`

## 🚀 Despliegue en Producción

### 1. Configuración
```bash
# Cambiar valores en .env
ENV=production
PORT=8080
DB_HOST=prod-db.example.com
JWT_SECRET_KEY=<strong-random-key>
JWT_REFRESH_SECRET=<strong-random-key>
```

### 2. Build
```bash
go build -o ./bin/el-campeon ./cmd/main.go
```

### 3. Docker
```bash
docker build -t el-campeon:latest -f dockerfiles/Dockerfile .
docker push el-campeon:latest
```

### 4. Recomendaciones
- HTTPS obligatorio (Let's Encrypt)
- Base de datos en host separado
- Variables de entorno en secrets manager
- Logging centralizado
- Monitoreo y alertas
- Rate limiting
- CORS restringido

## 📈 Escalabilidad Futura

El sistema está diseñado para facilitar:

1. **Caché**: Redis para productos y carritos
2. **Búsqueda**: Elasticsearch para full-text search
3. **Microservicios**: Separación por dominio
4. **Queue**: RabbitMQ para procesamiento asynchrono
5. **CDN**: Cloudflare para imágenes
6. **Analytics**: Event tracking y reporting

## 🐛 Troubleshooting

### Puerto 3306 en uso
```bash
docker ps | grep mysql
docker stop <container-id>
```

### Errores de conexión a BD
```bash
# Verificar que MySQL está corriendo
docker-compose ps

# Ver logs
docker-compose logs db
```

### Problemas con módulos Go
```bash
go mod tidy
go mod verify
```

## 📝 Licencia

MIT License - Ver LICENSE para detalles

## 👨‍💻 Contribuciones

Las contribuciones son bienvenidas. Por favor:

1. Fork el proyecto
2. Crea una rama (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📧 Soporte

Para preguntas o reportar bugs, abrir un issue en GitHub.

## 🎓 Recursos

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [JWT-GO](https://github.com/golang-jwt/jwt)
- [MercadoPago API](https://developers.mercadopago.com/)

---

Made with ❤️ for El Campeón Web

