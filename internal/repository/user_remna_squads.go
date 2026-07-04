package repository

import (
	"context"
	"fmt"
	"strings"
)

func (r *UserRepository) SetRemnaActiveSquads(ctx context.Context, userID int64, squads []string) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO user_remna_squads (user_id, active_internal_squads)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET active_internal_squads = EXCLUDED.active_internal_squads,
		    updated_at = now()
	`, userID, normalizeRepositorySquads(squads))
	if err != nil {
		return fmt.Errorf("set user remnawave squads: %w", err)
	}

	return nil
}

func (r *UserRepository) GetRemnaActiveSquads(ctx context.Context, userID int64) ([]string, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var squads []string

	err := r.tx.Querier(ctx).QueryRow(ctx, `
		SELECT active_internal_squads
		FROM user_remna_squads
		WHERE user_id = $1
	`, userID).Scan(&squads)
	if err != nil {
		// If there is no row yet, this user just does not have saved squads.
		return nil, nil
	}

	return normalizeRepositorySquads(squads), nil
}

func normalizeRepositorySquads(squads []string) []string {
	out := make([]string, 0, len(squads))
	seen := make(map[string]struct{})

	for _, item := range squads {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}
