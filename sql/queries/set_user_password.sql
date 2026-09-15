-- name: UpdateUserPassword :one
UPDATE users SET hashed_password = $2, updated_at= NOW() WHERE id = $1 RETURNING *;