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

-- name: ListViolations :many
SELECT id, user_id, type, description, lat, lng, status, confirmations_count, created_at, updated_at
FROM violations
WHERE
  ($1::text IS NULL OR type = $1) AND
  ($2::text IS NULL OR status = $2) AND
  ($3::timestamptz IS NULL OR created_at >= $3) AND
  ($4::timestamptz IS NULL OR created_at <= $4) AND
  (
    ($5::float8 IS NULL OR $6::float8 IS NULL OR $7::float8 IS NULL OR $8::float8 IS NULL)
    OR (lng BETWEEN $5 AND $7 AND lat BETWEEN $6 AND $8)
  )
ORDER BY created_at DESC
LIMIT $9 OFFSET $10;

-- name: CountViolations :one
SELECT count(1)
FROM violations
WHERE
  ($1::text IS NULL OR type = $1) AND
  ($2::text IS NULL OR status = $2) AND
  ($3::timestamptz IS NULL OR created_at >= $3) AND
  ($4::timestamptz IS NULL OR created_at <= $4) AND
  (
    ($5::float8 IS NULL OR $6::float8 IS NULL OR $7::float8 IS NULL OR $8::float8 IS NULL)
    OR (lng BETWEEN $5 AND $7 AND lat BETWEEN $6 AND $8)
  );

-- Violation Requests

-- name: CreateViolationRequest :one
INSERT INTO violation_requests (id, violation_id, status, created_by_user_id, comment)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetViolationRequestsByViolationID :many
SELECT * FROM violation_requests WHERE violation_id = $1 ORDER BY created_at;

-- name: GetViolationRequestByID :one
SELECT * FROM violation_requests WHERE id = $1;

-- name: AddRequestPhoto :one
INSERT INTO violation_request_photos (id, request_id, url, thumb_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRequestPhotosByRequestID :many
SELECT * FROM violation_request_photos WHERE request_id = $1 ORDER BY created_at;

-- name: UpdateViolationStatus :one
UPDATE violations
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- Violation chat messages

-- name: CreateViolationChatMessage :one
INSERT INTO violation_chat_messages (id, violation_id, user_id, text, is_system)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListViolationChatMessages :many
SELECT *
FROM violation_chat_messages
WHERE violation_id = $1
ORDER BY created_at ASC, id ASC
LIMIT $2 OFFSET $3;

-- name: UpdateViolationChatMessage :one
UPDATE violation_chat_messages
SET text = $1, updated_at = NOW()
WHERE id = $2 AND user_id = $3 AND is_system = FALSE
RETURNING *;

-- name: DeleteViolationChatMessage :one
DELETE FROM violation_chat_messages
WHERE id = $1 AND user_id = $2 AND is_system = FALSE
RETURNING violation_id;

-- Violation votes

-- name: UpsertViolationVote :one
INSERT INTO violation_votes (id, violation_id, user_id, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (violation_id, user_id) DO UPDATE
SET value = EXCLUDED.value, updated_at = NOW()
RETURNING *;

-- name: UpsertViolationRequestVote :one
INSERT INTO violation_votes (id, violation_id, request_id, user_id, value)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (request_id, user_id) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW(),
    request_id = EXCLUDED.request_id
RETURNING *;

-- name: DeleteViolationVote :exec
DELETE FROM violation_votes
WHERE violation_id = $1 AND user_id = $2;

-- name: GetViolationVotesAggregated :one
SELECT
    violation_id,
    COALESCE(sum(CASE WHEN value = 'like' THEN 1 ELSE 0 END), 0)   AS likes,
    COALESCE(sum(CASE WHEN value = 'dislike' THEN 1 ELSE 0 END), 0) AS dislikes,
    COALESCE(
        max(
            CASE
                WHEN user_id = $2 THEN value
                ELSE NULL
            END
        ),
        ''
    ) AS user_vote
FROM violation_votes
WHERE violation_id = $1
GROUP BY violation_id;

-- Violation complaints

-- name: CreateViolationComplaint :one
INSERT INTO violation_complaints (id, violation_id, user_id, reason, message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateViolationRequestComplaint :one
INSERT INTO violation_complaints (id, violation_id, request_id, user_id, reason, message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

