package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterOpenRoutes(internalApi *huma.Group) {
	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/{player_id}",
		Tags:        []string{"players"},
		Summary:     "get player",
		Description: "get a player by player_id",
		OperationID: "get-player",
	}, HandleGetPlayer)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players",
		Tags:        []string{"players"},
		Summary:     "get all players",
		Description: "get all players",
		OperationID: "get-players",
	}, HandleGetPlayers)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/{event_kind}/{kind_id}",
		Tags:        []string{"events"},
		Summary:     "get event",
		Description: "get an event by its kind and kind_id",
		OperationID: "get-event",
	}, HandleGetEvent)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/{event_kind}",
		Tags:        []string{"events"},
		Summary:     "get all of event kind",
		Description: "get all of a kind of event",
		OperationID: "get-event-kinds",
	}, HandleGetEventKinds)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/recent",
		Tags:        []string{"events"},
		Summary:     "get recent events",
		Description: "get most recent reoccuring events",
		OperationID: "get-recent-events",
	}, HandleGetRecentEvents)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/maps",
		Tags:        []string{"maps"},
		Summary:     "get maps",
		Description: "get all maps",
		OperationID: "get-maps",
	}, HandleGetMaps)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/leaderboards/{leaderboard_id}/times",
		Tags:        []string{"times"},
		Summary:     "get leaderboard times",
		Description: "get all times for an event's leaderboard",
		OperationID: "get-leaderboard-times",
	}, HandleGetLeaderboardTimes)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/players/{player_id}",
		Tags:        []string{"times"},
		Summary:     "get player PRs",
		Description: "get a player's PRs for all events",
		OperationID: "get-player-prs",
	}, HandleGetPlayerPRs)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/leaderboards/{leaderboard_id}/prizepool",
		Tags:        []string{"times"},
		Summary:     "get prizepool",
		Description: "get an event leaderboard's prizepool",
		OperationID: "get-leaderboard-prizepool",
	}, HandleGetLeaderboardPrizepool)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/{event_id}/prizepool",
		Tags:        []string{"times"},
		Summary:     "get prizepool total",
		Description: "get an event's prizepool total",
		OperationID: "get-prizepool-total",
	}, HandleGetPrizepoolTotal)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/stats",
		Summary:     "get site stats",
		Description: "get site stats",
		OperationID: "get-stats",
	}, HandleGetStats)
}

func RegisterSessionRoutes(sessionApi *huma.Group) {
	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/self",
		Tags:        []string{"players"},
		Summary:     "get player self",
		Description: "get a player self from session",
		OperationID: "get-player-self",
	}, HandleGetPlayerSelf)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/timeslot/{event_id}",
		Tags:        []string{"players"},
		Summary:     "get timeslot info for an event",
		Description: "get your timeslot info for an event (or without an event)",
		OperationID: "get-timeslot-info",
	}, HandleGetTimeslotInfo)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/tempusid/{tempus_id}",
		Tags:        []string{"players"},
		Summary:     "set Tempus ID",
		Description: "set your Tempus ID and country",
		OperationID: "set-tempus-id",
	}, HandleSetTempusID)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/tradetoken/{steam_trade_url}",
		Tags:        []string{"player"},
		Summary:     "set Steam trade token",
		Description: "set your own Steam trade token from your Steam Trade URL, found at https://steamcommunity.com/id/{steamid}/tradeoffers/privacy",
		OperationID: "set-trade-token",
	}, HandleSetTradeToken)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/class/{player_class}",
		Tags:        []string{"players"},
		Summary:     "update class pref",
		Description: "update your class preference",
		OperationID: "update-class-pref",
	}, HandleUpdateClassPref)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/map/{map_name}",
		Tags:        []string{"players"},
		Summary:     "update map pref",
		Description: "update your map preference",
		OperationID: "update-map-pref",
	}, HandleUpdateMapPref)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/launcher/{launcher}",
		Tags:        []string{"players"},
		Summary:     "update launcher pref",
		Description: "update your soldier launcher preference",
		OperationID: "update-launcher-pref",
	}, HandleUpdateLauncherPref)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards/{leaderboard_id}/times",
		Tags:        []string{"times"},
		Summary:     "submit a time",
		Description: "submit a time for an event's leaderboard",
		OperationID: "submit-time",
	}, HandleSubmitTime)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards/{leaderboard_id}/times/{run_time}",
		Tags:        []string{"times"},
		Summary:     "submit an unverified time",
		Description: "submit a time manually for an event's leaderboard",
		OperationID: "submit-unverified-time",
	}, HandleSubmitUnverifiedTime)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/{event_id}/leaderboards/times/pr",
		Tags:        []string{"times"},
		Summary:     "get event PR",
		Description: "get your PR for an event",
		OperationID: "get-event-pr",
	}, HandleGetEventPR)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/requests/{request_kind}/{content}",
		Tags:        []string{"players"},
		Summary:     "submit a request",
		Description: "submit an alias or div update request",
		OperationID: "submit-request",
	}, HandleSubmitRequest)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/requests",
		Tags:        []string{"players"},
		Summary:     "get requests",
		Description: "get your own pending requests",
		OperationID: "get-requests",
	}, HandleGetRequestsSelf)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/motw/timeslots/{timeslot_id}",
		Tags:        []string{"players"},
		Summary:     "update timeslot pref",
		Description: "update your map of the week timeslot preference",
		OperationID: "update-timeslot-pref",
	}, HandleUpdateTimeslotPref)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/{event_kind}/{kind_id}",
		Tags:        []string{"events"},
		Summary:     "get motw",
		Description: "get a motw accounting for player timeslot",
		OperationID: "get-motw",
	}, HandleGetMotw)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events/leaderboards/{leaderboard_id}/times",
		Tags:        []string{"times"},
		Summary:     "get motw leaderboard times",
		Description: "get all times for an motw's leaderboard",
		OperationID: "get-motw-leaderboard-times",
	}, HandleGetMotwLeaderboardTimes)
}

