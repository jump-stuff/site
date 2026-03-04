-- name: CountPlayers :one
select count(*) from player;

-- name: CountTimes :one
select count(*) from time;

-- name: CountEvents :one
select count(*) from event;