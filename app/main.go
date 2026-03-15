package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/infosec554/clean-archtectura/config"
	_ "github.com/infosec554/clean-archtectura/docs"
	"github.com/infosec554/clean-archtectura/internal/repository/postgres"
	"github.com/infosec554/clean-archtectura/internal/rest"
	"github.com/infosec554/clean-archtectura/internal/rest/middleware"
	"github.com/infosec554/clean-archtectura/pkg/cache"
	"github.com/infosec554/clean-archtectura/pkg/token"
	user_service "github.com/infosec554/clean-archtectura/service/user"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	logger := zerolog.New(os.Stdout)
	zerolog.ErrorFieldName = "error"

	logger.Info().Any("config", cfg).Msg("loading config")

	c := cache.NewCache(cfg)

	store, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Storage init error: %v", err)
	}
	defer store.Close()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.SetRequestContextWithTimeout(cfg.RedisTTL))

	api := e.Group("/api/v1")

	jwtManager := token.NewJWTManager(cfg.JWTSecretKey)

	public := api.Group("")
	authGroup := api.Group("")

	m := middleware.NewMiddleware(cfg.JWTSecretKey, logger)

	authGroup.Use(m.JWTAuth())
	{

		userRepo := postgres.NewUserRepository(store.DB, logger)
		userService := user_service.NewUserService(userRepo, cfg, c, logger, jwtManager)
		rest.NewUserHandler(public, authGroup, userService, cfg, c, logger)

	}

	// Redirect /api/v1/docs to /api/swagger/index.html
	e.GET("/api/v1/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/api/swagger/index.html")
	})

	e.GET("/api/swagger/*", echoSwagger.WrapHandler)

	api.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "✅ OK")
	})

	for _, r := range e.Routes() {
		log.Printf("%s %s", r.Method, r.Path)
	}

	log.Printf("🚀 %s running on %s", cfg.AppName, cfg.AppPort)
	e.Logger.Fatal(e.Start(cfg.AppPort))
}
