package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (r *UserRepository) SetRemnaActiveSquads(ctx context.Context, userID int64, squads []string) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	normalized := normalizeRepositorySquads(squads)

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO user_remna_squads (
		    user_id,
		    active_internal_squads,
		    desired_internal_squads,
		    sync_status
		)
		VALUES ($1, $2, $2, 'pending')
		ON CONFLICT (user_id) DO UPDATE
		SET active_internal_squads = EXCLUDED.active_internal_squads,
		    desired_internal_squads = EXCLUDED.desired_internal_squads,
		    sync_status = 'pending',
		    last_error = NULL,
		    updated_at = now()
	`, userID, normalized)
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
		return nil, nil
	}

	return normalizeRepositorySquads(squads), nil
}

func (r *UserRepository) MarkRemnaSquadsSynced(ctx context.Context, userID int64, squads []string, syncedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	normalized := normalizeRepositorySquads(squads)

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO user_remna_squads (
		    user_id,
		    active_internal_squads,
		    desired_internal_squads,
		    sync_status,
		    last_synced_at,
		    last_error
		)
		VALUES ($1, $2, $2, 'synced', $3, NULL)
		ON CONFLICT (user_id) DO UPDATE
		SET active_internal_squads = EXCLUDED.active_internal_squads,
		    desired_internal_squads = EXCLUDED.desired_internal_squads,
		    sync_status = 'synced',
		    last_synced_at = EXCLUDED.last_synced_at,
		    last_error = NULL,
		    updated_at = now()
	`, userID, normalized, syncedAt)
	if err != nil {
		return fmt.Errorf("mark remnawave squads synced: %w", err)
	}

	return nil
}

func (r *UserRepository) MarkRemnaSquadsSyncFailed(ctx context.Context, userID int64, squads []string, errorText string) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO user_remna_squads (
		    user_id,
		    active_internal_squads,
		    desired_internal_squads,
		    sync_status,
		    last_error
		)
		VALUES ($1, '{}', $2, 'failed', $3)
		ON CONFLICT (user_id) DO UPDATE
		SET desired_internal_squads = EXCLUDED.desired_internal_squads,
		    sync_status = 'failed',
		    last_error = EXCLUDED.last_error,
		    updated_at = now()
	`, userID, normalizeRepositorySquads(squads), errorText)
	if err != nil {
		return fmt.Errorf("mark remnawave squads sync failed: %w", err)
	}

	return nil
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
