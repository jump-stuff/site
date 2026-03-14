package client

import (
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/tasks"
)

func ServeTaskClient() {
	redisAddress := env.GetString("JUMP_UPSTASH_REDIS_REST_URL")
	redisPass := env.GetString("JUMP_UPSTASH_REDIS_REST_PASS")
	taskClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddress,
		Password: redisPass,
		DB:       0,
	})

	readyTask, err := tasks.NewWebhookReadyTask("task handling ready")
	if err != nil {
		slog.Error("couldn't create ready task", "error", err)
	}

	info, err := taskClient.Enqueue(readyTask, asynq.TaskID("ready"))
	if err != nil {
		slog.Error("couldn't enqueue ready task", "error", err)
	}
	slog.Info("enqueued ready task", "info", info)
}
