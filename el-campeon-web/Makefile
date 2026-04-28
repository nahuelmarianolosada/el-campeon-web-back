.PHONY: help build run run-docker stop down test test-coverage fmt lint clean docker-build

# Variables
APP_NAME=el-campeon-web
DOCKER_COMPOSE=docker-compose
GO=go
GOFLAGS=-v

# Colors for output
YELLOW=\033[0;33m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help:
	@echo "$(GREEN)Comandos disponibles:$(NC)"
	@echo "  $(YELLOW)make build$(NC)           - Compilar la aplicación"
	@echo "  $(YELLOW)make run$(NC)             - Ejecutar localmente"
	@echo "  $(YELLOW)make run-docker$(NC)      - Ejecutar con Docker Compose"
	@echo "  $(YELLOW)make stop$(NC)            - Detener Docker Compose"
	@echo "  $(YELLOW)make down$(NC)            - Detener y eliminar contenedores"
	@echo "  $(YELLOW)make test$(NC)            - Ejecutar tests"
	@echo "  $(YELLOW)make test-coverage$(NC)   - Tests con coverage"
	@echo "  $(YELLOW)make fmt$(NC)             - Formatear código"
	@echo "  $(YELLOW)make lint$(NC)            - Ejecutar linter"
	@echo "  $(YELLOW)make clean$(NC)           - Limpiar binarios y temporales"
	@echo "  $(YELLOW)make docker-build$(NC)    - Construir imagen Docker"
	@echo "  $(YELLOW)make db-init$(NC)         - Inicializar base de datos"
	@echo "  $(YELLOW)make logs$(NC)            - Ver logs de Docker"

## Build y Run

build:
	@echo "$(GREEN)Compilando...$(NC)"
	@$(GO) build -o ./bin/$(APP_NAME) ./cmd/main.go
	@echo "$(GREEN)✓ Compilación completada: ./bin/$(APP_NAME)$(NC)"

run: build
	@echo "$(GREEN)Ejecutando $(APP_NAME)...$(NC)"
	@./bin/$(APP_NAME)

run-docker:
	@echo "$(GREEN)Ejecutando con Docker Compose...$(NC)"
	@$(DOCKER_COMPOSE) up

run-docker-background:
	@echo "$(GREEN)Ejecutando Docker Compose en background...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✓ Servicios levantados. Usa 'make logs' para ver logs$(NC)"

## Docker

docker-build:
	@echo "$(GREEN)Construyendo imagen Docker...$(NC)"
	@$(DOCKER_COMPOSE) build

docker-build-no-cache:
	@echo "$(GREEN)Construyendo imagen Docker (sin caché)...$(NC)"
	@$(DOCKER_COMPOSE) build --no-cache

.PHONY: stop down

stop:
	@echo "$(GREEN)Deteniendo servicios...$(NC)"
	@$(DOCKER_COMPOSE) stop
	@echo "$(GREEN)✓ Servicios detenidos$(NC)"

down:
	@echo "$(GREEN)Deteniendo y eliminando contenedores...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✓ Contenedores eliminados$(NC)"

down-volumes:
	@echo "$(RED)Eliminando contenedores y volúmenes...$(NC)"
	@$(DOCKER_COMPOSE) down -v
	@echo "$(GREEN)✓ Completado$(NC)"

logs:
	@$(DOCKER_COMPOSE) logs -f app

logs-db:
	@$(DOCKER_COMPOSE) logs -f db

ps:
	@$(DOCKER_COMPOSE) ps

## Database

db-init:
	@echo "$(GREEN)Inicializando base de datos...$(NC)"
	@$(DOCKER_COMPOSE) exec db mysql -u root -proot_password_change_me < migrations/init.sql
	@echo "$(GREEN)✓ Base de datos inicializada$(NC)"

db-shell:
	@$(DOCKER_COMPOSE) exec db mysql -u el_campeon_user -puser_password_change_me el_campeon_web

db-logs:
	@$(DOCKER_COMPOSE) logs -f db

## Testing

test:
	@echo "$(GREEN)Ejecutando tests...$(NC)"
	@$(GO) test $(GOFLAGS) ./...

test-short:
	@echo "$(GREEN)Ejecutando tests (corto)...$(NC)"
	@$(GO) test -short ./...

test-coverage:
	@echo "$(GREEN)Ejecutando tests con coverage...$(NC)"
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

test-services:
	@echo "$(GREEN)Tests de servicios...$(NC)"
	@$(GO) test $(GOFLAGS) ./internal/services/...

test-repositories:
	@echo "$(GREEN)Tests de repositorios...$(NC)"
	@$(GO) test $(GOFLAGS) ./internal/repositories/...

test-verbose:
	@echo "$(GREEN)Ejecutando tests (verbose)...$(NC)"
	@$(GO) test -v ./...

## Code Quality

fmt:
	@echo "$(GREEN)Formateando código...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)✓ Código formateado$(NC)"

vet:
	@echo "$(GREEN)Ejecutando vet...$(NC)"
	@$(GO) vet ./...
	@echo "$(GREEN)✓ Vet completado$(NC)"

lint:
	@echo "$(GREEN)Ejecutando linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)Instalando golangci-lint...$(NC)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...
	@echo "$(GREEN)✓ Linter completado$(NC)"

## Utilities

health-check:
	@echo "$(GREEN)Verificando estado del servidor...$(NC)"
	@curl -s http://localhost:8080/health | jq .
	@echo "$(GREEN)✓ Servidor disponible$(NC)"

dependencies:
	@echo "$(GREEN)Descargando dependencias...$(NC)"
	@$(GO) mod download
	@$(GO) mod verify
	@echo "$(GREEN)✓ Dependencias descargadas$(NC)"

update-deps:
	@echo "$(GREEN)Actualizando dependencias...$(NC)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencias actualizadas$(NC)"

tidy:
	@echo "$(GREEN)Limpiando módulos...$(NC)"
	@$(GO) mod tidy
	@echo "$(GREEN)✓ Módulos limpios$(NC)"

## Cleanup

clean:
	@echo "$(GREEN)Limpiando...$(NC)"
	@$(GO) clean
	@rm -rf ./bin/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

clean-docker:
	@echo "$(RED)ADVERTENCIA: Esto eliminará volúmenes de BD$(NC)"
	@read -p "¿Continuar? (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) down -v; \
		echo "$(GREEN)✓ Limpieza Docker completada$(NC)"; \
	fi

## Development

install-tools:
	@echo "$(GREEN)Instalando herramientas de desarrollo...$(NC)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/cosmtrek/air@latest
	@echo "$(GREEN)✓ Herramientas instaladas$(NC)"

dev:
	@echo "$(GREEN)Ejecutando en modo desarrollo (con hot reload)...$(NC)"
	@which air > /dev/null || (echo "$(YELLOW)Instalando air...$(NC)" && go install github.com/cosmtrek/air@latest)
	@air

## Production

build-release:
	@echo "$(GREEN)Compilando para producción...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o ./bin/$(APP_NAME)-linux ./cmd/main.go
	@echo "$(GREEN)✓ Build release completado: ./bin/$(APP_NAME)-linux$(NC)"

## Info

version:
	@$(GO) version

env:
	@echo "$(GREEN)Variables de entorno requeridas:$(NC)"
	@cat .env.example

info:
	@echo "$(GREEN)Información del proyecto:$(NC)"
	@echo "  Nombre: $(APP_NAME)"
	@echo "  Go Version: $$($(GO) version)"
	@echo "  Docker: $$(docker --version)"
	@echo "  Docker Compose: $$(docker-compose --version)"

## Default

.DEFAULT_GOAL := help

