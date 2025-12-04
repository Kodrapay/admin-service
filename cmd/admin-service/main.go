package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kodra-pay/admin-service/internal/config"
	"github.com/kodra-pay/admin-service/internal/middleware"
	"github.com/kodra-pay/admin-service/internal/routes"
)

func main() {
	cfg := config.Load("admin-service", "7003")

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	// Enable CORS for frontend access
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins:     "http://localhost:5173, http://localhost:5174, http://127.0.0.1:5173, http://127.0.0.1:5174, http://localhost:3000, http://127.0.0.1:3000",
	// 	AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
	// 	AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	// 	AllowCredentials: true,
	// }))

	app.Use(middleware.RequestID())

	routes.Register(app, cfg.ServiceName, cfg.MerchantServiceURL)

	log.Printf("%s listening on :%s", cfg.ServiceName, cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
