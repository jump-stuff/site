package workers

import (
	"context"
	"log"
	"log/slog"

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
			Password: redisPass},
		asynq.Config{Concurrency: 10},
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
