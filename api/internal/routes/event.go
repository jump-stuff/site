package routes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/db/queries"
	"github.com/jump-fortress/site/internal/principal"
	"github.com/jump-fortress/site/internal/rows"
	"github.com/jump-fortress/site/models"
	"github.com/jump-fortress/site/tasks"
	"github.com/jump-fortress/site/tasks/client"
)

func motwNotEnded(kind string, endsAt time.Time) bool {
	return endsAt.After(time.Now().UTC())
}

func getEndsAt(kind string, starts_at time.Time) time.Time {
	switch kind {
	case "monthly":
		return starts_at.Add(time.Hour * 24 * 2)
	case "motw":
		// already calculated
		return starts_at
	default:
		// return, since ends_at was actually passed in
		return starts_at
	}
}

// todo: starts_at and ends_at tasks will not complete if changed after event is visible
func ScheduleUpdatedEventTasks(old models.Event, new models.Event) {
	if !old.VisibleAt.Equal(new.VisibleAt) {
		task, err := tasks.NewEventVisibleTask(new)
		if err != nil {
			return
		}
		client.QueueScheduledTask(task, fmt.Sprintf("%s%dvisible_updated", new.Kind, new.KindID), new.VisibleAt)
	}
}

func HandleGetEvent(ctx context.Context, input *models.EventKindAndIDInput) (*models.EventWithLeaderboardsOutput, error) {
	// get event and leaderboards
	els, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
		Kind:   input.Kind,
		KindID: input.KindID,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	now := time.Now().UTC()

	sensitive := els[0].Event.StartsAt.After(now)
	if els[0].Event.Kind == "motw" {
		sensitive = motwNotEnded(els[0].Event.Kind, els[0].Event.EndsAt)
	}
	if sensitive && els[0].Event.VisibleAt.After(now) {
		return nil, huma.Error400BadRequest("event not visible")
	}

	resp := &models.EventWithLeaderboardsOutput{
		Body: models.GetEventWithLeaderboardsResponse(els, sensitive),
	}

	return resp, nil

}

func HandleGetEventKinds(ctx context.Context, input *models.EventKindInput) (*models.EventsWithLeaderboardsOutput, error) {
	// validate input
	if !slices.Contains(models.EventKinds, input.Kind) {
		return nil, models.EventKindErr(input.Kind)
	}

	// get event kinds
	events, err := db.Queries.SelectEventKinds(ctx, input.Kind)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	now := time.Now().UTC()
	resp := &models.EventsWithLeaderboardsOutput{
		Body: []models.EventWithLeaderboards{},
	}

	for _, e := range events {
		// skip non-visible events
		if e.VisibleAt.After(now) {
			continue
		}

		// append leaderboards to each event
		ls, err := db.Queries.SelectLeaderboards(ctx, e.ID)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		els := []queries.SelectEventLeaderboardsRow{}
		for _, l := range ls {
			els = append(els, queries.SelectEventLeaderboardsRow{
				Event:       e,
				Leaderboard: l,
			})
		}

		sensitive := e.StartsAt.After(now)
		if els[0].Event.Kind == "motw" {
			sensitive = motwNotEnded(els[0].Event.Kind, els[0].Event.EndsAt)
		}
		if len(els) != 0 {
			resp.Body = append(resp.Body, models.GetEventWithLeaderboardsResponse(els, sensitive))
		}
	}

	if len(resp.Body) == 0 {
		return nil, huma.Error400BadRequest("no events visible")
	}
	return resp, nil
}

func HandleGetRecentEvents(ctx context.Context, _ *struct{}) (*models.EventsWithLeaderboardsOutput, error) {
	// prepare response
	resp := &models.EventsWithLeaderboardsOutput{
		Body: []models.EventWithLeaderboards{},
	}

	now := time.Now().UTC()

	// get monthly
	monthly, err := db.Queries.SelectLastEventKind(ctx, "monthly")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, models.WrapDBErr(err)
	}
	if err == nil {
		mewls, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
			Kind:   "monthly",
			KindID: monthly.KindID,
		})
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		sensitive := mewls[0].Event.StartsAt.After(now)
		resp.Body = append(resp.Body, models.GetEventWithLeaderboardsResponse(mewls, sensitive))
	}

	// get motw
	motw, err := db.Queries.SelectLastEventKind(ctx, "motw")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, models.WrapDBErr(err)
	}
	if err == nil {
		mewls, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
			Kind:   "motw",
			KindID: motw.KindID,
		})
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		sensitive := motwNotEnded(mewls[0].Event.Kind, mewls[0].Event.EndsAt)
		resp.Body = append(resp.Body, models.GetEventWithLeaderboardsResponse(mewls, sensitive))
	}

	return resp, nil
}

