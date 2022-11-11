// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: query.sql

package client

import (
	"context"

	"github.com/google/uuid"
)

const getUserByEmail = `-- name: GetUserByEmail :one
select
    id,
    email
from auth.users
where email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i AuthUser
	err := row.Scan(&i.ID, &i.Email)
	return i, err
}

const getUserOAuthToken = `-- name: GetUserOAuthToken :one
select
    user_id,
    provider,
    token,
    created_at,
    updated_at
from public.user_oauth_token
where user_id = $1 and provider = $2
`

type GetUserOAuthTokenParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Provider string    `json:"provider"`
}

func (q *Queries) GetUserOAuthToken(ctx context.Context, arg GetUserOAuthTokenParams) (UserOauthToken, error) {
	row := q.db.QueryRowContext(ctx, getUserOAuthToken, arg.UserID, arg.Provider)
	var i UserOauthToken
	err := row.Scan(
		&i.UserID,
		&i.Provider,
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listOAuthTokensByProvider = `-- name: ListOAuthTokensByProvider :many
select
    user_id,
    provider,
    token,
    created_at,
    updated_at
from public.user_oauth_token
where provider = $1
`

func (q *Queries) ListOAuthTokensByProvider(ctx context.Context, provider string) ([]UserOauthToken, error) {
	rows, err := q.db.QueryContext(ctx, listOAuthTokensByProvider, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UserOauthToken
	for rows.Next() {
		var i UserOauthToken
		if err := rows.Scan(
			&i.UserID,
			&i.Provider,
			&i.Token,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const uspertUserEmailSyncHistory = `-- name: UspertUserEmailSyncHistory :exec
insert into public.user_email_sync_history(user_id, history_id)
values ($1, $2)
on conflict (user_id) do update set history_id = excluded.history_id
`

type UspertUserEmailSyncHistoryParams struct {
	UserID    uuid.UUID `json:"user_id"`
	HistoryID int64     `json:"history_id"`
}

func (q *Queries) UspertUserEmailSyncHistory(ctx context.Context, arg UspertUserEmailSyncHistoryParams) error {
	_, err := q.db.ExecContext(ctx, uspertUserEmailSyncHistory, arg.UserID, arg.HistoryID)
	return err
}
