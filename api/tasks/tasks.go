package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/gtuk/discordwebhook"
	"github.com/hibiken/asynq"
	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/db/queries"
	"github.com/jump-fortress/site/env"
	"github.com/jump-fortress/site/internal"
	"github.com/jump-fortress/site/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// todo: task to set event prizes / points on end, and on submissions after end

// task types
const (
	TypeWebhookReady = "test:ready"
	TypeEventVisible = "event:visible"
	TypeEventStarted = "event:started"
	TypeEventEnded   = "event:ended"
)

// webhook URLs
var (
	testWebhookUrl    string
	eventsWebhookUrl  string
	playersWebhookUrl string

	ColorDecimalContent = "13489908"
	ColorDecimalPrimary = "11845374"
	ColorTempus         = "16744192" // default orange

	DefaultImage = "https://files.catbox.moe/s1ounn.png"

	MentionRoleIDs = map[string]string{
		"Soldier Monthly": "1479824276948254953",
		"Demo Monthly":    "1479824444204650601",
		"Soldier Motw":    "1479824528959082749",
		"Demo Motw":       "1479824572399489166"}
)

// webhook ready payload
type readyMessagePayload struct {
	Content string
}

type eventMessagePayload struct {
	Event models.EventWithLeaderboards
}

type playerPayload struct {
	Player queries.Player
}

type requestPayload struct {
	Payload models.RequestWithPlayer
}

func roleMention(eventName string) string {
	roleID := MentionRoleIDs[eventName]
	if len(roleID) != 0 {
		return fmt.Sprintf("<@&%s>\n", roleID)
	}
	return ""
}

func relativeTimestamp(t time.Time) string {
	return fmt.Sprintf("<t:%d:R", t.UnixMilli())
}

func InitWebhookUrls() {
	testWebhookUrl = env.GetString("JUMP_WEBHOOK_TEST_URL")
	eventsWebhookUrl = env.GetString("JUMP_WEBHOOK_EVENTS_URL")
	playersWebhookUrl = env.GetString("JUMP_WEBHOOK_PLAYERS_URL")
}

