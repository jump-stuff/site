package routes

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jump-fortress/site/db"
	"github.com/jump-fortress/site/db/queries"
	"github.com/jump-fortress/site/internal/principal"
	"github.com/jump-fortress/site/internal/tempus"
	"github.com/jump-fortress/site/models"
	"github.com/jump-fortress/site/tasks"
	"github.com/jump-fortress/site/tasks/client"
)

var (
	// alphanumeric, allow non-leading, non-repeating underscores, dashes, whitespace, and periods
	AliasRegex = regexp.MustCompile(`^((([a-z]|[A-Z]|\d|\.)+(_|\ |\-)?)+)*([a-z]|[A-Z]|\d|\.)+$`)
)

func GetPlayersResponse(ctx context.Context, sensitive bool) (*models.PlayersOutput, error) {
	players, err := db.Queries.SelectPlayers(ctx)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.PlayersOutput{
		Body: []models.Player{},
	}
	for _, p := range players {
		resp.Body = append(resp.Body, models.GetPlayerResponse(p, sensitive))
	}
	return resp, nil
}

func GetMapNames(ctx context.Context) ([]string, error) {
	maps, err := db.Queries.SelectMaps(ctx)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	mapNames := []string{}
	for _, m := range maps {
		mapNames = append(mapNames, m.Name)
	}
	return mapNames, nil
}

