-- name: CreateUser :one
INSERT INTO public.users(login, password) 
VALUES ($1, $2)
RETURNING "ID";

-- name: GetUser :one
select
	U."ID",
	U.LOGIN as login,
	U.PASSWORD as password
from
	users u
where
	u.login = $1;

-- name: SaveContent :one
INSERT INTO public.content(
	user_id, type, data, meta)
	VALUES ($1, $2, $3, $4)
    RETURNING id;

-- name: DeleteContent :exec
DELETE FROM public."content"
WHERE id = $1;

-- name: UpdateContent :exec
UPDATE public."content"
SET user_id = $1, "type" = $2, "data" = $3, meta = $4
WHERE id=$5;

-- name: GetUserContentByID :one
select
	id,	user_id, "type", "data",	meta
from
	public."content" c
where
	c.id = $1 and c.user_id = $2;

-- name: GetAllUserContent :many
select
	id,	user_id, "type", "data",	meta
from
	public."content" c
where
    c.user_id = $1;

-- name: GetUserContentByType :many
select
	id,	user_id, "type", "data",	meta
from
	public."content" c
where
	c.user_id = $1 and c."type" =$2;

