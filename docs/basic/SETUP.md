# 📖 Guía de Instalación - El Campeón Web

Guía paso a paso para instalar y ejecutar el sistema en tu máquina local.

## Tabla de Contenidos

1. [Requisitos](#requisitos)
2. [Opción A: Instalación con Docker (Recomendado)](#opción-a-instalación-con-docker-recomendado)
3. [Opción B: Instalación Local](#opción-b-instalación-local)
4. [Opción C: Instalación con Docker solo BD](#opción-c-instalación-con-docker-solo-bd)
5. [Verificación de la Instalación](#verificación-de-la-instalación)
6. [Ejecutar Tests](#ejecutar-tests)
7. [Troubleshooting](#troubleshooting)

---

## Requisitos

### Requisitos Globales

- **Git** (para clonar el repositorio)
- **Acceso a terminal/CMD**

### Opción A: Docker
- **Docker** 20.10+ ([Descargar](https://www.docker.com/products/docker-desktop))
- **Docker Compose** 2.0+ (incluido en Docker Desktop)

### Opción B: Local
- **Go** 1.21+ ([Descargar](https://golang.org/))
- **MySQL** 8.0+ ([Descargar](https://dev.mysql.com/downloads/mysql/))
- **MySQL Client** (generalmente incluido con MySQL)

### Opción C: Híbrida
- **Go** 1.21+
- **Docker** y **Docker Compose**

---

## Opción A: Instalación con Docker (Recomendado)

**Ventajas**: Todo automatizado, sin dependencias locales

### Paso 1: Clonar el repositorio

```bash
git clone https://github.com/nahuelmarianolosada/el-campeon-web.git
cd el-campeon-web
```

### Paso 2: Configurar variables de entorno (opcional)

```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar si necesitas valores diferentes
# nano .env  (Linux/Mac)
# notepad .env  (Windows)
```

**Valores por defecto en docker-compose.yml:**
```env
PORT=8080
ENV=development
DB_HOST=db
DB_PORT=3306
DB_USER=el_campeon_user
DB_PASSWORD=user_password_change_me
DB_NAME=el_campeon_web
```

### Paso 3: Construir y ejecutar

```bash
# Construir imágenes y levantar servicios
docker-compose up

# O en background
docker-compose up -d

# Ver logs en tiempo real (si está en background)
docker-compose logs -f app
```

**Primera ejecución**: Docker descargará las imágenes base, instalará dependencias de Go y compilará el código (puede tomar 5-10 minutos).

### Paso 4: Verificar que está funcionando

```bash
# En otra terminal
curl http://localhost:8080/health

# Respuesta esperada:
# {"status":"ok","service":"el-campeon-web"}
```

### Paso 5: Detener servicios

```bash
# Detener (los containers siguen existiendo)
docker-compose stop

# Detener y eliminar containers
docker-compose down

# Detener, eliminar y limpiar volúmenes
docker-compose down -v
```

---

## Opción B: Instalación Local

**Requiere**: Go 1.21+ y MySQL 8.0+

### Paso 1: Instalar dependencias

**macOS (con Homebrew):**
```bash
brew install go mysql
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install golang-1.21 mysql-server
```

**Windows:**
- Descargar Go desde [golang.org](https://golang.org/)
- Descargar MySQL desde [mysql.com](https://dev.mysql.com/downloads/mysql/)

### Paso 2: Clonar el repositorio

```bash
git clone https://github.com/nahuelmarianolosada/el-campeon-web.git
cd el-campeon-web
```

### Paso 3: Descargar dependencias de Go

```bash
go mod download
go mod verify
```

### Paso 4: Inicializar Base de Datos

**macOS/Linux:**
```bash
# Iniciar MySQL (si no está corriendo)
mysql.server start  # macOS
# sudo service mysql start  # Linux

# Crear base de datos
mysql -u root -p < migrations/init.sql
# Se pedirá el password de MySQL

# Verificar
mysql -u root -p -e "USE el_campeon_web; SHOW TABLES;"
```

**Windows CMD:**
```cmd
:: Iniciar MySQL
net start MySQL80

:: Crear base de datos
mysql -u root -p < migrations\init.sql

:: Verificar
mysql -u root -p -e "USE el_campeon_web; SHOW TABLES;"
```

### Paso 5: Configurar variables de entorno

```bash
cp .env.example .env

# Editar .env:
# DB_HOST=localhost
# DB_USER=root
# DB_PASSWORD=<your-mysql-password>
```

### Paso 6: Ejecutar la aplicación

```bash
go run ./cmd/main.go

# Salida esperada:
# 2024/04/27 10:30:00 Database initialized successfully
# 2024/04/27 10:30:01 Starting server on :8080
```

### Paso 7: Verificar en otra terminal

```bash
curl http://localhost:8080/health
```

---

## Opción C: Instalación con Docker solo BD

**Híbrida**: MySQL en Docker, Go local

### Paso 1: Clonar y descargar dependencias

```bash
git clone https://github.com/nahuelmarianolosada/el-campeon-web.git
cd el-campeon-web

go mod download
```

### Paso 2: Levantar solo MySQL

Crear `docker-compose-db.yml`:
```yaml
version: '3.8'
services:
  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: el_campeon_web
      MYSQL_USER: el_campeon_user
      MYSQL_PASSWORD: user_pass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

Ejecutar:
```bash
docker-compose -f docker-compose-db.yml up -d

# Esperar a que MySQL esté listo (20-30 segundos)
sleep 30

# Inicializar BD
docker exec -i <container-id> mysql -u root -proot < migrations/init.sql
```

### Paso 3: Configurar variables de entorno

```bash
cp .env.example .env

# .env debe tener:
DB_HOST=localhost
DB_PORT=3306
DB_USER=el_campeon_user
DB_PASSWORD=user_pass
```

### Paso 4: Ejecutar Go

```bash
go run ./cmd/main.go
```

---

## Verificación de la Instalación

### ✅ Verificación Completa

#### 1. Health Check
```bash
curl -s http://localhost:8080/health | jq .
```

Respuesta esperada:
```json
{
  "status": "ok",
  "service": "el-campeon-web"
}
```

#### 2. Registro de Usuario
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "Password123!",
    "phone": "+5491123456789"
  }' | jq .
```

Respuesta esperada: `access_token`, `refresh_token`, `user`

#### 3. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Password123!"
  }' | jq .
```

#### 4. Listar Productos
```bash
curl http://localhost:8080/api/products | jq .
```

Debería retornar la lista de productos de ejemplo.

### 📊 Verificar Base de Datos

```bash
# Con Docker
docker-compose exec db mysql -u el_campeon_user -puser_password_change_me el_campeon_web

# O local
mysql -u root -p el_campeon_web

# Dentro de MySQL
SHOW TABLES;
SELECT * FROM users;
SELECT COUNT(*) FROM products;
```

---

## Ejecutar Tests

### Tests Unitarios

```bash
# Todos los tests
go test ./...

# Tests específicos
go test ./internal/services/...
go test ./internal/repositories/...

# Con verbose
go test -v ./...

# Con coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Ejecutar Test Específico

```bash
# Test de registro
go test -run TestRegisterSuccess ./internal/services

# Test de login
go test -run TestLoginSuccess ./internal/services
```

---

## Troubleshooting

### ❌ Docker "Port 8080 already in use"

```bash
# Cambiar puerto en docker-compose.yml
ports:
  - "8081:8080"

# O encontrar y matar proceso
lsof -i :8080
kill -9 <PID>

# macOS
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill -9
```

### ❌ Docker "Port 3306 already in use"

```bash
# Parar outro MySQL
docker ps | grep mysql
docker stop <container-id>

# O cambiar puerto
ports:
  - "3307:3306"
```

### ❌ "Connection refused" a BD

```bash
# Verificar que MySQL está corriendo
docker-compose ps

# Ver logs
docker-compose logs db

# Esperar más tiempo
sleep 30 && go run ./cmd/main.go
```

### ❌ "Unknown database 'el_campeon_web'"

```bash
# Inicializar BD
docker-compose exec db mysql -u root -proot_password_change_me < migrations/init.sql

# O manual dentro del container
docker exec -it <db-container-id> /bin/sh
mysql -u root -proot_password_change_me
CREATE DATABASE IF NOT EXISTS el_campeon_web;
USE el_campeon_web;
source /path/to/init.sql
```

### ❌ "Module not found" en Go

```bash
go mod tidy
go mod download
go mod verify
```

### ❌ Build error en Docker

```bash
# Limpiar caché de Docker
docker-compose down -v
docker system prune -a

# Reconstruir
docker-compose build --no-cache
docker-compose up
```

### ❌ Tests fallan localmente

```bash
# Asegurar que Go y dependencias están correctas
go version  # Debe ser 1.21+

# Limpiar y reinstalar
go clean -testcache
go mod tidy

# Correr tests de nuevo
go test ./...
```

### ⚠️ Logs para Debugging

**Con Docker:**
```bash
docker-compose logs app          # Logs de la app
docker-compose logs db           # Logs de BD
docker-compose logs -f app       # Ver en tiempo real
```

**Localmente:**
```bash
# La app imprime logs en stdout
# LEVEL timestamp message

# Para más info, editar cmd/main.go y agregar logging
```

---

## Variables de Entorno Completas

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=el_campeon_user
DB_PASSWORD=user_password_change_me
DB_NAME=el_campeon_web

# JWT
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_REFRESH_SECRET=your-refresh-secret-change-in-production
JWT_EXPIRY_HOURS=24

# MercadoPago
MERCADOPAGO_ACCESS_TOKEN=
MERCADOPAGO_PUBLIC_KEY=

# API
API_BASE_URL=http://localhost:8080
```

---

## Siguientes Pasos

1. 📖 Leer [ARCHITECTURE.md](ARCHITECTURE.md) para entender el diseño
2. 🔗 Revisar [API.md](API.md) para endpoints disponibles
3. 🗄️ Explorar [DATABASE.md](DATABASE.md) para esquema de BD
4. 🧪 Escribir tests para nuevas features
5. 🚀 Desplegar en producción cuando esté listo

---

## Ayuda

Si tienes problemas:

1. Verifica los [Troubleshooting](#troubleshooting) arriba
2. Revisa los logs: `docker-compose logs`
3. Abre un issue en GitHub
4. Contacta al equipo de soporte

¡Bienvenido a El Campeón Web! 🎉

