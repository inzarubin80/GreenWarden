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


-- Violations

-- name: CreateViolation :one
INSERT INTO violations (id, user_id, type, description, lat, lng, status, confirmations_count)
VALUES ($1, $2, $3, $4, $5, $6, 'new', 0)
RETURNING *;

-- name: GetViolationByID :one
SELECT * FROM violations WHERE id = $1;

-- name: AddViolationPhoto :one
INSERT INTO violation_photos (id, violation_id, url, thumb_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPhotosByViolationID :many
SELECT * FROM violation_photos WHERE violation_id = $1 ORDER BY id;


