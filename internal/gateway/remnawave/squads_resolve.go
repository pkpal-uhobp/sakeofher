package remnawave

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) resolveActiveInternalSquads(ctx context.Context, input []string) ([]string, error) {
	if input == nil {
		return nil, nil
	}

	items := normalizeSquadInput(input)
	if len(items) == 0 {
		return []string{}, nil
	}

	if requestsAllSquads(items) {
		squads, err := c.ListInternalSquads(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(squads))
		for _, squad := range squads {
			if !squad.IsActive {
				continue
			}
			uuid := strings.TrimSpace(squad.UUID)
			if uuid != "" {
				out = append(out, uuid)
			}
		}
		return out, nil
	}

	allAreUUID := true
	for _, item := range items {
		if !looksLikeUUID(item) {
			allAreUUID = false
			break
		}
	}
	if allAreUUID {
		return items, nil
	}

	squads, err := c.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]string, len(squads))
	byUUID := make(map[string]string, len(squads))
	for _, squad := range squads {
		uuid := strings.TrimSpace(squad.UUID)
		if uuid == "" {
			continue
		}
		byUUID[strings.ToLower(uuid)] = uuid
		name := strings.TrimSpace(squad.Name)
		if name != "" {
			byName[strings.ToLower(name)] = uuid
		}
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if uuid, ok := byUUID[key]; ok {
			out = append(out, uuid)
			continue
		}
		if uuid, ok := byName[key]; ok {
			out = append(out, uuid)
			continue
		}
		return nil, fmt.Errorf("remnawave internal squad %q not found; use squad UUID, exact squad name, or REMNAWAVE_DEFAULT_INTERNAL_SQUADS=all", item)
	}
	return uniqueStrings(out), nil
}

func normalizeSquadInput(input []string) []string {
	out := make([]string, 0, len(input))
	for _, raw := range input {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return uniqueStrings(out)
}

func requestsAllSquads(items []string) bool {
	if len(items) != 1 {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(items[0])) {
	case "all", "*", "все":
		return true
	default:
		return false
	}
}

func uniqueStrings(input []string) []string {
	out := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func looksLikeUUID(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for i, ch := range value {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
	}
	return true
}
