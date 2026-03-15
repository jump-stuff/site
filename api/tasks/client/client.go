package client

import (
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/tasks"
)

var (
	taskClient *asynq.Client
)

func QueueTask(t *asynq.Task, tID string) {
	info, err := taskClient.Enqueue(t, asynq.TaskID(tID))
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't enqueue %s task", tID), "error", err)
	}
	slog.Info(fmt.Sprintf("enqueued %s task", tID), "info", info)
}

func ServeTaskClient() {
	redisAddress := env.GetString("JUMP_UPSTASH_REDIS_REST_URL")
	redisPass := env.GetString("JUMP_UPSTASH_REDIS_REST_PASS")
	taskClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddress,
		Password: redisPass,
		DB:       0,
	})

	readyTask, err := tasks.NewWebhookReadyTask("task handling ready")
	if err != nil {
		slog.Error("couldn't create ready task", "error", err)
	}

	QueueTask(readyTask, "ready")
}
