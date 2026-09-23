package sessions

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const lifetimeKey = "skid_lifetime"

func validLifetime(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == 16 && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func newLifetime() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

// Claims are serialized by Manager.mutations. A failed or unknown write never
// manufactures an addressable reference; a later inventory may read its token.
func (manager *Manager) claimWorkspace(ctx context.Context, workspace workspaceInfo) (WorkspaceTarget, error) {
	if token, exists := workspace.Tokens[lifetimeKey]; exists {
		if !validLifetime(token) {
			return WorkspaceTarget{}, errors.New("invalid workspace lifetime token")
		}
		return WorkspaceTarget{WorkspaceID: workspace.ID, IdentityToken: token}, nil
	}
	token, err := newLifetime()
	if err != nil {
		return WorkspaceTarget{}, err
	}
	workspaces, err := manager.workspaces(ctx)
	if err != nil {
		return WorkspaceTarget{}, err
	}
	found := false
	for _, current := range workspaces {
		if current.ID != workspace.ID {
			continue
		}
		found = true
		if !validWorkspaceLabel(current.Label) {
			return WorkspaceTarget{}, errors.New("workspace label changed before lifetime claim")
		}
		if existing, exists := current.Tokens[lifetimeKey]; exists {
			if !validLifetime(existing) {
				return WorkspaceTarget{}, errors.New("invalid workspace lifetime token")
			}
			return WorkspaceTarget{WorkspaceID: current.ID, IdentityToken: existing}, nil
		}
		break
	}
	if !found {
		return WorkspaceTarget{}, errors.New("workspace disappeared before lifetime claim")
	}
	if err := manager.reportWorkspace(ctx, workspace.ID, map[string]string{lifetimeKey: token}); err != nil {
		return WorkspaceTarget{}, err
	}
	workspaces, err = manager.workspaces(ctx)
	if err != nil {
		return WorkspaceTarget{}, err
	}
	for _, current := range workspaces {
		if current.ID == workspace.ID && validWorkspaceLabel(current.Label) && current.Tokens[lifetimeKey] == token {
			return WorkspaceTarget{WorkspaceID: current.ID, IdentityToken: token}, nil
		}
	}
	return WorkspaceTarget{}, errors.New("workspace lifetime token was not retained")
}

func (manager *Manager) claimPane(ctx context.Context, pane paneInfo) (TerminalTarget, error) {
	if token, exists := pane.Tokens[lifetimeKey]; exists {
		if !validLifetime(token) {
			return TerminalTarget{}, errors.New("invalid terminal lifetime token")
		}
		return TerminalTarget{TerminalID: pane.TerminalID, IdentityToken: token}, nil
	}
	token, err := newLifetime()
	if err != nil {
		return TerminalTarget{}, err
	}
	current, err := manager.pane(ctx, pane.ID)
	if err != nil || current.TerminalID != pane.TerminalID || current.WorkspaceID != pane.WorkspaceID {
		return TerminalTarget{}, errors.New("terminal changed before lifetime claim")
	}
	if existing, exists := current.Tokens[lifetimeKey]; exists {
		if !validLifetime(existing) {
			return TerminalTarget{}, errors.New("invalid terminal lifetime token")
		}
		return TerminalTarget{TerminalID: current.TerminalID, IdentityToken: existing}, nil
	}
	if err := manager.reportPane(ctx, pane.ID, map[string]string{lifetimeKey: token}); err != nil {
		return TerminalTarget{}, err
	}
	current, err = manager.pane(ctx, pane.ID)
	if err != nil || current.TerminalID != pane.TerminalID || current.WorkspaceID != pane.WorkspaceID || current.Tokens[lifetimeKey] != token {
		return TerminalTarget{}, errors.New("terminal lifetime token was not retained")
	}
	return TerminalTarget{TerminalID: pane.TerminalID, IdentityToken: token}, nil
}

func encodeObjective(objective string) map[string]string {
	if objective == "" {
		return nil
	}
	encoded := base64.RawURLEncoding.EncodeToString([]byte(objective))
	count := (len(encoded) + 79) / 80
	parts := make(map[string]string, count+1)
	parts["skid_objective_count"] = strconv.Itoa(count)
	for index := 0; index < count; index++ {
		start := index * 80
		end := min(start+80, len(encoded))
		parts[fmt.Sprintf("skid_objective_%02d", index)] = encoded[start:end]
	}
	return parts
}

func decodeObjective(tokens map[string]string) string {
	countText, found := tokens["skid_objective_count"]
	if !found {
		return ""
	}
	count, err := strconv.Atoi(countText)
	if err != nil || count < 1 || count > 16 || strconv.Itoa(count) != countText {
		return ""
	}
	var encoded strings.Builder
	for index := 0; index < count; index++ {
		part, present := tokens[fmt.Sprintf("skid_objective_%02d", index)]
		if !present || len(part) == 0 || len(part) > 80 {
			return ""
		}
		encoded.WriteString(part)
	}
	for key := range tokens {
		if strings.HasPrefix(key, "skid_objective_") && key != "skid_objective_count" {
			indexText := strings.TrimPrefix(key, "skid_objective_")
			index, err := strconv.Atoi(indexText)
			if err != nil || index < 0 || index >= count || fmt.Sprintf("%02d", index) != indexText {
				return ""
			}
		}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded.String())
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != encoded.String() || validateObjective(string(decoded)) != nil {
		return ""
	}
	return string(decoded)
}

func (manager *Manager) writeObjective(ctx context.Context, paneID, objective string) error {
	parts := encodeObjective(objective)
	if len(parts) == 0 {
		return nil
	}
	count, _ := strconv.Atoi(parts["skid_objective_count"])
	if err := manager.reportPane(ctx, paneID, map[string]string{"skid_objective_count": parts["skid_objective_count"]}); err != nil {
		return err
	}
	for start := 0; start < count; start += 16 {
		batch := make(map[string]string)
		for index := start; index < min(start+16, count); index++ {
			key := fmt.Sprintf("skid_objective_%02d", index)
			batch[key] = parts[key]
		}
		if err := manager.reportPane(ctx, paneID, batch); err != nil {
			return err
		}
	}
	current, err := manager.pane(ctx, paneID)
	if err != nil || decodeObjective(current.Tokens) != objective {
		return errors.New("objective metadata was not retained")
	}
	return nil
}
