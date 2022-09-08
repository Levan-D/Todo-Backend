package main

import (
	"fmt"
	_ "github.com/Levan-D/Todo-Backend/docs/app"
	internal "github.com/Levan-D/Todo-Backend/internal/app"
	"github.com/Levan-D/Todo-Backend/pkg/cli"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/database/postgres"
	"github.com/Levan-D/Todo-Backend/pkg/storage"
	_ "github.com/Levan-D/Todo-Backend/pkg/storage"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gobuffalo/packr/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
)

// @title Todo Platform App API
// @version v1
// @description API Server for Todo Platform

// @contact.name API Support
// @contact.url https://todo.lan
// @contact.email info@todo.lan

// @host todo.sns.ge
// @BasePath /api/v1/
// @query.collection.format multi

// @schemes http

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization

func main() {
	// Allow usage all cpu for service
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set service module
	config.Initialize(config.MODULE_APP)
	storage.Initialize()

	// Migrations Box
	config.SetBox(packr.New("migrations", "../../migrations"))

	// Initialize CLI module
	cli.Run(func() {
		// Instance
		app := fiber.New(fiber.Config{
			Prefork:               false,
			UnescapePath:          true,
			ServerHeader:          "Todo",
			DisableStartupMessage: false,
			BodyLimit:             20 * 1024 * 1024, // 20MB
		})

		// Gracefully shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			_ = <-c
			fmt.Println("Gracefully shutting down...")
			_ = app.Shutdown()
		}()

		// Middleware
		app.Use(recover.New())
		app.Use(logger.New())
		app.Use(requestid.New())
		app.Use(etag.New())
		app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
		app.Use(cors.New(cors.Config{
			Next:             nil,
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
			AllowHeaders:     "*",
			AllowCredentials: true,
			ExposeHeaders:    "",
			MaxAge:           0,
		}))

		// Database
		db, err := postgres.NewClient()
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}

		// Handlers & Routes
		internal.Initialize(app, db)

		// Initialize Docs
		app.Get("/api/v1/docs/*", swagger.Handler)

		// Start server
		addr := fmt.Sprintf("%s:%d", config.Get().Server.Host, config.Get().Server.Port)
		if err := app.Listen(addr); err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	})
}
