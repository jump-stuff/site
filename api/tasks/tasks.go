package tasks

import (
	"context"
	"encoding/json"

	"github.com/gtuk/discordwebhook"
	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/env"
)

// task types
const (
	TypeWebhookReady = "test:ready"
)

// webhook URLs
var (
	testWebhookUrl string
)

// webhook ready payload
type webhookMessagePayload struct {
	Content string
}

func InitWebhookUrls() {
	testWebhookUrl = env.GetString("JUMP_WEBHOOK_TEST_URL")
}

func NewWebhookReadyTask(content string) (*asynq.Task, error) {
	payload, err := json.Marshal(webhookMessagePayload{
		Content: content,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeWebhookReady, payload), nil
}

func HandleWebhookReadyTask(ctx context.Context, t *asynq.Task) error {
	var p webhookMessagePayload
	err := json.Unmarshal(t.Payload(), &p)
	if err != nil {
		return err
	}

	// send message
	err = discordwebhook.SendMessage(testWebhookUrl, discordwebhook.Message{
		Content: &p.Content,
	})
	if err != nil {
		return err
	}
	return nil
}