// session
func HandleGetMotw(ctx context.Context, input *models.EventKindAndIDInput) (*models.EventWithLeaderboardsOutput, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// get motw and leaderboards
	els, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
		Kind:   "motw",
		KindID: input.KindID,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// get player timeslot
	playerTimeslot, err := db.Queries.SelectPlayerTimeslot(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	eventPts := GetTimeslotDatetimes(playerTimeslot.MotwTimeslot, els[0].Event.StartsAt)

	now := time.Now().UTC()

	// if motw hasn't started for player's timeslot
	sensitive := eventPts.StartsAt.After(now)
	if sensitive && els[0].Event.VisibleAt.After(now) {
		return nil, huma.Error400BadRequest("event not visible")
	}

	resp := &models.EventWithLeaderboardsOutput{
		Body: models.GetEventWithLeaderboardsResponse(els, sensitive),
	}

	return resp, nil

}

// admin

func HandleGetFullEvents(ctx context.Context, _ *struct{}) (*models.EventsWithLeaderboardsOutput, error) {
	// get events
	events, err := db.Queries.SelectEvents(ctx)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.EventsWithLeaderboardsOutput{
		Body: []models.EventWithLeaderboards{},
	}

	for _, e := range events {
		// append leaderboards to each event
		ls, err := db.Queries.SelectLeaderboards(ctx, e.ID)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		els := []queries.SelectEventLeaderboardsRow{}
		for _, l := range ls {
			els = append(els, queries.SelectEventLeaderboardsRow{
				Event:       e,
				Leaderboard: l,
			})
		}

		// don't throw out events with no divisions
		if len(els) != 0 {
			resp.Body = append(resp.Body, models.GetEventWithLeaderboardsResponse(els, false))
		} else {
			resp.Body = append(resp.Body, models.GetEventWithLeaderboardsResponse([]queries.SelectEventLeaderboardsRow{{
				Event:       e,
				Leaderboard: queries.Leaderboard{},
			}}, false))
		}
	}

	if len(resp.Body) == 0 {
		return nil, huma.Error400BadRequest("no events found")
	}
	return resp, nil
}

func HandleCreateEvent(ctx context.Context, input *models.EventInput) (*struct{}, error) {
	// validate event info
	ie := input.Body
	if !slices.Contains(models.EventKinds, ie.Kind) {
		return nil, models.EventKindErr(ie.Kind)
	}
	if ie.PlayerClass != "Soldier" && ie.PlayerClass != "Demo" {
		return nil, models.PlayerClassErr(ie.PlayerClass)
	}

	now := time.Now().UTC()
	if ie.EndsAt.Before(ie.StartsAt) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't end before it starts (%s)", ie.EndsAt.String()))
	}
	if ie.StartsAt.Before(now) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't start in the past (%s)", ie.StartsAt.String()))
	}
	if ie.StartsAt.Before(ie.VisibleAt) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't start before it's visible (%s)", ie.VisibleAt.String()))
	}

	// set ID for next event
	kindID, err := db.Queries.CountEventKinds(ctx, ie.Kind)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	kindID++

	endsAt := ie.EndsAt
	// set motw start and end times based on current timeslots
	if ie.Kind == "motw" {
		firstTimeslot, err := db.Queries.SelectFirstTimeslot(ctx)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		lastTimeslot, err := db.Queries.SelectLastTimeslot(ctx)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		eventStartTs := GetTimeslotDatetimes(firstTimeslot, ie.StartsAt)
		ie.StartsAt = eventStartTs.StartsAt
		eventEndTs := GetTimeslotDatetimes(lastTimeslot, ie.EndsAt)
		endsAt = eventEndTs.EndsAt
	}
	endsAt = getEndsAt(ie.Kind, endsAt)

	// create event
	createdEvent, err := db.Queries.InsertEvent(ctx, queries.InsertEventParams{
		Kind:      ie.Kind,
		KindID:    kindID,
		Class:     ie.PlayerClass,
		VisibleAt: ie.VisibleAt,
		StartsAt:  ie.StartsAt,
		EndsAt:    endsAt,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// relay delayed visible_at announcement to #event-updates
	// starts_at task is queued after a successful visible_at announcement
	// ends_at task is queued after a successful starts_at announcement
	task, err := tasks.NewEventVisibleTask(models.GetEventResponse(createdEvent))
	client.QueueScheduledTask(task, fmt.Sprintf("%s%d%svisible", createdEvent.Kind, createdEvent.KindID, createdEvent.VisibleAt.String()), createdEvent.VisibleAt)

	return nil, nil
}

func HandleUpdateEvent(ctx context.Context, input *models.EventInput) (*struct{}, error) {
	// validate event info
	ie := input.Body
	event, err := db.Queries.SelectEvent(ctx, ie.ID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	if ie.Kind != event.Kind || ie.KindID != event.KindID {
		return nil, huma.Error400BadRequest("can't modify the event kind or kind_id")
	}

	now := time.Now().UTC()
	if event.StartsAt.Before(now) {
		return nil, huma.Error400BadRequest("event has already started")
	}
	if ie.EndsAt.Before(ie.StartsAt) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't end before it starts (%s)", ie.EndsAt.String()))
	}
	if ie.StartsAt.Before(now) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't start in the past (%s)", ie.StartsAt.String()))
	}
	if ie.StartsAt.Before(ie.VisibleAt) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("event can't start before it's visible (%s)", ie.VisibleAt.String()))
	}

	// if changing event kind, kind id needs to be re-calculated
	// todo: this should never happen because of the earlier check?
	kindID := event.KindID
	if ie.Kind != event.Kind {
		// set ID for next event
		kindID, err := db.Queries.CountEventKinds(ctx, ie.Kind)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
		kindID++
	}

	// update event
	ie.EndsAt = getEndsAt(ie.Kind, ie.EndsAt)
	err = db.Queries.UpdateEvent(ctx, queries.UpdateEventParams{
		Kind:      ie.Kind,
		KindID:    kindID,
		Class:     ie.PlayerClass,
		VisibleAt: ie.VisibleAt,
		StartsAt:  ie.StartsAt,
		EndsAt:    ie.EndsAt,
		ID:        ie.ID,
	})

	ScheduleUpdatedEventTasks(models.GetEventResponse(event), ie)

	return nil, nil
}

