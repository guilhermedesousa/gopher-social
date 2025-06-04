package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `
		SELECT id, name, description, level FROM roles
		WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryDurationTimeout)
	defer cancel()

	var role Role

	err := s.db.QueryRowContext(
		ctx,
		query,
		roleName,
	).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}
