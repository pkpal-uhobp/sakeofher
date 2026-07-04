package remnawave

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sakeofher/internal/domain"
)

func (c *Client) ListInternalSquads(ctx context.Context) ([]domain.RemnaInternalSquad, error) {
	if c.isStub() {
		return []domain.RemnaInternalSquad{}, nil
	}

	var raw any
	if err := c.doJSON(ctx, http.MethodGet, "/api/internal-squads", nil, &raw); err != nil {
		return nil, fmt.Errorf("remnawave list internal squads: %w", err)
	}

	seen := make(map[string]struct{})
	out := make([]domain.RemnaInternalSquad, 0)

	collectInternalSquads(raw, seen, &out)

	return out, nil
}

func (c *Client) ListInternalSquadUUIDs(ctx context.Context) ([]string, error) {
	squads, err := c.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(squads))
	for _, squad := range squads {
		if strings.TrimSpace(squad.UUID) != "" {
			out = append(out, squad.UUID)
		}
	}

	return out, nil
}

func collectInternalSquads(value any, seen map[string]struct{}, out *[]domain.RemnaInternalSquad) {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			collectInternalSquads(item, seen, out)
		}

	case map[string]any:
		if response, ok := typed["response"]; ok {
			collectInternalSquads(response, seen, out)
			return
		}

		for _, key := range []string{"internalSquads", "internal_squads", "squads", "items", "data"} {
			if nested, ok := typed[key]; ok {
				collectInternalSquads(nested, seen, out)
			}
		}

		rawUUID, _ := typed["uuid"].(string)
		uuid := strings.TrimSpace(rawUUID)
		if uuid == "" {
			return
		}

		if _, exists := seen[uuid]; exists {
			return
		}
		seen[uuid] = struct{}{}

		name := firstString(typed, "name", "title", "squadName")
		if name == "" {
			name = uuid
		}

		isActive := true
		if value, ok := typed["isActive"].(bool); ok {
			isActive = value
		}
		if value, ok := typed["is_active"].(bool); ok {
			isActive = value
		}

		*out = append(*out, domain.RemnaInternalSquad{
			UUID: uuid,
			Name: name,
			IsActive: isActive,
		})
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key].(string); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}

	return ""
}
