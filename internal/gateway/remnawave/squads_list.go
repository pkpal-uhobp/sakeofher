package remnawave

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type internalSquadDTO struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

func (c *Client) ListInternalSquads(ctx context.Context) ([]internalSquadDTO, error) {
	if c.isStub() {
		return []internalSquadDTO{{UUID: "00000000-0000-4000-8000-000000000001", Name: "Default-Squad", IsActive: true}}, nil
	}

	var raw json.RawMessage
	if err := c.doJSON(ctx, http.MethodGet, "/api/internal-squads", nil, &raw); err != nil {
		return nil, fmt.Errorf("remnawave list internal squads: %w", err)
	}

	out, err := decodeInternalSquads(raw)
	if err != nil {
		return nil, fmt.Errorf("decode remnawave internal squads: %w", err)
	}
	return out, nil
}

func decodeInternalSquads(raw json.RawMessage) ([]internalSquadDTO, error) {
	var direct []internalSquadDTO
	if err := json.Unmarshal(raw, &direct); err == nil && direct != nil {
		return normalizeInternalSquads(direct), nil
	}

	var object struct {
		InternalSquads []internalSquadDTO `json:"internalSquads"`
		Squads         []internalSquadDTO `json:"squads"`
		Items          []internalSquadDTO `json:"items"`
		Data           []internalSquadDTO `json:"data"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}

	switch {
	case object.InternalSquads != nil:
		return normalizeInternalSquads(object.InternalSquads), nil
	case object.Squads != nil:
		return normalizeInternalSquads(object.Squads), nil
	case object.Items != nil:
		return normalizeInternalSquads(object.Items), nil
	case object.Data != nil:
		return normalizeInternalSquads(object.Data), nil
	default:
		return []internalSquadDTO{}, nil
	}
}

func normalizeInternalSquads(input []internalSquadDTO) []internalSquadDTO {
	out := make([]internalSquadDTO, 0, len(input))
	for _, squad := range input {
		squad.UUID = strings.TrimSpace(squad.UUID)
		squad.Name = strings.TrimSpace(squad.Name)
		if squad.UUID == "" {
			continue
		}
		out = append(out, squad)
	}
	return out
}
