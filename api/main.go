package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/internal"
	"github.com/jump-fortress/site/slogger"
	"github.com/jump-fortress/site/tasks"
	"github.com/jump-fortress/site/tasks/client"
	"github.com/jump-fortress/site/tasks/workers"
)

func main() {
	// env is a wrapper around the `godotenv` library
	err := env.Load("JUMP_ENV")
	if err != nil {
		slog.Error("error loading .env", "error", err)
		log.Fatal()
	}

	env.Require(
		"JUMP_DB_PATH",
		"JUMP_SLOG_LEVEL",
		"JUMP_SLOG_MODE",
		"JUMP_HTTP_ADDRESS",
		"JUMP_HTTPLOG_LEVEL",
		"JUMP_HTTPLOG_MODE",
		"JUMP_HTTPLOG_CONCISE",
		"JUMP_HTTPLOG_REQUEST_HEADERS",
		"JUMP_HTTPLOG_RESPONSE_HEADERS",
		"JUMP_HTTPLOG_REQUEST_BODIES",
		"JUMP_HTTPLOG_RESPONSE_BODIES",
		"JUMP_SESSION_TOKEN_SECRET",
		"JUMP_STEAM_API_KEY",
		"JUMP_OID_REALM",
		"JUMP_OID_CALLBACK_URL",
		"JUMP_UPSTASH_REDIS_REST_URL",
		"JUMP_UPSTASH_REDIS_REST_PASS",
		"JUMP_WEBHOOK_TEST_URL",
		"JUMP_WEBHOOK_EVENTS_URL",
	)

	err = slogger.Setup()
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(slogger.Logger)

	dbPath := env.GetString("JUMP_DB_PATH")
	database := db.OpenDB(fmt.Sprintf("%s?_foreign_keys=on", dbPath))
	defer database.Close()

	var foreign_keys_enabled int
	err = database.QueryRow("pragma foreign_keys").Scan(&foreign_keys_enabled)
	slog.Info(fmt.Sprintf("foreign keys: %d", foreign_keys_enabled))

	slog.Info("db up")

	// catch exit signal to stop api & tasks
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg sync.WaitGroup

	// serve api
	address := env.GetString("JUMP_HTTP_ADDRESS")
	wg.Add(1)
	go func() {
		defer wg.Done()
		internal.ServeAPI(ctx, address)
	}()
	slog.Info("serving api", "address", address)

	// task handling
	tasks.InitWebhookUrls()
	client.ServeTaskClient()
	wg.Add(1)
	go func() {
		defer wg.Done()
		workers.ServeWorker(ctx)
	}()
	slog.Info("task worker up")

	// exit
	<-ctx.Done()
	wg.Wait()
	slog.Info("exit")
}
