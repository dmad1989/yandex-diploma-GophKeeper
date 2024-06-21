-- name: CreateUser :one
INSERT INTO public.users(login, password) 
VALUES ($1, $2)
RETURNING "ID";