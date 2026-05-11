# 🔐 Guía de Seguridad - El Campeón Web

Recomendaciones y mejores prácticas de seguridad para el sistema.

## Tabla de Contenidos

1. [Autenticación y Autorización](#autenticación-y-autorización)
2. [Manejo de Datos Sensibles](#manejo-de-datos-sensibles)
3. [Protección de Contraseñas](#protección-de-contraseñas)
4. [HTTPS y Transporte](#https-y-transporte)
5. [Validación de Inputs](#validación-de-inputs)
6. [Inyección SQL y NoSQL](#inyección-sql-y-nosql)
7. [CORS y Seguridad del Navegador](#cors-y-seguridad-del-navegador)
8. [Rate Limiting](#rate-limiting)
9. [Logging y Monitoreo](#logging-y-monitoreo)
10. [Secretos y Configuración](#secretos-y-configuración)
11. [Errores de Seguridad Comunes](#errores-de-seguridad-comunes)
12. [Checklist de Producción](#checklist-de-producción)

---

## Autenticación y Autorización

### ✅ Implementado

- **JWT con HS256**: Tokens firmados criptográficamente
- **Access Token corto**: 24 horas (expira rápidamente)
- **Refresh Token largo**: 7 días (permite renovar sin re-login)
- **Roles**: USER y ADMIN
- **Middleware de autorización**: Valida tokens en endpoints protegidos

### 🔒 Mejoras para Producción

#### 1. Implementar RS256 (recomendado)

```go
// En lugar de HS256, usar RS256 con pares de claves pública/privada
// HS256: la misma clave firma y verifica (menos seguro)
// RS256: clave privada firma, clave pública verifica (más seguro)

// Generar claves:
// openssl genrsa -out private.key 2048
// openssl rsa -in private.key -pubout -out public.key

// Usar en config:
JWT_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----
JWT_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----
```

#### 2. Agregar Vinculación de IP

```go
// En JWT claims, agregar IP del cliente
type JWTClaims struct {
    // ...
    ClientIP string `json:"client_ip"`
}

// Al validar, verificar que la IP coincide
if claims.ClientIP != c.ClientIP() {
    return nil, errors.New("token used from different IP")
}
```

#### 3. Implementar Logout y Blacklist de Tokens

```go
// Mantener lista de tokens revocados (Redis)
type TokenBlacklist interface {
    Add(token string, exp time.Time) error
    IsBlacklisted(token string) bool
}

// En middleware:
if isBlacklisted(token) {
    return errors.New("token has been revoked")
}

// POST /auth/logout
// Agregar token a blacklist
```

#### 4. Agregar niveles de permiso fino

```go
// En lugar de solo USER/ADMIN, usar permisos granulares
type Permission string

const (
    PermReadProducts Permission = "products:read"
    PermCreateProduct Permission = "products:create"
    PermDeleteProduct Permission = "products:delete"
    PermListOrders Permission = "orders:list"
)

// En JWT claims
type JWTClaims struct {
    // ...
    Permissions []Permission `json:"permissions"`
}

// Middleware
func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.Get("claims").(*JWTClaims)
        if !hasPermission(claims.Permissions, perm) {
            c.JSON(403, "Permission denied")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## Manejo de Datos Sensibles

### ❌ NO hacer

```go
// ❌ Guardar contraseñas en texto plano
user.Password = "SuperSecret123"

// ❌ Retornar contraseñas en respuestas
json.Marshal(user) // Incluye Password

// ❌ Loguear datos sensibles
log.Printf("User: %+v", user) // Incluye contraseña

// ❌ Guardar tokens JWT en BD sin encripción
```

### ✅ Hacer

```go
// ✅ Hashear contraseñas con bcrypt
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// ✅ Usar struct sin campos sensibles para respuestas
type UserResponse struct {
    ID    uint
    Email string
    // NO incluir Password
}

// ✅ Enmascarar datos en logs
log.Printf("User login: %s (hash: %s...)", email, hash[:10])

// ✅ Encriptar tokens en BD si se persisten
```

---

## Protección de Contraseñas

### Política de Contraseñas

```go
// En RegisterRequest, agregar validadores más fuertes
const (
    MinPasswordLength = 12  // Cambiar de 8 a 12
    RequireUpperCase = true
    RequireNumbers = true
    RequireSpecialChars = true
)

func ValidatePassword(pwd string) error {
    if len(pwd) < MinPasswordLength {
        return fmt.Errorf("password must be at least %d chars", MinPasswordLength)
    }
    if !hasUpperCase(pwd) {
        return errors.New("password must contain uppercase letter")
    }
    if !hasNumber(pwd) {
        return errors.New("password must contain number")
    }
    if !hasSpecialChar(pwd) {
        return errors.New("password must contain special character (@!#$%)")
    }
    return nil
}
```

### Rate Limiting en Login

```go
// Limitar intentos fallidos de login
type LoginAttempts struct {
    attempts map[string][]time.Time // email -> []timestamp
    mu sync.Mutex
}

func (la *LoginAttempts) IsLocked(email string) bool {
    la.mu.Lock()
    defer la.mu.Unlock()

    attempts := la.attempts[email]
    // Limpiar intentos antiguos (> 15 minutos)
    validAttempts := []time.Time{}
    for _, t := range attempts {
        if time.Since(t) < 15*time.Minute {
            validAttempts = append(validAttempts, t)
        }
    }

    la.attempts[email] = validAttempts
    
    // Bloquear después de 5 intentos
    return len(validAttempts) >= 5
}

// En handler de login:
if loginAttempts.IsLocked(email) {
    c.JSON(429, gin.H{"error": "Too many login attempts. Try again later"})
    return
}
```

---

## HTTPS y Transporte

### ❌ Desarrollo

```bash
HTTP para desarrollo local es OK
```

### ✅ Producción

```go
// Forzar HTTPS
func httpsRedirect() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Header.Get("X-Forwarded-Proto") != "https" {
            c.Redirect(301, "https://"+c.Request.Host+c.Request.RequestURI)
            c.Abort()
            return
        }
        c.Next()
    }
}

// En main.go
if cfg.ServerEnv == "production" {
    router.Use(httpsRedirect())
}

// Headers de seguridad
func securityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}

// Registrar:
router.Use(securityHeaders())
```

### Certificado SSL

```bash
# Con Let's Encrypt y Certbot
sudo certbot certonly --standalone -d api.elcampeon.com

# Copiar certificados
sudo cp /etc/letsencrypt/live/api.elcampeon.com/fullchain.pem /app/cert.pem
sudo cp /etc/letsencrypt/live/api.elcampeon.com/privkey.pem /app/key.pem

# En Go, usar TLS:
// router.RunTLS(":443", "cert.pem", "key.pem")
```

---

## Validación de Inputs

### ❌ Sin Validación

```go
// ❌ Confiar en datos del cliente
user.Email = c.PostForm("email")

// ❌ No verificar listas permitidas
order.Status = c.PostForm("status") // Cualquier valor
```

### ✅ Con Validación

```go
// ✅ Usar binding con validadores
type RegisterRequest struct {
    Email    string `binding:"required,email"`
    Password string `binding:"required,min=12"`
    Name     string `binding:"required,max=100"`
}

// Enums permitidos
var validOrderStatuses = map[string]bool{
    "PENDING":    true,
    "CONFIRMED":  true,
    "SHIPPED":    true,
    "DELIVERED":  true,
    "CANCELLED":  true,
}

func ValidateOrderStatus(status string) error {
    if !validOrderStatuses[status] {
        return fmt.Errorf("invalid status: %s", status)
    }
    return nil
}

// Usar en handler:
if err := ValidateOrderStatus(req.Status); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```

### Sanitización

```go
// Para campos de texto libre
func sanitizeText(text string) string {
    // Remover caracteres potencialmente peligrosos
    text = strings.TrimSpace(text)
    text = strings.ReplaceAll(text, "<script>", "")
    text = strings.ReplaceAll(text, "</script>", "")
    return text
}
```

---

## Inyección SQL y NoSQL

### ✅ GORM previene inyección SQL

```go
// ✅ GORM con prepared statements (seguro)
var user models.User
db.Where("email = ?", email).First(&user)

// ❌ Esto sería unsafe (no hacerlo)
// db.Where("email = '" + email + "'").First(&user)
```

### Query Complejas Seguras

```go
// Para queries más complejas, usar SQL con placeholders
db.Raw(
    "SELECT * FROM orders WHERE status = ? AND user_id = ?", 
    "CONFIRMED", 
    userID,
).Scan(&orders)

// Nunca concatenar strings
// db.Raw("SELECT * FROM orders WHERE status = '" + status + "'")
```

---

## CORS y Seguridad del Navegador

### ❌ Desarrollo

```go
// ❌ CORS abierto para todo (solo desarrollo)
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

### ✅ Producción

```go
import "github.com/gin-contrib/cors"

func setupCORS(router *gin.Engine, cfg *config.Config) {
    if cfg.ServerEnv == "production" {
        router.Use(cors.New(cors.Config{
            AllowOrigins:     []string{"https://www.elcampeon.com"},
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Authorization", "Content-Type"},
            ExposeHeaders:    []string{"Content-Length"},
            AllowCredentials: true,
            MaxAge:           12 * time.Hour,
        }))
    } else {
        router.Use(cors.Default())
    }
}
```

---

## Rate Limiting

### Implementar con Middleware

```go
import "github.com/go-redis/redis/v8"
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu sync.RWMutex
}

func (rl *RateLimiter) Limit(ip string) bool {
    rl.mu.RLock()
    limiter, exists := rl.limiters[ip]
    rl.mu.RUnlock()

    if !exists {
        // 100 requests per minute
        limiter = rate.NewLimiter(rate.Every(time.Minute/100), 1)
        rl.mu.Lock()
        rl.limiters[ip] = limiter
        rl.mu.Unlock()
    }

    return limiter.Allow()
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Limit(c.ClientIP()) {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Registrar:
limiter := &RateLimiter{limiters: make(map[string]*rate.Limiter)}
router.Use(RateLimitMiddleware(limiter))
```

---

## Logging y Monitoreo

### ❌ Qué NO loguear

```go
❌ Contraseñas
❌ Tokens JWT
❌ Información de BD (credenciales)
❌ Datos de tarjetas de crédito
❌ SSN/IDs privados
```

### ✅ Qué SÍ loguear

```go
✅ Intentos de login fallidos (con IP)
✅ Cambios en órdenes
✅ Errores de BD
✅ Acceso a endpoints admin
✅ Webhooks recibidos
```

### Estructura de Logs

```go
type AuditLog struct {
    ID        uint
    UserID    uint
    Action    string              // "CREATE_ORDER", "UPDATE_PAYMENT"
    Resource  string              // "orders", "payments"
    Details   map[string]interface{} // Datos relevantes
    ClientIP  string
    Status    string              // "success", "failure"
    CreatedAt time.Time
}

// Logger helper
func LogAction(db *gorm.DB, userID uint, action, resource string, details map[string]interface{}, status string) {
    log := AuditLog{
        UserID:   userID,
        Action:   action,
        Resource: resource,
        Details:  details,
        Status:   status,
    }
    db.Create(&log)
}

// En handlers:
LogAction(db, userID, "CREATE_ORDER", "orders", map[string]interface{}{
    "order_id": order.ID,
    "total": order.Total,
}, "success")
```

---

## Secretos y Configuración

### ❌ Nunca Hacer

```go
❌ Guardar secrets en código
❌ Commitear .env a Git
❌ Hardcodear API keys
❌ Compartir secrets en Slack/Email
```

### ✅ Hacer

```go
✅ Usar variables de entorno
✅ Agregar .env al .gitignore
✅ Usar secret manager (AWS Secrets, HashiCorp Vault)
✅ Rotar secrets regularmente
✅ Usar diferentes secrets para dev/prod
```

### Ejemplo con Vault

```go
import "github.com/hashicorp/vault/api"

func getSecret(secretPath string) (string, error) {
    client, _ := api.NewClient(api.DefaultConfig())
    secret, _ := client.Logical().Read(secretPath)
    return secret.Data["value"].(string), nil
}

// Usar en config:
cfg.JWTSecretKey, _ = getSecret("secret/jwt-secret")
cfg.MercadopagoAccessToken, _ = getSecret("secret/mercadopago-token")
```

---

## Errores de Seguridad Comunes

### 1. Time-Based Information Disclosure

```go
// ❌ Tiempo de respuesta revela si email existe
start := time.Now()
if user := findByEmail(email); user != nil {
    // ... takes longer
}
duration := time.Since(start)
// Attacker puede medir duración

// ✅ Usar duración consistente
start := time.Now()
user := findByEmail(email)
exists := user != nil
elapsed := time.Since(start)
if elapsed < 100*time.Millisecond {
    time.Sleep(time.Until(time.Now().Add(100*time.Millisecond)))
}
return exists
```

### 2. Payment Amount Manipulation

```go
// ❌ Confiar en amount del cliente
CreatePayment(orderID, c.PostForm("amount"))

// ✅ Obtener amount de orden en BD
order, _ := GetOrder(orderID)
// Validar que amount coincide
if amount != order.Total {
    return errors.New("invalid amount")
}
CreatePayment(orderID, order.Total)
```

### 3. Privilege Escalation

```go
// ❌ Confiar en role del cliente
user.Role = c.PostForm("role")

// ✅ Solo admin puede cambiar roles
if role, exists := c.Get("role"); !exists || role != "ADMIN" {
    c.JSON(403, "Permission denied")
    return
}
user.Role = newRole
```

---

## Checklist de Producción

### Seguridad

- [ ] HTTPS habilitado con certificado válido
- [ ] JWT usando RS256 (no HS256)
- [ ] Secretos en secret manager (no .env)
- [ ] Rate limiting implementado
- [ ] CORS restringido a dominios permitidos
- [ ] Headers de seguridad agregados
- [ ] SQL Injection protección (GORM con placeholders)
- [ ] Validación de inputs en todos los endpoints
- [ ] Autenticación en todos los endpoints protegidos
- [ ] Autorización verificada en operaciones sensibles

### Base de Datos

- [ ] Backups automáticos diarios
- [ ] Encriptación en tránsito (SSL/TLS)
- [ ] Encriptación en reposo (para datos sensibles)
- [ ] Índices optimizados
- [ ] Usuarios de BD con permisos mínimos
- [ ] No usar root en producción

### Aplicación

- [ ] Logging y monitoreo configurado
- [ ] Alertas para errores críticos
- [ ] Error handling sin exponer detalles internos
- [ ] Timeouts en requests externos
- [ ] Health checks implementados
- [ ] Graceful shutdown

### Infraestructura

- [ ] Firewall configurado
- [ ] Solo puertos necesarios abiertos (80, 443)
- [ ] SSH con claves (no passwords)
- [ ] Actualizaciones de seguridad aplicadas
- [ ] Monitoreo de recursos (CPU, memoria)
- [ ] Respaldo de datos críticos

### Desarrollo

- [ ] Código revisado por peer
- [ ] STATIC análisis ejecutado (golangci-lint)
- [ ] Tests de seguridad incluidos
- [ ] Documentación de seguridad actualizada
- [ ] Plan de respuesta a incidentes

---

## Recursos

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security/best-practices)
- [JWT Best Current Practices](https://datatracker.ietf.org/doc/html/rfc8949)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

---

**Recuerda**: La seguridad es un proceso continuo, no un destino.

