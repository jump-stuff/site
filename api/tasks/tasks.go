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
	"github.com/jump-fortress/site/models"
	"github.com/jump-fortress/site/tasks/client"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// todo: task to set event prizes / points on end, and on submissions after end

// task types
const (
	TypeEventVisible = "event:visible"
	TypeEventStarted = "event:started"
	TypeEventEnded   = "event:ended"
	TypeSetTempusID  = "player:tempusid"
	TypeNewRequest   = "player:request"
)

// webhook URLs
var (
	testWebhookUrl    string
	eventsWebhookUrl  string
	playersWebhookUrl string
	siteRealm         string

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

type visibleEventMessagePayload struct {
	Event models.Event
}

type eventMessagePayload struct {
	Ewl models.EventWithLeaderboards
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
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

func InitWebhookUrls() {
	siteRealm = env.GetString("JUMP_OID_REALM")
	testWebhookUrl = env.GetString("JUMP_WEBHOOK_TEST_URL")
	eventsWebhookUrl = env.GetString("JUMP_WEBHOOK_EVENTS_URL")
	playersWebhookUrl = env.GetString("JUMP_WEBHOOK_PLAYERS_URL")
}

func HandleWebhookReadyTask(ctx context.Context, t *asynq.Task) error {
	var p client.ReadyMessagePayload
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

func NewEventVisibleTask(event models.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(visibleEventMessagePayload{
		Event: event,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeEventVisible, payload), nil
}

func HandleEventVisibleTask(ctx context.Context, t *asynq.Task) error {
	var ep visibleEventMessagePayload
	err := json.Unmarshal(t.Payload(), &ep)
	if err != nil {
		return err
	}

	// check visible_at datetime hasn't been changed
	verifyEvent, err := db.Queries.SelectEvent(ctx, ep.Event.ID)
	if err != nil {
		return err
	}
	if !verifyEvent.VisibleAt.Equal(ep.Event.VisibleAt) {
		fmt.Println("event visible_at date changed, not announcing event")
		return err
	}

	// relay delayed starts_at announcement to #event-updates
	// event is missing leaderboards for starts_at task, so get them
	// if not a motw (multiple timezones), show maps
	isMotw := ep.Event.Kind == "motw"
	ewl, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
		Kind:   ep.Event.Kind,
		KindID: ep.Event.KindID,
	})
	if err != nil {
		return err
	}
	taskEwl := models.GetEventWithLeaderboardsResponse(ewl, isMotw)
	task, err := NewEventStartedTask(taskEwl)
	if err != nil {
		return err
	}
	client.QueueScheduledTask(task, fmt.Sprintf("%s%dstart", ep.Event.Kind, ep.Event.KindID), ep.Event.StartsAt)

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Event.PlayerClass, cases.Title(language.English).String(ep.Event.Kind))
	content := fmt.Sprintf("%sA new %s has appeared!", roleMention(eventName), ep.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Event.KindID)
	eDesc := fmt.Sprintf("starts %s", relativeTimestamp(ep.Event.StartsAt))
	eUrl := fmt.Sprintf("%s/formats/%s/%d", siteRealm, ep.Event.Kind, ep.Event.KindID)
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
		Ewl: ewl,
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

	// check starts_at datetime hasn't been changed
	verifyEvent, err := db.Queries.SelectEvent(ctx, ep.Ewl.Event.ID)
	if err != nil {
		return err
	}
	if !verifyEvent.StartsAt.Equal(ep.Ewl.Event.StartsAt) {
		fmt.Println("event starts_at date changed, not announcing event")
		return err
	}

	// relay delayed ends_at announcement to #event-updates
	// event is still "sensitive" data (missing maps) if a motw
	// if a motw, get maps
	if ep.Ewl.Event.Kind == "motw" {
		fullEwl, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
			Kind:   ep.Ewl.Event.Kind,
			KindID: ep.Ewl.Event.KindID,
		})
		if err != nil {
			return err
		}
		ep.Ewl = models.GetEventWithLeaderboardsResponse(fullEwl, false)
	}
	task, err := NewEventEndedTask(ep.Ewl)
	if err != nil {
		return err
	}
	client.QueueScheduledTask(task, fmt.Sprintf("%s%dend", ep.Ewl.Event.Kind, ep.Ewl.Event.KindID), ep.Ewl.Event.EndsAt)

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Ewl.Event.PlayerClass, cases.Title(language.English).String(ep.Ewl.Event.Kind))
	content := fmt.Sprintf("%sA %s has started!", roleMention(eventName), ep.Ewl.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Ewl.Event.KindID)
	eDesc := fmt.Sprintf("ends %s", relativeTimestamp(ep.Ewl.Event.EndsAt))
	eUrl := fmt.Sprintf("%s/formats/%s/%d", siteRealm, ep.Ewl.Event.Kind, ep.Ewl.Event.KindID)
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
		Ewl: ewl,
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

	// check ends_at datetime hasn't changed
	verifyEvent, err := db.Queries.SelectEvent(ctx, ep.Ewl.Event.ID)
	if err != nil {
		return err
	}
	if !verifyEvent.EndsAt.Equal(ep.Ewl.Event.EndsAt) {
		fmt.Println("event ends_at date changed, not announcing event")
		return err
	}

	// fill message fields
	eventName := fmt.Sprintf("%s %s", ep.Ewl.Event.PlayerClass, cases.Title(language.English).String(ep.Ewl.Event.Kind))
	// winning times per div
	var winningTimes strings.Builder
	for _, l := range ep.Ewl.Leaderboards {
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
	content := fmt.Sprintf("A %s has ended! Valid times can still be submitted if they took place during the %s.", ep.Ewl.Event.Kind, ep.Ewl.Event.Kind)

	mentions := slices.Collect(maps.Values(MentionRoleIDs))
	embeds := []discordwebhook.Embed{}

	// embed fields
	eTitle := fmt.Sprintf("%s #%d", eventName, ep.Ewl.Event.KindID)
	eDesc := fmt.Sprintf("ended %s\n\n the winning times are..\n%s", relativeTimestamp(ep.Ewl.Event.EndsAt), &winningTimes)
	eUrl := fmt.Sprintf("%s/formats/%s/%d", siteRealm, ep.Ewl.Event.Kind, ep.Ewl.Event.KindID)
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
	fmt.Println("yes we are makingg")
	payload, err := json.Marshal(playerPayload{
		Player: player,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSetTempusID, payload), nil
}

// players

func HandleNewPlayerSetTempusIDTask(ctx context.Context, t *asynq.Task) error {
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
	eAUrl := fmt.Sprintf("%s/players/%s", siteRealm, p.Player.ID)
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
	return asynq.NewTask(TypeNewRequest, payload), nil
}

func HandleNewPlayerRequestTask(ctx context.Context, t *asynq.Task) error {
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
	eAUrl := fmt.Sprintf("%s/players/%s", siteRealm, player.ID)
	eAIconUrl := player.AvatarURL
	eDesc := fmt.Sprintf("Type - %s\nContent - %s\nCreated %s", request.Kind, request.Content, relativeTimestamp(request.CreatedAt))
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
