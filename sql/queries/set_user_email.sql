-- name: UpdateUserMail :one
UPDATE users SET email = $2, updated_at= NOW() WHERE id = $1 RETURNING *;