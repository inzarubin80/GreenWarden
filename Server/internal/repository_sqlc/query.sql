-- name: CreateUser :one
INSERT INTO users (name)
VALUES ($1)
returning user_id;

-- name: UpdateUserName :one
UPDATE users
SET name = $1
WHERE user_id = $2
RETURNING *;


-- name: GetUsersByIDs :many
SELECT * FROM users
WHERE user_id = ANY($1::bigint[]);

-- name: GetUserByID :one
SELECT * FROM users
WHERE user_id = $1;

-- name: GetUserAuthProvidersByProviderUid :one
SELECT * FROM user_auth_providers
WHERE provider_uid = $1 AND provider = $2;

-- name: AddUserAuthProviders :one
INSERT INTO user_auth_providers (user_id, provider_uid, provider, name)
VALUES ($1, $2, $3, $4)
returning *;


