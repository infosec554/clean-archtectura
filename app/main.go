package main

import (
	"context"
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
	admin_service "github.com/infosec554/clean-archtectura/service/admin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	logger := zerolog.New(os.Stdout)
	zerolog.ErrorFieldName = "error"

	logger.Info().Str("app", cfg.AppName).Msg("starting")

	c := cache.NewCache(cfg)

	store, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Fatalf("storage init error: %v", err)
	}
	defer store.Close()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.SetRequestContextWithTimeout(cfg.RedisTTL))

	api := e.Group("/api/v1")

	jwtManager := token.NewJWTManager(cfg.JWTSecretKey)

	// Route groups
	public     := api.Group("")
	private    := api.Group("")
	superadmin := api.Group("")

	m := middleware.NewMiddleware(cfg.JWTSecretKey, logger)
	private.Use(m.JWTAuth())
	superadmin.Use(m.JWTAuth(), middleware.SuperAdminOnly())

	// Admin
	adminRepo    := postgres.NewAdminRepository(store.DB, logger)
	adminSvc     := admin_service.NewAdminService(adminRepo, cfg, c, logger, jwtManager)
	rest.NewAdminHandler(public, private, superadmin, adminSvc, logger)

	// Swagger
	e.GET("/api/v1/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/api/swagger/index.html")
	})
	e.GET("/api/swagger/*", echoSwagger.WrapHandler)

	api.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	for _, r := range e.Routes() {
		log.Printf("%s %s", r.Method, r.Path)
	}

	log.Printf("%s running on %s", cfg.AppName, cfg.AppPort)
	e.Logger.Fatal(e.Start(cfg.AppPort))
}
