package internal

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/internal/routes"
	"github.com/jump-fortress/site/slogger"
	"github.com/rotisserie/eris"
	"github.com/rs/cors"
)

var (
	api                          huma.API
	sessionCookieSecurityMap     = []map[string][]string{{"Steam": {}}}
	requireSessionMiddlewares    huma.Middlewares
	requireConsultantMiddlewares huma.Middlewares
	requireModMiddlewares        huma.Middlewares
	requireAdminMiddlewares      huma.Middlewares
	requireDevMiddlewares        huma.Middlewares
)

func setupRouter() (*chi.Mux, error) {
	// TUTORIAL: we're configuring our logger to always log in the ECS v9 schema.
	//           this schema is easily consumed by many observability/telemetry tools.
	logConcise := env.GetBool("JUMP_HTTPLOG_CONCISE")
	logFormat := httplog.SchemaECS.Concise(logConcise)

	logLevel, matchedErr := env.GetMapped("JUMP_HTTPLOG_LEVEL", slogger.SlogLevelMap)
	if matchedErr != nil {
		return nil, matchedErr
	}

	handlerOptions := &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
		Level:       logLevel,
	}

	var logger *slog.Logger
	slogMode := env.GetString("JUMP_HTTPLOG_MODE")
	switch slogMode {
	case "Text":
		logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	case "JSON":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	default:
		return nil, eris.Errorf("invalid value for JUMP_HTTPLOG_Mode: %s", slogMode)
	}

	router := chi.NewMux()

	options := &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: logLevel,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: logFormat,

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,

		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  env.GetList("JUMP_HTTPLOG_REQUEST_HEADERS"),
		LogResponseHeaders: env.GetList("JUMP_HTTPLOG_RESPONSE_HEADERS"),
	}

	if env.GetBool("JUMP_HTTPLOG_REQUEST_BODIES") {
		options.LogRequestBody = func(r *http.Request) bool { return true }
	}

	if env.GetBool("JUMP_HTTPLOG_RESPONSE_BODIES") {
		options.LogResponseBody = func(r *http.Request) bool { return true }
	}

	// TUTORIAL: every request should get logged using our configured logger
	router.Use(httplog.RequestLogger(logger, options))

	// todo (spiritov): use strict `AllowedOrigins`
	// todo (spiritov): use CSRF middleware (?)
	// todo (spiritov): rate limit
	router.Use(cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"}, // default value
	}).Handler)

	return router, nil
}

func setupHumaConfig() huma.Config {
	config := huma.DefaultConfig("Jump Stuff API", "1.0.0")

	// steam security scheme, a JWT with user's OpenID information
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"Steam": {
			Type:        "apiKey",
			In:          "cookie",
			Description: "a session cookie stores the user's session token.",
			Name:        SessionCookieName,
		},
	}
	return config
}

// A readiness endpoint is important - it can be used to inform your infrastructure
// (e.g. fly.io) that the API is available. Readiness checks can help keep your API
// alive, by informing fly on when it should try restarting a machine in case of a
// crash.
func registerHealthCheck(internalApi *huma.Group) {
	type ReadyResponse struct{ OK bool }

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/readyz",
		OperationID: "readyz",
		Summary:     "get readiness",
		Description: "get whether or not the API is ready to process requests",
	}, func(ctx context.Context, _ *struct{}) (*ReadyResponse, error) {
		return &ReadyResponse{OK: true}, nil
	})
}

func ServeAPI(ctx context.Context, address string) {
	router, err := setupRouter()
	if err != nil {
		slog.Error("failed to setup router", "error", err)
		log.Fatal()
	}
	config := setupHumaConfig()
	api = humachi.New(router, config)

	registerRoutes()

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	go func() {
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("failed to serve api", "error", err)
			log.Fatal()
		}
	}()

	// done
	<-ctx.Done()
	doneCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(doneCtx)
	slog.Info("api shutdown")
}

func registerRoutes() {
	// create api groups and middlewares
	internalApi := huma.NewGroup(api, "/internal")
	sessionApi := huma.NewGroup(internalApi, "/session")
	modApi := huma.NewGroup(internalApi, "/mod")
	consultantApi := huma.NewGroup(internalApi, "/consultant")
	adminApi := huma.NewGroup(internalApi, "/admin")
	devApi := huma.NewGroup(internalApi, "/dev")

	requireSessionMiddlewares = huma.Middlewares{AuthHandler, RequireUserAuthHandler(internalApi)}
	requireConsultantMiddlewares = huma.Middlewares{AuthHandler, RequireConsultantHandler(consultantApi)}
	requireModMiddlewares = huma.Middlewares{AuthHandler, RequireModHandler(modApi)}
	requireAdminMiddlewares = huma.Middlewares{AuthHandler, RequireAdminHandler(adminApi)}
	requireDevMiddlewares = huma.Middlewares{AuthHandler, RequireDevHandler(devApi)}

	sessionApi.UseMiddleware(requireSessionMiddlewares...)
	consultantApi.UseMiddleware(requireConsultantMiddlewares...)
	modApi.UseMiddleware(requireModMiddlewares...)
	adminApi.UseMiddleware(requireAdminMiddlewares...)
	devApi.UseMiddleware(requireDevMiddlewares...)

	// register essential routes
	registerHealthCheck(internalApi)
	registerAuth(internalApi, sessionApi)

	// register all other routes
	routes.RegisterOpenRoutes(internalApi)
	routes.RegisterSessionRoutes(sessionApi)
	routes.RegisterConsultantRoutes(consultantApi)
	routes.RegisterModRoutes(modApi)
	routes.RegisterAdminRoutes(adminApi)
	routes.RegisterDevRoutes(devApi)
}
