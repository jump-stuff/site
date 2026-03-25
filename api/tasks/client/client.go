package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/env"
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

func QueueScheduledTask(t *asynq.Task, tID string, datetime time.Time) {
	info, err := taskClient.Enqueue(t, asynq.TaskID(tID), asynq.ProcessAt(datetime))
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't enqueue %s task", tID), "error", err)
	}
	slog.Info(fmt.Sprintf("enqueued %s task", tID), "info", info)
}

// webhook ready payload
type ReadyMessagePayload struct {
	Content string
}

func NewWebhookReadyTask(content string) (*asynq.Task, error) {
	payload, err := json.Marshal(ReadyMessagePayload{
		Content: content,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("test:ready", payload), nil
}

func ServeTaskClient() {
	redisAddress := env.GetString("JUMP_UPSTASH_REDIS_REST_URL")
	redisPass := env.GetString("JUMP_UPSTASH_REDIS_REST_PASS")
	taskClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddress,
		Password: redisPass,
		DB:       0,
	})

	readyTask, err := NewWebhookReadyTask("task handling ready")
	if err != nil {
		slog.Error("couldn't create ready task", "error", err)
	}

	QueueTask(readyTask, "ready")
}
