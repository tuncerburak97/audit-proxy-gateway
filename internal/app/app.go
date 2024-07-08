package app

import (
	"audit-proxy-gateway/internal/client"
	_ "audit-proxy-gateway/internal/client"
	"audit-proxy-gateway/internal/config"
	"audit-proxy-gateway/internal/proxy"
	logger2 "audit-proxy-gateway/pkg/logger"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func InitApp() *fiber.App {
	config.LoadConfig()
	logger2.InitLogger()

	log := logger2.GetLogger()
	log.Info("Logger initialized")

	client.InitClient()

	app := fiber.New()
	app.All("/*", proxy.ReverseProxy)

	return app
}

func StartApp(app *fiber.App) {
	cfg := config.GetConfig()
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	log := logger2.GetLogger()
	log.Infof("Starting server on %s...", address)
	if err := app.Listen(address); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