func HandleCancelEvent(ctx context.Context, input *models.EventIDInput) (*struct{}, error) {
	// check that the event hasn't started
	event, err := db.Queries.SelectEvent(ctx, input.ID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	now := time.Now().UTC()
	if event.StartsAt.Before(now) {
		return nil, huma.Error400BadRequest("event has already started")
	}

	err = db.Queries.DeleteEvent(ctx, event.ID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// backup event
	return nil, rows.InsertDeleted(ctx, event, "event", event.ID)
}

func HandleUpdateEventResults(ctx context.Context, input *models.EventIDInput) (*struct{}, error) {
	event, err := db.Queries.SelectEvent(ctx, input.ID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	ewls, err := db.Queries.SelectEventLeaderboards(ctx, queries.SelectEventLeaderboardsParams{
		Kind:   event.Kind,
		KindID: event.KindID,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	for _, ewl := range ewls {
		placements, err := db.Queries.SelectPrizepool(ctx, ewl.Leaderboard.ID)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}

		twps, err := db.Queries.SelectPRTimesFromLeaderboard(ctx, ewl.Leaderboard.ID)
		if err != nil {
			return nil, models.WrapDBErr(err)
		}

		for i, placement := range placements {
			if len(twps) > i {
				err = db.Queries.UpdatePrize(ctx, queries.UpdatePrizeParams{
					PlayerID: sql.NullString{
						String: twps[i].Player.ID,
						Valid:  true,
					},
					LeaderboardID: ewl.Leaderboard.ID,
					Position:      placement.Position,
				})
			}
		}
	}

	return nil, nil
}