func NewWebhookReadyTask(content string) (*asynq.Task, error) {
	payload, err := json.Marshal(readyMessagePayload{
		Content: content,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWebhookReady, payload), nil
}

func HandleWebhookReadyTask(ctx context.Context, t *asynq.Task) error {
	var p readyMessagePayload
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

// events

func NewEventVisibleTask(ewl models.EventWithLeaderboards) (*asynq.Task, error) {
	payload, err := json.Marshal(eventMessagePayload{
		Event: ewl,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventVisible, payload), nil
}

func HandleEventVisibleTask(ctx context.Context, t *asynq.Task) error {
	var ep eventMessagePayload
	err := json.Unmarshal(t.Payload(), &ep)
	if err != nil {
		return err
	}

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Event.Event.PlayerClass, cases.Title(language.English).String(ep.Event.Event.Kind))
	content := fmt.Sprintf("%sA new %s has appeared!", roleMention(eventName), ep.Event.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Event.Event.KindID)
	eDesc := fmt.Sprintf("starts %s", relativeTimestamp(ep.Event.Event.StartsAt))
	eUrl := fmt.Sprintf("%s/formats/%s/%d", internal.OidRealm, ep.Event.Event.Kind, ep.Event.Event.KindID)
	embeds = append(embeds, discordwebhook.Embed{
		Title:       &eTitle,
		Url:         &eUrl,
		Description: &eDesc,
		Color:       &ColorDecimalContent,
	})

	// send message
	err = discordwebhook.SendMessage(eventsWebhookUrl, discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
		AllowedMentions: &discordwebhook.AllowedMentions{
			Roles: &mentions,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewEventStartedTask(ewl models.EventWithLeaderboards) (*asynq.Task, error) {
	payload, err := json.Marshal(eventMessagePayload{
		Event: ewl,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventStarted, payload), nil
}

func HandleEventStartedTask(ctx context.Context, t *asynq.Task) error {
	var ep eventMessagePayload
	err := json.Unmarshal(t.Payload(), &ep)
	if err != nil {
		return err
	}

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Event.Event.PlayerClass, cases.Title(language.English).String(ep.Event.Event.Kind))
	content := fmt.Sprintf("%sA %s has started!", roleMention(eventName), ep.Event.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Event.Event.KindID)
	eDesc := fmt.Sprintf("ends %s", relativeTimestamp(ep.Event.Event.EndsAt))
	eUrl := fmt.Sprintf("%s/formats/%s/%d", internal.OidRealm, ep.Event.Event.Kind, ep.Event.Event.KindID)
	embeds = append(embeds, discordwebhook.Embed{
		Title:       &eTitle,
		Url:         &eUrl,
		Description: &eDesc,
		Color:       &ColorDecimalPrimary,
		Image:       &discordwebhook.Image{Url: &DefaultImage},
	})

	// send message
	err = discordwebhook.SendMessage(eventsWebhookUrl, discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
		AllowedMentions: &discordwebhook.AllowedMentions{
			Roles: &mentions,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewEventEndedTask(ewl models.EventWithLeaderboards) (*asynq.Task, error) {
	payload, err := json.Marshal(eventMessagePayload{
		Event: ewl,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventEnded, payload), nil
}

func HandleEventEndedTask(ctx context.Context, t *asynq.Task) error {
	var ep eventMessagePayload
	err := json.Unmarshal(t.Payload(), &ep)
	if err != nil {
		return err
	}

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Event.Event.PlayerClass, cases.Title(language.English).String(ep.Event.Event.Kind))
	// winning times per div
	var winningTimes strings.Builder
	for _, l := range ep.Event.Leaderboards {
		// get duration
		twps, err := db.Queries.SelectPRTimesFromLeaderboard(ctx, l.ID)
		if err != nil {
			return err
		}
		if len(twps) != 0 {
			duration := twps[0].Time.Duration
			minutes := int64(math.Floor(duration / 60))
			seconds := math.Mod(duration, 60)
			fmt.Fprintf(&winningTimes, "%s (%s) - %d:%.3f by %s\n", l.Div, l.Map, minutes, seconds, twps[0].Player.Alias.String)
		}
	}
	content := fmt.Sprintf("A %s has ended! Valid times can still be submitted if they took place during the %s.", ep.Event.Event.Kind, ep.Event.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Event.Event.KindID)
	eDesc := fmt.Sprintf("ended %s\n\n the winning times are..\n%s", relativeTimestamp(ep.Event.Event.EndsAt), &winningTimes)
	eUrl := fmt.Sprintf("%s/formats/%s/%d", internal.OidRealm, ep.Event.Event.Kind, ep.Event.Event.KindID)
	embeds = append(embeds, discordwebhook.Embed{
		Title:       &eTitle,
		Url:         &eUrl,
		Description: &eDesc,
		Color:       &ColorDecimalPrimary,
		Image:       &discordwebhook.Image{Url: &DefaultImage},
	})

	// send message
	err = discordwebhook.SendMessage(eventsWebhookUrl, discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
		AllowedMentions: &discordwebhook.AllowedMentions{
			Roles: &mentions,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewPlayerSetTempusIDTask(player queries.Player) (*asynq.Task, error) {
	payload, err := json.Marshal(playerPayload{
		Player: player,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventEnded, payload), nil
}

// players

func HandleNewPlayerSetTempusIDTask(ctx context.Context, t asynq.Task) error {
	var p playerPayload
	err := json.Unmarshal(t.Payload(), &p)
	if err != nil {
		return err
	}

	// fill message fields
	content := "A player set their Tempus ID"
	embeds := []discordwebhook.Embed{}

	// embed fields
	eAuthor := p.Player.Alias.String
	eAUrl := fmt.Sprintf("%s/players/%s", internal.OidRealm, p.Player.ID)
	eAIconUrl := p.Player.AvatarUrl.String
	eDesc := fmt.Sprintf("Soldier Div - %s\nDemo Div - %s\n Joined %s", p.Player.SoldierDiv.String, p.Player.DemoDiv.String, relativeTimestamp(p.Player.CreatedAt))
	embeds = append(embeds, discordwebhook.Embed{
		Author: &discordwebhook.Author{
			Name:    &eAuthor,
			Url:     &eAUrl,
			IconUrl: &eAIconUrl,
		},
		Description: &eDesc,
		Color:       &ColorTempus,
	})

	// send message
	err = discordwebhook.SendMessage(playersWebhookUrl, discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewPlayerRequestTask(rwp models.RequestWithPlayer) (*asynq.Task, error) {
	payload, err := json.Marshal(requestPayload{
		Payload: rwp,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventEnded, payload), nil
}

func HandleNewPlayerRequestTask(ctx context.Context, t asynq.Task) error {
	var rp requestPayload
	err := json.Unmarshal(t.Payload(), &rp)
	if err != nil {
		return err
	}
	request := rp.Payload.Request
	player := rp.Payload.Player

	// fill message fields
	content := "A player created a request"
	embeds := []discordwebhook.Embed{}

	// embed fields
	eAuthor := player.Alias
	eAUrl := fmt.Sprintf("%s/players/%s", internal.OidRealm, player.ID)
	eAIconUrl := player.AvatarURL
	eDesc := fmt.Sprintf("Type - %s\nContent - %s\nCreated %s", request.Kind, request.Content, request.CreatedAt)
	embeds = append(embeds, discordwebhook.Embed{
		Author: &discordwebhook.Author{
			Name:    &eAuthor,
			Url:     &eAUrl,
			IconUrl: &eAIconUrl,
		},
		Description: &eDesc,
		Color:       &ColorDecimalContent,
	})

	// send message
	err = discordwebhook.SendMessage(playersWebhookUrl, discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
	})
	if err != nil {
		return err
	}
	return nil
}
