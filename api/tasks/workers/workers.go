package workers

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/tasks"
)

func ServeWorker(ctx context.Context) {
	redisAddress := env.GetString("JUMP_UPSTASH_REDIS_REST_URL")
	redisPass := env.GetString("JUMP_UPSTASH_REDIS_REST_PASS")
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddress,
			Password: redisPass,
			DB:       0},
		asynq.Config{Concurrency: 0, TaskCheckInterval: time.Minute, HealthCheckInterval: time.Minute * 15, DelayedTaskCheckInterval: time.Minute, JanitorInterval: time.Hour, JanitorBatchSize: 5},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeWebhookReady, tasks.HandleWebhookReadyTask)

	go func() {
		err := srv.Run(mux)
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	srv.Stop()
	srv.Shutdown()
	slog.Info("worker shutdown")
}
