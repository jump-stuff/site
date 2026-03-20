package routes

import (
	"context"
	"fmt"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/db/queries"
	"github.com/jump-fortress/site/internal/principal"
	"github.com/jump-fortress/site/models"
	"github.com/jump-fortress/site/tasks"
	"github.com/jump-fortress/site/tasks/client"
)

func HandleSubmitRequest(ctx context.Context, input *models.RequestInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// check for existing pending request
	requestExists, err := db.Queries.CheckPendingRequestExists(ctx, queries.CheckPendingRequestExistsParams{
		PlayerID: principal.SteamID.String(),
		Kind:     input.Kind,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	if requestExists == 1 {
		return nil, huma.Error409Conflict(fmt.Sprintf("%s request already exists", input.Kind))
	}

	player, err := db.Queries.SelectPlayer(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// validate alias
	if input.Kind == "alias update" {
		if len(input.Content) > 32 {
			return nil, huma.Error400BadRequest("alias is too long (<32 characters)")
		}
		if !AliasRegex.MatchString(input.Content) {
			return nil, huma.Error400BadRequest("alias is invalid (alphanumeric only and in-between spaces, dots, underscores)")
		}
		// div request
	} else {
		if !player.TempusID.Valid {
			return nil, huma.Error400BadRequest("missing a Tempus ID")
		}
		if !slices.Contains(models.Divs, input.Content) && input.Content != "none" {
			return nil, huma.Error400BadRequest(fmt.Sprintf("%s isn't a div", input.Content))
		}
	}

	request, err := db.Queries.InsertRequest(ctx, queries.InsertRequestParams{
		PlayerID: principal.SteamID.String(),
		Kind:     input.Kind,
		Content:  input.Content,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// relay to #div-talk
	task, err := tasks.NewPlayerRequestTask(models.RequestWithPlayer{
		Request: models.GetRequestResponse(request),
		Player:  models.GetPlayerResponse(player, false),
	})
	client.QueueTask(task, fmt.Sprintf("%s%s", player.ID, request.Kind))

	return nil, nil
}

func HandleGetRequestsSelf(ctx context.Context, _ *struct{}) (*models.RequestsWithPlayerOutput, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	requests, err := db.Queries.SelectPlayerRequests(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.RequestsWithPlayerOutput{
		Body: []models.RequestWithPlayer{},
	}
	for _, rwp := range requests {
		resp.Body = append(resp.Body, models.RequestWithPlayer{
			Request: models.GetRequestResponse(rwp.Request),
			Player:  models.GetPlayerResponse(rwp.Player, false),
		})
	}

	return resp, nil
}

// consultant

func HandleGetRequests(ctx context.Context, _ *struct{}) (*models.RequestsWithPlayerOutput, error) {
	requests, err := db.Queries.SelectPendingRequests(ctx)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.RequestsWithPlayerOutput{
		Body: []models.RequestWithPlayer{},
	}
	for _, rwp := range requests {
		resp.Body = append(resp.Body, models.RequestWithPlayer{
			Request: models.GetRequestResponse(rwp.Request),
			Player:  models.GetPlayerResponse(rwp.Player, false),
		})
	}

	return resp, nil

}

// mod

func HandleResolveRequest(ctx context.Context, input *models.RequestIDInput) (*struct{}, error) {
	err := db.Queries.ResolveRequest(ctx, input.ID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	return nil, nil
}