func HandleGetPlayer(ctx context.Context, input *models.PlayerIDInput) (*models.PlayerOutput, error) {
	player, err := db.Queries.SelectPlayer(ctx, input.PlayerID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.PlayerOutput{
		Body: models.GetPlayerResponse(player, true),
	}
	return resp, nil
}

// include full info for self
func HandleGetPlayerSelf(ctx context.Context, input *struct{}) (*models.PlayerOutput, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	player, err := db.Queries.SelectPlayer(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.PlayerOutput{
		Body: models.GetPlayerResponse(player, false),
	}
	return resp, nil
}

func HandleGetPlayers(ctx context.Context, input *struct{}) (*models.PlayersOutput, error) {
	resp, err := GetPlayersResponse(ctx, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func HandleSetTempusID(ctx context.Context, input *models.TempusIDInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// check if already set
	player, err := db.Queries.SelectPlayer(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	if player.TempusID.Valid {
		return nil, huma.Error400BadRequest(fmt.Sprintf("Tempus ID is already set to %d", player.TempusID.Int64))
	}

	// get Tempus info
	tp, err := tempus.GetPlayerInfo(input.TempusID)
	if err != nil {
		return nil, models.WrapTempusErr(err)
	}

	// parse Tempus's SteamID3 to matching SteamID
	var tpSteamID3 uint32
	_, err = fmt.Sscanf(tp.SteamID, "STEAM_0:1:%d", &tpSteamID3)
	tpSteamID3 = tpSteamID3*2 + 1

	if principal.SteamID.AccountId() != tpSteamID3 {
		return nil, huma.Error400BadRequest(fmt.Sprintf("the Tempus player found (last seen as %s) doesn't match your SteamID", tp.Name))
	}

	// update Tempus info
	err = db.Queries.UpdatePlayerTempusInfo(ctx, queries.UpdatePlayerTempusInfoParams{
		TempusID: sql.NullInt64{
			Int64: tp.ID,
			Valid: true,
		},
		Country: sql.NullString{
			String: tp.Country,
			Valid:  true,
		},
		CountryCode: sql.NullString{
			String: strings.ToLower(tp.CountryCode),
			Valid:  true,
		},
		ID: principal.SteamID.String(),
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// relay to #div-talk
	player.TempusID.Int64 = input.TempusID
	task, err := tasks.NewPlayerSetTempusIDTask(player)
	fmt.Println("making task")
	client.QueueTask(task, fmt.Sprintf("%s%d", player.ID, input.TempusID))
	fmt.Println("and made i hope")
	return nil, nil
}

func HandleSetTradeToken(ctx context.Context, input *models.TradeTokenInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	if len(input.SteamTradeURL) > 100 {
		return nil, huma.Error400BadRequest("URL is too long. please paste the full Steam Trade URL.")
	}

	// extract steam3ID and token from input
	var scannedSteam3ID uint32
	var token string
	_, err := fmt.Sscanf(input.SteamTradeURL, "https://steamcommunity.com/tradeoffer/new/?partner=%d&token=%s", &scannedSteam3ID, &token)
	if err != nil {
		return nil, huma.Error400BadRequest("URL couldn't be resolved. please paste the full Steam Trade URL.")
	}

	if principal.SteamID.AccountId() != scannedSteam3ID {
		return nil, huma.Error400BadRequest(fmt.Sprintf("URL didn't match your profile's SteamID (%d). please check if you're logged in to the correct Steam account.", scannedSteam3ID))
	}

	if err := db.Queries.UpdatePlayerTradeToken(ctx, queries.UpdatePlayerTradeTokenParams{
		TradeToken: sql.NullString{
			String: token,
			Valid:  true,
		},
		ID: principal.SteamID.String(),
	}); err != nil {
		return nil, models.WrapDBErr(err)
	}

	return nil, nil
}

func HandleUpdateClassPref(ctx context.Context, input *models.PlayerClassInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	err := db.Queries.UpdatePlayerClassPref(ctx, queries.UpdatePlayerClassPrefParams{
		ClassPref: input.PlayerClass,
		ID:        principal.SteamID.String(),
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	return nil, nil
}

func HandleUpdateMapPref(ctx context.Context, input *models.MapNameInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	nullInput := input.MapName == "none"
	maps, err := GetMapNames(ctx)
	if err != nil {
		return nil, err
	}
	if !nullInput && !slices.Contains(maps, input.MapName) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("%s isn't a map name", input.MapName))
	}

	err = db.Queries.UpdatePlayerMapPref(ctx, queries.UpdatePlayerMapPrefParams{
		MapPref: sql.NullString{
			String: input.MapName,
			Valid:  !nullInput,
		},
		ID: principal.SteamID.String(),
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	return nil, nil
}

func HandleUpdateLauncherPref(ctx context.Context, input *models.LauncherInput) (*struct{}, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	nullInput := input.Launcher == "none"
	if !nullInput && !slices.Contains([]string{"stock", "original", "mangler"}, input.Launcher) {
		return nil, huma.Error400BadRequest(fmt.Sprintf("%s isn't a launcher name", input.Launcher))
	}

	err := db.Queries.UpdatePlayerLauncherPref(ctx, queries.UpdatePlayerLauncherPrefParams{
		LauncherPref: sql.NullString{
			String: input.Launcher,
			Valid:  !nullInput,
		},
		ID: principal.SteamID.String(),
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	return nil, nil
}

// consultant

func HandleGetFullPlayer(ctx context.Context, input *models.PlayerIDInput) (*models.PlayerOutput, error) {
	player, err := db.Queries.SelectPlayer(ctx, input.PlayerID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	resp := &models.PlayerOutput{
		Body: models.GetPlayerResponse(player, false),
	}
	return resp, nil
}

func HandleGetFullPlayers(ctx context.Context, input *struct{}) (*models.PlayersOutput, error) {
	resp, err := GetPlayersResponse(ctx, false)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func HandleGetAuditLogs(ctx context.Context, input *struct{}) (*models.AuditLogOutput, error) {
	logs, err := db.Queries.SelectAuditLogs(ctx)
	if err != nil {
		return nil, err
	}

	resp := &models.AuditLogOutput{
		Body: []models.AuditLog{},
	}
	for _, l := range logs {
		resp.Body = append(resp.Body, models.GetAuditLogResponse(l.AuditLog, l.Player, l.Player_2))
	}
	return resp, nil
}

// mod

func HandleUpdatePlayerDiv(ctx context.Context, input *models.UpdatePlayerDivInput) (*struct{}, error) {
	// need SteamID for audit log
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// validate player exists
	player, err := db.Queries.SelectPlayer(ctx, input.PlayerID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	fromDiv := "none"

	// validate inputs
	if input.PlayerClass != "Soldier" && input.PlayerClass != "Demo" {
		return nil, models.PlayerClassErr(input.PlayerClass)
	}

	nullInput := input.Div == "none"
	if !slices.Contains(models.Divs, input.Div) && !nullInput {
		return nil, huma.Error400BadRequest(fmt.Sprintf("%s isn't a div", input.Div))
	}

	// update div
	if input.PlayerClass == "Soldier" {
		if player.SoldierDiv.Valid {
			fromDiv = player.SoldierDiv.String
		}

		err := db.Queries.UpdatePlayerSoldierDiv(ctx, queries.UpdatePlayerSoldierDivParams{
			SoldierDiv: sql.NullString{
				String: input.Div,
				Valid:  !nullInput,
			},
			ID: input.PlayerID,
		})
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
	} else {
		if player.DemoDiv.Valid {
			fromDiv = player.DemoDiv.String
		}

		err := db.Queries.UpdatePlayerDemoDiv(ctx, queries.UpdatePlayerDemoDivParams{
			DemoDiv: sql.NullString{
				String: input.Div,
				Valid:  !nullInput,
			},
			ID: input.PlayerID,
		})
		if err != nil {
			return nil, models.WrapDBErr(err)
		}
	}

	// create audit log
	err = db.Queries.InsertAuditLog(ctx, queries.InsertAuditLogParams{
		FromPlayerID: principal.SteamID.String(),
		ToPlayerID:   player.ID,
		Kind:         fmt.Sprintf("updated %s div", input.PlayerClass),
		FromContent:  fromDiv,
		ToContent:    input.Div,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	return nil, nil
}

func HandleUpdatePlayerAlias(ctx context.Context, input *models.UpdatePlayerAliasInput) (*struct{}, error) {
	// need SteamID for audit log
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// validate alias
	if len(input.Alias) > 32 {
		return nil, huma.Error400BadRequest("alias is too long (<32 characters)")
	}
	if !AliasRegex.MatchString(input.Alias) {
		return nil, huma.Error400BadRequest("alias is invalid (alphanumeric only and in-between spaces, dots, underscores)")
	}

	// validate player exists
	player, err := db.Queries.SelectPlayer(ctx, input.PlayerID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// update alias
	err = db.Queries.UpdatePlayerAlias(ctx, queries.UpdatePlayerAliasParams{
		Alias: sql.NullString{
			String: input.Alias,
			Valid:  true,
		},
		ID: input.PlayerID,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// create audit log
	err = db.Queries.InsertAuditLog(ctx, queries.InsertAuditLogParams{
		FromPlayerID: principal.SteamID.String(),
		ToPlayerID:   player.ID,
		Kind:         "updated alias",
		FromContent:  player.Alias.String,
		ToContent:    input.Alias,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	return nil, nil
}

// admin

func HandleUpdatePlayerRole(ctx context.Context, input *models.UpdatePlayerRoleInput) (*struct{}, error) {
	// need SteamID for audit log
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	// update role
	err := db.Queries.UpdatePlayerRole(ctx, queries.UpdatePlayerRoleParams{
		Role: input.Role,
		ID:   input.PlayerID,
	})
	if err != nil {
		return nil, models.WrapDBErr(err)
	}

	// create audit log
	err = db.Queries.InsertAuditLog(ctx, queries.InsertAuditLogParams{
		FromPlayerID: principal.SteamID.String(),
		ToPlayerID:   input.PlayerID,
		Kind:         "updated role",
		FromContent:  "", // no player info
		ToContent:    input.Role,
	})

	return nil, nil
}
