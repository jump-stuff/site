package routes

import (
	"context"

	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/models"
)

func HandleGetStats(ctx context.Context, input *struct{}) (*models.SiteStatsOutput, error) {
	players, err := db.Queries.CountPlayers(ctx)
	times, err := db.Queries.CountTimes(ctx)
	events, err := db.Queries.CountEvents(ctx)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	resp := &models.SiteStatsOutput{
		Body: models.SiteStats{
			PlayerCount: players,
			TimesCount:  times,
			EventCount:  events,
		},
	}
	return resp, err
}
