package remnawave

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type bulkUpdateSquadsRequestDTO struct {
	UUIDs                 []string `json:"uuids"`
	ActiveInternalSquads []string `json:"activeInternalSquads"`
}

// UpdateUserSquads updates internal squads for one or more Remnawave users.
//
// Important:
// - For assigning squads, pass one or more squad UUIDs/names or "all".
// - For disabling/removing access, do not call this with an empty squads list.
//   Some Remnawave versions return A088: "Bulk add inbounds to users error".
//   Disable user through /api/users/{uuid}/actions/disable instead.
func (c *Client) UpdateUserSquads(ctx context.Context, uuids []string, activeInternalSquads []string) error {
	if c.isStub() {
		return nil
	}

	cleanUUIDs := make([]string, 0, len(uuids))
	seen := make(map[string]struct{}, len(uuids))
	for _, uuid := range uuids {
		uuid = strings.TrimSpace(uuid)
		if uuid == "" {
			continue
		}
		key := strings.ToLower(uuid)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cleanUUIDs = append(cleanUUIDs, uuid)
	}
	if len(cleanUUIDs) == 0 {
		return nil
	}

	resolvedSquads, err := c.resolveActiveInternalSquads(ctx, activeInternalSquads)
	if err != nil {
		return err
	}

	// Empty squads are not a safe "remove access" operation for Remnawave.
	// The disable endpoint is used for that.
	if len(resolvedSquads) == 0 {
		return nil
	}

	payload := bulkUpdateSquadsRequestDTO{
		UUIDs:                 cleanUUIDs,
		ActiveInternalSquads: resolvedSquads,
	}
	if err := c.doNoResponse(ctx, http.MethodPost, "/api/users/bulk/update-squads", payload); err != nil {
		return fmt.Errorf("remnawave update user squads: %w", err)
	}
	return nil
}
