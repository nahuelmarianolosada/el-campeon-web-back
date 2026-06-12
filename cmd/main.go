package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/database"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/handlers"
)

func init() {
	// Cargar variables de entorno desde .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func main() {
	// En producción los valores sensibles vienen de SSM Parameter Store, no
	// del .env en disco. Se activa con SECRETS_PROVIDER=ssm (o ENV=production)
	// y el prefijo se controla con SSM_PATH_PREFIX (default /el-campeon/prod/).
	if useSSM() {
		prefix := os.Getenv("SSM_PATH_PREFIX")
		if prefix == "" {
			prefix = "/el-campeon/prod/"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := config.LoadSecretsFromSSM(ctx, prefix); err != nil {
			log.Fatalf("Failed to load secrets from SSM: %v", err)
		}
	}

	// Cargar configuración
	cfg := config.Load()

	// Inicializar base de datos
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Configurar modo de Gin
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Crear router
	router := gin.Default()

	// Middleware CORS
	router.Use(corsMiddleware())

	// Setup routes
	handlers.SetupRoutes(router, db, cfg)

	// Iniciar servidor
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8080"
	}

	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func useSSM() bool {
	switch os.Getenv("SECRETS_PROVIDER") {
	case "ssm":
		return true
	case "env", "none":
		return false
	}
	return os.Getenv("ENV") == "production"
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Guest-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
