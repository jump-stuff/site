-- name: InsertRequest :one
insert into request (player_id, kind, content)
  values (?, ?, ?)
  returning *;

-- name: SelectPlayerRequests :many
select sqlc.embed(player), sqlc.embed(request) from request
  join player on player.id = request.player_id
  where player_id = ?
  and pending = true;

-- name: CheckPendingRequestExists :one
select exists (
  select 1
  from request
  where player_id = ?
  and kind = ?
  and pending = true);

-- name: SelectPendingRequests :many
select sqlc.embed(player), sqlc.embed(request) from request
  join player on player.id = request.player_id
  where pending = true;

-- name: ResolveRequest :exec
update request
  set pending = false
  where id = ?;