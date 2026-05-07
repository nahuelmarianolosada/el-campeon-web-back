# 📚 Tutorial Simplificado: Desplegar El Campeón Web en AWS EC2 (Con IP Pública)

## 📖 Tabla de Contenidos

1. [Preparación Inicial](#preparación-inicial)
2. [Configurar AWS Credentials](#configurar-aws-credentials)
3. [Crear Instancia EC2](#crear-instancia-ec2)
4. [Configurar Seguridad](#configurar-seguridad)
5. [Instalar Dependencias en EC2](#instalar-dependencias-en-ec2)
6. [Desplegar la Aplicación](#desplegar-la-aplicación)
7. [Configurar Nginx](#configurar-nginx)
8. [Probar en Producción](#probar-en-producción)
9. [Monitoreo y Mantenimiento](#monitoreo-y-mantenimiento)

---

## Preparación Inicial

### Antes de Empezar

Necesitarás:
- ✅ Cuenta de AWS activa (con tarjeta de crédito)
- ✅ AWS CLI v2 instalado en tu máquina local
- ✅ SSH client (macOS/Linux tienen uno, Windows: PuTTY o usar WSL)
- ✅ Editor de texto o IDE
- ✅ Conocimiento básico de terminal/command line

### Verificar Instalaciones Locales

**macOS:**
```bash
# Verificar que tengas AWS CLI
aws --version
# Respuesta esperada: aws-cli/2.x.x

# Si no lo tienes, instalar con Homebrew
brew install awscli

# SSH ya está incluido
ssh -V
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y awscli openssh-client

aws --version
ssh -V
```

**Windows (PowerShell):**
```powershell
# Instalar AWS CLI con MSI desde:
# https://awscli.amazonaws.com/AWSCLIV2.msi

# Verificar instalación
aws --version

# Para SSH, usar:
# 1. OpenSSH en Windows 10+
# 2. O PuTTY (https://www.putty.org/)
# 3. O Git Bash que incluye SSH
```

---

## Configurar AWS Credentials

AWS usa credenciales para autenticar tus solicitudes. Necesitas crear una clave de acceso.

### Paso 1: Crear Usuario IAM en AWS Console

1. Abre [AWS Console](https://console.aws.amazon.com)
2. Navega a **IAM** → **Users** → **Create User**
3. Username: `el-campeon-deployer`
4. Click **Next**

### Paso 2: Asignar Permisos

1. En "Set permissions", selecciona **Attach policies directly**
2. Busca y selecciona estas políticas:
   - `AmazonEC2FullAccess` (para crear instancias)
   - `CloudWatchAgentServerPolicy` (para monitoreo)
3. Click **Next** → **Create User**

### Paso 3: Crear Clave de Acceso

1. Click en el usuario creado
2. **Security Credentials** → **Create Access Key**
3. Selecciona: **Command Line Interface (CLI)**
4. Acepta términos y click **Create**
5. **Copiar y guardar seguro**:
   ```
   Access Key ID: AKIA...
   Secret Access Key: wJal...
   ```

⚠️ **IMPORTANTE**: Guarda estas llaves en lugar seguro. No las compartas.

### Paso 4: Configurar AWS CLI Localmente

```bash
# Ejecutar configuración interactiva
aws configure

# Se te pedirá:
# AWS Access Key ID: [Pega la clave del paso anterior]
# AWS Secret Access Key: [Pega la clave secreta]
# Default region name: us-east-1 (o tu región preferida)
# Default output format: json

# Verificar que funciona
aws sts get-caller-identity
# Respuesta: Deberá mostrar tu AWS Account ID
```

**Guardará credenciales en:**
```bash
# macOS/Linux:
~/.aws/credentials
~/.aws/config

# Windows:
C:\Users\<USERNAME>\.aws\credentials
C:\Users\<USERNAME>\.aws\config
```

---

## Crear Instancia EC2

### Paso 1: Generar Key Pair (Para SSH)

Una key pair es necesaria para acceder por SSH a tu instancia.

**Opción A: Crear en AWS Console (Recomendado para principiantes)**

1. Abre [AWS Console](https://console.aws.amazon.com)
2. Navega a **EC2** → **Key Pairs** → **Create key pair**
3. Nombre: `el-campeon-prod`
4. Format: `.pem` (para SSH)
5. Click **Create** (descargará automáticamente)
6. Guarda en lugar seguro, por ejemplo: `~/.ssh/el-campeon-prod.pem`

```bash
# Proteger el archivo (solo macOS/Linux)
chmod 400 ~/.ssh/el-campeon-prod.pem
```

**Opción B: Crear con AWS CLI**

```bash
aws ec2 create-key-pair \
  --key-name el-campeon-prod \
  --region us-east-1 \
  --query 'KeyMaterial' \
  --output text > ~/.ssh/el-campeon-prod.pem

chmod 400 ~/.ssh/el-campeon-prod.pem
```

### Paso 2: Crear Security Group

Security Group es un firewall virtual que controla tráfico.

**Opción A: Por Console**

1. Navega a **EC2** → **Security Groups** → **Create Security Group**
2. Nombre: `el-campeon-prod-sg`
3. Descripción: "Security group for El Campeón Web production"
4. **VPC**: Selecciona tu VPC default
5. Agrega Inbound Rules:

| Protocol | Port  | Source      | Descripción        |
|----------|-------|-----------|-------------------|
| TCP      | 22    | Your IP   | SSH access        |
| TCP      | 80    | 0.0.0.0/0 | HTTP              |
| TCP      | 443   | 0.0.0.0/0 | HTTPS             |

6. Click **Create**

**Opción B: Con AWS CLI**

```bash
# Crear security group
SG_ID=$(aws ec2 create-security-group \
  --group-name el-campeon-prod-sg \
  --description "Security group for El Campeón Web production" \
  --region us-east-1 \
  --query 'GroupId' \
  --output text)

echo "Security Group ID: $SG_ID"

# Obtener tu IP pública
MY_IP=$(curl -s https://checkip.amazonaws.com)
echo "Your IP: $MY_IP"

# Agregar regla SSH (solo tu IP)
aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr $MY_IP/32 \
  --region us-east-1

# Agregar regla HTTP
aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0 \
  --region us-east-1

# Agregar regla HTTPS (opcional)
aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0 \
  --region us-east-1
```

### Paso 3: Lanzar Instancia EC2

**Opción A: Por Console (Más fácil para principiantes)**

1. Navega a **EC2** → **Instances** → **Launch Instances**
2. **Name**: `el-campeon-prod`
3. **AMI**: Busca y selecciona **Ubuntu 22.04 LTS**
4. **Instance Type**: `t3.small` (suficiente para dev/test)
   - t3.small: 2 vCPU, 2 GB RAM, ~$0.0208/hora
   - t3.medium: 2 vCPU, 4 GB RAM, ~$0.0416/hora
5. **Key pair**: `el-campeon-prod`
6. **Security Group**: `el-campeon-prod-sg`
7. **Storage**: 30 GB EBS (gp3 si disponible)
8. **Advanced Details** → **Monitoring**: Habilita CloudWatch detallado (opcional)
9. Click **Launch Instance**

**Opción B: Con AWS CLI**

```bash
# Usar AMI ID para Ubuntu 22.04 en us-east-1
AMI_ID="ami-0ac80df6eff0e70b5"  # Ubuntu 22.04 LTS (us-east-1)

aws ec2 run-instances \
  --image-id $AMI_ID \
  --instance-type t3.small \
  --key-name el-campeon-prod \
  --security-group-ids $SG_ID \
  --block-device-mappings 'DeviceName=/dev/sda1,Ebs={VolumeSize=30,VolumeType=gp3}' \
  --monitoring Enabled=true \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=el-campeon-prod}]' \
  --region us-east-1
```

### Paso 4: Obtener Public IP

```bash
# Espera 30 segundos a que la instancia esté running

aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=el-campeon-prod" \
  --query 'Reservations[0].Instances[0].PublicIpAddress' \
  --output text \
  --region us-east-1

# Respuesta: 54.123.45.67 (ejemplo)
# Guarda esta IP, la usarás en todo el tutorial
```

---

## Configurar Seguridad

### Paso 1: Conectar por SSH

```bash
# macOS/Linux:
ssh -i ~/.ssh/el-campeon-prod.pem ubuntu@<PUBLIC_IP>

# Ejemplo:
ssh -i ~/.ssh/el-campeon-prod.pem ubuntu@54.123.45.67

# Respuesta esperada:
# ubuntu@ip-172-31-xxx-xxx:~$

# Agregar a config para facilitar conexión futura
# Editar ~/.ssh/config:
cat >> ~/.ssh/config << EOF
Host el-campeon
    HostName <PUBLIC_IP>
    User ubuntu
    IdentityFile ~/.ssh/el-campeon-prod.pem
    ServerAliveInterval 60
EOF

# Luego simplemente usar:
ssh el-campeon
```

### Paso 2: Crear Usuario No-Root

Por seguridad, nunca ejecutes aplicaciones como root.

```bash
# Conectado en la instancia EC2

# Crear usuario deployer
sudo useradd -m -s /bin/bash deployer
sudo usermod -aG docker deployer  # Agregar a grupo docker (después de instalar)

# Crear contraseña segura
sudo passwd deployer

# Copiar SSH key del usuario ubuntu a deployer (opcional)
sudo rsync -av /home/ubuntu/.ssh/ /home/deployer/.ssh/
sudo chown -R deployer:deployer /home/deployer/.ssh/
sudo chmod 700 /home/deployer/.ssh
sudo chmod 600 /home/deployer/.ssh/authorized_keys

# Cambiar a usuario deployer
su - deployer

# Verificar
whoami  # Debería mostrar: deployer
```

### Paso 3: Actualizar Sistema

```bash
# Aún en EC2, como usuario ubuntu o deployer

# Actualizar paquetes del sistema
sudo apt-get update
sudo apt-get upgrade -y

# Instalar herramientas básicas
sudo apt-get install -y \
  curl \
  wget \
  git \
  htop \
  net-tools \
  vim \
  unzip \
  build-essential

# Opcional: Instalar Go (si quieres compilar localmente)
sudo apt-get install -y golang-1.21
```

---

## Instalar Dependencias en EC2

### Paso 1: Instalar Docker y Docker Compose

```bash
# Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Agregar usuario a grupo docker
sudo usermod -aG docker ubuntu
sudo usermod -aG docker deployer

# Necesario iniciar nueva sesión para que funcione el grupo
# O simplemente usar: docker con sudo

# Verificar instalación
docker --version
# Respuesta: Docker version 24.x.x

# Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verificar
docker-compose --version
# Respuesta: Docker Compose version 2.24.0
```

### Paso 2: Instalar Nginx (Reverse Proxy)

```bash
# Instalar Nginx
sudo apt-get install -y nginx

# Iniciar servicio
sudo systemctl start nginx
sudo systemctl enable nginx

# Verificar
sudo systemctl status nginx

# Abre en navegador: http://<PUBLIC_IP>
# Deberías ver la página de bienvenida de Nginx
```

---

## Desplegar la Aplicación

### Paso 1: Clonar Repositorio

```bash
# En la instancia EC2, como usuario deployer
cd /home/deployer

# Clonar repositorio
git clone https://github.com/nahuelmarianolosada/el-campeon-web.git
cd el-campeon-web

# Verificar que está el Dockerfile
ls -la dockerfiles/
ls -la docker-compose.yml
```

### Paso 2: Configurar Variables de Entorno

```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar variables para producción
nano .env

# Las variables importantes a cambiar:
```

**Contenido de `.env` para producción:**
```env
# Server
PORT=8080
ENV=production

# Database (Usaremos la BD local en Docker por ahora, después migrar a RDS)
DB_HOST=db
DB_PORT=3306
DB_USER=el_campeon_user
DB_PASSWORD=ChangeMe123!@#SuperSecure456
DB_NAME=el_campeon_web

# JWT Secrets (Generar con: openssl rand -base64 32)
JWT_SECRET_KEY=Base64EncodedString32CharsMinimum12345678
JWT_REFRESH_SECRET=AnotherBase64StringFor32Chars123456789
JWT_EXPIRY_HOURS=24

# MercadoPago (Obtener de tu dashboard en mercadopago.com)
MERCADOPAGO_ACCESS_TOKEN=YOUR_ACCESS_TOKEN_HERE
MERCADOPAGO_PUBLIC_KEY=YOUR_PUBLIC_KEY_HERE

# API (Usar IP pública)
API_BASE_URL=http://13.220.16.34
```

⚠️ **IMPORTANTE**: 
- El archivo `.env` **DEBE estar** en `/home/deployer/el-campeon-web/.env`
- Asegúrate de que NO está en `.gitignore` local pero SÍ en el repositorio (git lo ignora)
- Si la app dice "No .env file found", el archivo no está en el lugar correcto

**Generar secretos seguros:**
```bash
# Generar JWT_SECRET_KEY
openssl rand -base64 32

# Ejemplo de salida:
# K7jH/2qP9L+3nM8xQ/vW5sR4tY6uI7oP8aS9dF0gH1j=

# Repetir para JWT_REFRESH_SECRET
openssl rand -base64 32
```

⚠️ **GUARDAR ESTOS SECRETOS** - Ahora y después en AWS Secrets Manager.

### Paso 3: Ajustar docker-compose.yml para Producción

```bash
# Editar docker-compose.yml
nano docker-compose.yml
```

**Cambios importantes:**

```yaml
version: '3.8'

services:
  db:
    image: mysql:8.0
    container_name: el_campeon_db
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASSWORD}
    ports:
      - "127.0.0.1:3306:3306"  # Solo acceso local
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations/init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - el_campeon_network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 20s
      retries: 10
      interval: 10s
    restart: unless-stopped

  app:
    build:
      context: .
      dockerfile: dockerfiles/Dockerfile
    container_name: el_campeon_app
    environment:
      PORT: ${PORT}
      ENV: ${ENV}
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      JWT_SECRET_KEY: ${JWT_SECRET_KEY}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      JWT_EXPIRY_HOURS: ${JWT_EXPIRY_HOURS}
      MERCADOPAGO_ACCESS_TOKEN: ${MERCADOPAGO_ACCESS_TOKEN}
      MERCADOPAGO_PUBLIC_KEY: ${MERCADOPAGO_PUBLIC_KEY}
      API_BASE_URL: ${API_BASE_URL}
    ports:
      - "127.0.0.1:8080:8080"  # Solo localhost (Nginx accede desde aquí)
    depends_on:
      db:
        condition: service_healthy
    networks:
      - el_campeon_network
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  el_campeon_network:
    driver: bridge

volumes:
  mysql_data:
    driver: local
```

### Paso 4: Construir y Levantar Contenedores

```bash
# En /home/deployer/el-campeon-web

# Construir imágenes (primera vez tarda 5-10 minutos)
docker-compose build

# Levantar servicios en background
docker-compose up -d

# Verificar estado
docker-compose ps
# Deberías ver: db running + app running

# Ver logs de la app
docker-compose logs -f app

# Ver logs de la BD
docker-compose logs -f db

# Probar conexión local
curl http://127.0.0.1:8080/health
```

**Respuesta esperada:**
```json
{"status":"ok","service":"el-campeon-web"}
```

---

## Configurar Nginx

Nginx actúa como reverse proxy, redirigiendo tráfico HTTP a tu app en puerto 8080.

### Paso 1: Crear Configuración

```bash
# Crear archivo de configuración
sudo nano /etc/nginx/sites-available/el-campeon
```

**Archivo: `/etc/nginx/sites-available/el-campeon`**

```nginx
# HTTP Server (Sin SSL, acceso directo por IP)
server {
    listen 80;
    listen [::]:80;
    server_name _;  # Acepta cualquier servidor

    # Proxy reverso a la app
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Logs
    access_log /var/log/nginx/el-campeon-access.log;
    error_log /var/log/nginx/el-campeon-error.log;

    # Compresión
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
}
```

### Paso 2: Habilitar Configuración

```bash
# Crear symlink (para habilitar)
sudo ln -s /etc/nginx/sites-available/el-campeon /etc/nginx/sites-enabled/

# Deshabilitar default si es necesario
sudo rm /etc/nginx/sites-enabled/default

# Verificar sintaxis
sudo nginx -t

# Si hay errores, revisar y corregir
# Si dice "successful", continuar

# Recargar Nginx
sudo systemctl reload nginx
```

### Paso 3: Permitir Firewall de Ubuntu si está habilitado

```bash
# Ver estado
sudo ufw status

# Si está habilitado (Status: active):
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp

# Recargar
sudo ufw reload
```

---

## Probar en Producción

### Prueba 1: Acceso HTTP

```bash
# Desde tu máquina local (NO en EC2)

# 1. Acceso a health check
curl -s http://13.220.16.34/health | jq .

# Respuesta esperada:
{
  "status": "ok",
  "service": "el-campeon-web"
}

# 2. En navegador
# Abre: http://54.123.45.67
```

### Prueba 2: Registro de Usuario

```bash
# Crear usuario de prueba
curl -X POST http://13.220.16.34/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-prod@example.com",
    "first_name": "Test",
    "last_name": "Production",
    "password": "SecurePass123!",
    "phone": "+5491123456789"
  }' | jq .

# Respuesta esperada (guardá los tokens):
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": 1,
    "email": "test-prod@example.com",
    "first_name": "Test"
  }
}

# Guardar access_token para pruebas
export ACCESS_TOKEN="eyJhbGc..."
```

### Prueba 3: Login

```bash
curl -X POST http://13.220.16.34/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-prod@example.com",
    "password": "SecurePass123!"
  }' | jq .
```

### Prueba 4: Listar Productos

```bash
curl http://13.220.16.34/api/products | jq .

# Con token (si es necesario proteger):
curl http://13.220.16.34/api/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

### Prueba 5: Carrito de Compras

```bash
# Obtener carrito actual
curl http://13.220.16.34/api/cart \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# Agregar item (primero obtener ID de producto)
curl -X POST http://13.220.16.34/api/cart/items \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }' | jq .
```

### Prueba 6: Verificar Base de Datos

```bash
# En la instancia EC2
docker-compose exec db mysql -u el_campeon_user -p el_campeon_web -e "SELECT * FROM users;"

# Debería mostrar el usuario creado
```

### Prueba 7: Revisar Logs

```bash
# Logs de aplicación
docker-compose logs --tail=100 app

# Logs de Nginx
sudo tail -f /var/log/nginx/el-campeon-access.log
sudo tail -f /var/log/nginx/el-campeon-error.log

# Logs del sistema
journalctl -u nginx -n 50
```

### Prueba 8: Test de Performance

```bash
# Instalar herramienta de load testing (en máquina local)
# macOS:
brew install ab

# Ubuntu:
sudo apt-get install apache2-utils

# Hacer 100 requests
ab -n 100 -c 10 http://13.220.16.34/health

# Respuesta esperada: ~100 requests exitosos
```

---

## Monitoreo y Mantenimiento

### Monitoreo de Aplicación

```bash
# Ver estado de contenedores
docker-compose ps

# Ver consumo de recursos
docker stats

# Ver logs en tiempo real
docker-compose logs -f

# Revisar un contenedor específico
docker-compose logs -f app --tail=50
```

### Backups Automáticos

```bash
# Crear script de backup
sudo nano /usr/local/bin/backup-el-campeon.sh

# Contenido:
```

```bash
#!/bin/bash

BACKUP_DIR="/backups/el-campeon"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_BACKUP="$BACKUP_DIR/db_backup_$TIMESTAMP.sql"

mkdir -p $BACKUP_DIR

# Backup de BD
docker-compose -f /home/deployer/el-campeon-web/docker-compose.yml exec -T db \
  mysqldump -u el_campeon_user -pChangeMe123!@#SuperSecure456 el_campeon_web > $DB_BACKUP

# Comprimir
gzip $DB_BACKUP

# Mantener últimos 7 días
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Backup completado: $DB_BACKUP.gz"
```

```bash
# Hacer executable y programar con cron
sudo chmod +x /usr/local/bin/backup-el-campeon.sh

# Editar crontab
sudo crontab -e

# Agregar para ejecutar diariamente a las 2 AM:
0 2 * * * /usr/local/bin/backup-el-campeon.sh >> /var/log/el-campeon-backup.log 2>&1
```

### Monitoreo de CloudWatch

```bash
# Instalar CloudWatch agent
cd /tmp
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
sudo apt-get install ./amazon-cloudwatch-agent.deb

# Configurar agent
sudo nano /opt/aws/amazon-cloudwatch-agent/etc/config.json

# Contenido básico:
```

```json
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/nginx/el-campeon-access.log",
            "log_group_name": "/el-campeon/nginx/access",
            "log_stream_name": "{instance_id}"
          },
          {
            "file_path": "/var/log/nginx/el-campeon-error.log",
            "log_group_name": "/el-campeon/nginx/error",
            "log_stream_name": "{instance_id}"
          }
        ]
      }
    }
  },
  "metrics": {
    "metrics_collected": {
      "cpu": {"measurement": [{"name": "cpu_usage_idle", "rename": "CPU_IDLE", "unit": "Percent"}], "metrics_collection_interval": 60},
      "mem": {"measurement": [{"name": "mem_used_percent", "rename": "MEM_USED", "unit": "Percent"}], "metrics_collection_interval": 60}
    }
  }
}
```

```bash
# Iniciar agent
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -s \
  -c file:/opt/aws/amazon-cloudwatch-agent/etc/config.json
```

### Chequeo de Salud Regular

```bash
# Ver CPU y memoria
free -h
df -h

# Procesos ejecutándose
docker ps

# Conectividad de BD
docker-compose exec db mysql -u el_campeon_user -pChangeMe123!@#SuperSecure456 -e "SELECT 1;"

# Estado de firewall
sudo ufw status
```

### Actualizar Aplicación

```bash
# 1. Hacer pull del código nuevo
cd /home/deployer/el-campeon-web
git pull origin main

# 2. Reconstruir imagen Docker
docker-compose build

# 3. Reiniciar servicios
docker-compose up -d

# 4. Verificar
docker-compose logs -f app
curl http://13.220.16.34/health
```

---

## Solución de Problemas Comunes

### Problema: App no inicia

```bash
# Ver logs
docker-compose logs app

# Causas comunes:
# 1. Puerto en uso: Cambiar puerto en docker-compose.yml
# 2. Variables de entorno incompletas: Verificar .env
# 3. BD no accesible: Verificar estado de BD

# Reintentar
docker-compose down
docker-compose up -d
```

### Problema: No puedo conectar por SSH

```bash
# Verificar security group
aws ec2 describe-security-groups --group-id <SG_ID>

# Verificar que tu IP está autorizada
# Si cambió tu IP, agregar nueva regla:
aws ec2 authorize-security-group-ingress \
  --group-id <SG_ID> \
  --protocol tcp \
  --port 22 \
  --cidr <TU_IP>/32

# Probar conexión
ssh -v -i ~/.ssh/el-campeon-prod.pem ubuntu@<PUBLIC_IP>
```

### Problema: Nginx no redirecciona

```bash
# Verificar sintaxis de Nginx
sudo nginx -t

# Ver estado de Nginx
sudo systemctl status nginx

# Revisar configuración
sudo cat /etc/nginx/sites-enabled/el-campeon

# Recargar Nginx
sudo systemctl reload nginx

# Ver logs de Nginx
sudo tail -f /var/log/nginx/el-campeon-error.log
```

### Problema: Out of Memory

```bash
# Ver uso de memoria
free -h

# Detener contenedores innecesarios
docker-compose stop

# Limpiar volúmenes no usados
docker volume prune

# Aumentar swap (temporal)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### Problema: "No .env file found" pero la app funciona

Esto ocurre cuando el archivo `.env` no está en el directorio correcto o no se está leyendo correctamente.

**Solución:**

```bash
# Verificar que el .env existe
cd /home/deployer/el-campeon-web
ls -la .env

# Si no existe, crearlo:
cp .env.example .env

# Editar con tus valores
nano .env

# Asegurar que contiene (mínismo):
API_BASE_URL=http://13.220.16.34
DB_HOST=db
DB_PORT=3306
DB_USER=el_campeon_user
DB_PASSWORD=ChangeMe123!@#SuperSecure456
DB_NAME=el_campeon_web
PORT=8080
ENV=production
JWT_SECRET_KEY=<algo random>
JWT_REFRESH_SECRET=<algo random>
JWT_EXPIRY_HOURS=24

# Reconstruir sin caché
docker-compose down
docker system prune -f
docker-compose build --no-cache

# Levantar nuevamente
docker-compose up -d
```

### Problema: Nginx escucha solo en localhost (127.0.0.1:80)

Si Nginx está escuchando solo en localhost, no será accesible desde el exterior.

**Síntomas:**
```
LISTEN 0 128 127.0.0.1:80 0.0.0.0:*
```

**Solución:**

Verificar que `/etc/nginx/sites-available/el-campeon` tiene:

```nginx
server {
    listen 80;           # ← Debe estar SIN 127.0.0.1
    listen [::]:80;
    server_name _;

    # ... rest of config
}
```

**NO debe tener:**
```nginx
listen 127.0.0.1:80;  # ← INCORRECTO
```

Después de corregir:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

### Problema: Timeout al acceder desde máquina local

Si obtienes timeout al hacer curl desde tu máquina:

```bash
# Verificar que el Security Group permite puerto 80
aws ec2 describe-security-groups \
  --group-id sg-05cd98748396b25e3 \
  --region us-east-1 \
  --query 'SecurityGroups[0].IpPermissions'

# Debe permitir puerto 80 para 0.0.0.0/0 o tu IP
```

Si no está permitido, agregar regla:

```bash
aws ec2 authorize-security-group-ingress \
  --group-id sg-05cd98748396b25e3 \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0 \
  --region us-east-1
```

### Problema: Nginx sí funciona pero app no responde

Si accedes al IP y obtenes conexión rechazada:

```bash
# En la instancia EC2, verificar que los contenedores están running
docker-compose ps

# Si no están corriendo:
docker-compose up -d

# Ver logs de la app
docker-compose logs --tail=50 app

# Probar localmente en la instancia
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1/health
```

---

## Checklist Final

Antes de considerar esto en "producción":

- [ ] Aplicación accesible por IP pública
- [ ] Base de datos accesible
- [ ] Todos los endpoints probados
- [ ] JWT tokens funcionando
- [ ] Backups automáticos configurados
- [ ] Logs siendo monitoreados
- [ ] Security group restrictivo
- [ ] Usuarios sin permisos root
- [ ] Firewall habilitado y configurado
- [ ] Monitoreo habilitado (CloudWatch)
- [ ] Proceso de deploy documentado
- [ ] Plan de disaster recovery
- [ ] Alertas configuradas

---

## Próximos Pasos Recomendados

1. **SSL/HTTPS**: Agregar certificado Let's Encrypt si quieres dominio
2. **Base de Datos**: Migrar a AWS RDS para mejor backup y escalabilidad
3. **Load Balancer**: Usar AWS ALB para distribuir tráfico
4. **Auto Scaling**: Configurar grupo de auto-scaling
5. **CDN**: Usar CloudFront para servir static assets
6. **Monitoring**: Configurar dashboards en CloudWatch
7. **CI/CD**: Automatizar deploy con GitHub Actions
8. **Secrets Management**: Usar AWS Secrets Manager
9. **Logging Centralizado**: Usar CloudWatch Logs

---

## Conclusión

¡Felicidades! Has desplegado "El Campeón Web" en AWS EC2 con:
- ✅ Aplicación corriendo en Docker
- ✅ Base de datos MySQL
- ✅ Nginx como reverse proxy
- ✅ Acceso por IP pública HTTP
- ✅ Acceso seguro por SSH
- ✅ Backups automáticos

Para acceder a tu aplicación:
```
http://13.220.16.34/
```

Para ayuda adicional:
- AWS Documentation: https://docs.aws.amazon.com/
- Docker Documentation: https://docs.docker.com/
- Nginx Documentation: https://nginx.org/en/docs/

¡Éxito con tu aplicación! 🚀

---

**Última Actualización**: Mayo 2026
**Versión**: 1.0 (IP Only)