func RegisterModRoutes(modApi *huma.Group) {
	huma.Register(modApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/player/{player_id}",
		Tags:        []string{"mod"},
		Summary:     "get full player",
		Description: "get full player",
		OperationID: "get-full-player",
	}, HandleGetFullPlayer)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players",
		Tags:        []string{"mod"},
		Summary:     "get all players",
		Description: "get all players",
		OperationID: "get-full-players",
	}, HandleGetFullPlayers)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/{player_id}/{player_class}/{div}",
		Tags:        []string{"mod"},
		Summary:     "update div",
		Description: "update player div",
		OperationID: "update-div",
	}, HandleUpdatePlayerDiv)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/{player_id}/{alias}",
		Tags:        []string{"mod"},
		Summary:     "update alias",
		Description: "update player div",
		OperationID: "update-alias",
	}, HandleUpdatePlayerAlias)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards/{leaderboard_id}/times/{duration}/{player_id}",
		Tags:        []string{"mod"},
		Summary:     "submit player time",
		Description: "submit a player's time for an event's leaderboard (up to one week after event end)",
		OperationID: "submit-player-time",
	}, HandleSubmitPlayerTime)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards/times/{time_id}",
		Tags:        []string{"mod"},
		Summary:     "verify player time",
		Description: "verify a player's time for an event's leaderboard",
		OperationID: "verify-player-time",
	}, HandleVerifyPlayerTime)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/events/leaderboards/times/{time_id}",
		Tags:        []string{"mod"},
		Summary:     "delete player time",
		Description: "delete a player's unverified time for an event's leaderboard",
		OperationID: "delete-player-time",
	}, HandleDeletePlayerTime)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/players/requests",
		Tags:        []string{"players"},
		Summary:     "get pending player requests",
		Description: "get all pending player requests",
		OperationID: "get-all-requests",
	}, HandleGetRequests)

	huma.Register(modApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/requests/{request_id}",
		Tags:        []string{"mod"},
		Summary:     "resolve player request",
		Description: "resolve a player's request as no longer pending",
		OperationID: "resolve-request",
	}, HandleResolveRequest)
}

func RegisterAdminRoutes(adminApi *huma.Group) {
	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/events",
		Tags:        []string{"admin"},
		Summary:     "get all events",
		Description: "get all events",
		OperationID: "get-full-events",
	}, HandleGetFullEvents)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/players/{player_id}/role/{role}",
		Tags:        []string{"admin"},
		Summary:     "update role",
		Description: "update player role",
		OperationID: "update-role",
	}, HandleUpdatePlayerRole)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/create",
		Tags:        []string{"admin"},
		Summary:     "create event",
		Description: "create an event (competition)",
		OperationID: "create-event",
	}, HandleCreateEvent)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/update",
		Tags:        []string{"admin"},
		Summary:     "update event",
		Description: "update an event (competition)",
		OperationID: "update-event",
	}, HandleUpdateEvent)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/events/{event_id}",
		Tags:        []string{"admin"},
		Summary:     "cancel event",
		Description: "cancel an event (competition) that hasn't started",
		OperationID: "cancel-event",
	}, HandleCancelEvent)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards",
		Tags:        []string{"admin"},
		Summary:     "update leaderboards",
		Description: "update an event's leaderboards",
		OperationID: "update-leaderboards",
	}, HandleUpdateLeaderboards)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/maps",
		Tags:        []string{"admin"},
		Summary:     "update maps",
		Description: "update map list from Tempus",
		OperationID: "update-maps",
	}, HandleUpdateMaps)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/motw/timeslots",
		Tags:        []string{"admin"},
		Summary:     "update timeslot",
		Description: "update motw timeslot",
		OperationID: "update-timeslot",
	}, HandleUpdateTimeslot)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/leaderboards/{leaderboard_id}/prizepool",
		Tags:        []string{"admin"},
		Summary:     "update prizepool",
		Description: "update an event leaderboard's prizepool",
		OperationID: "update-leaderboard-prizepool",
	}, HandleUpdateLeaderboardPrizepool)

	huma.Register(adminApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/events/{event_id}/refresh-results",
		Tags:        []string{"admin"},
		Summary:     "update event results",
		Description: "update an event's prizes / points after it's ended",
		OperationID: "update-event-results",
	}, HandleUpdateEventResults)
}

func RegisterDevRoutes(devApi *huma.Group) {
	huma.Register(devApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "events/leaderboards/{leaderboard_id}",
		Tags:        []string{"dev"},
		Summary:     "update leaderboard Tempus times",
		Description: "update Tempus time IDs for a leaderboard's times",
		OperationID: "update-leaderboard-tempus-times",
	}, HandleUpdateLeaderboardTempusTimes)
}
