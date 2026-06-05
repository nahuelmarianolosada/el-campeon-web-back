# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar archivos de módulo
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación (binario estático para alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copiar el binario del stage de build
COPY --from=builder /app/main .

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]

