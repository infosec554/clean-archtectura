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

var publicRoutes = map[string]bool{
	"/api/v1/health":              true,
	"/api/v1/docs":                true,
	"/api/v1/dual_education.yaml": true,
}

var swaggerFile []byte

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

	addDoc(e)
	public := api.Group("")
	authGroup := api.Group("")

	m := middleware.NewMiddleware(cfg.JWTSecretKey, logger)

	authGroup.Use(m.JWTAuth())
	{

		userRepo := postgres.NewUserRepository(store.DB, logger)
		userService := user_service.NewUserService(userRepo, cfg, c, logger, jwtManager)
		rest.NewUserHandler(public, authGroup, userService, cfg, c, logger)

	}

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

func addDoc(e *echo.Echo) {
	var swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
    <title>API Docs</title>
    <meta charset="UTF-8">
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: "/api/v1/dual_education.yaml",
            dom_id: '#swagger-ui'
        });
    </script>
</body>
</html>`

	e.GET("/api/v1/dual_education.yaml", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/yaml", swaggerFile)
	})

	// Serve Swagger UI
	e.GET("/api/v1/docs", func(c echo.Context) error {
		return c.HTML(http.StatusOK, swaggerUIHTML)
	})
}
