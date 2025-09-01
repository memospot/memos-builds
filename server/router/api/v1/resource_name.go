package v1

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/usememos/memos/internal/util"
)

const (
	WorkspaceSettingNamePrefix = "workspace/settings/"
	UserNamePrefix             = "users/"
	MemoNamePrefix             = "memos/"
	AttachmentNamePrefix       = "attachments/"
	ReactionNamePrefix         = "reactions/"
	InboxNamePrefix            = "inboxes/"
	IdentityProviderNamePrefix = "identityProviders/"
	ActivityNamePrefix         = "activities/"
	WebhookNamePrefix          = "webhooks/"
)

// GetNameParentTokens returns the tokens from a resource name.
func GetNameParentTokens(name string, tokenPrefixes ...string) ([]string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2*len(tokenPrefixes) {
		return nil, errors.Errorf("invalid request %q", name)
	}

	var tokens []string
	for i, tokenPrefix := range tokenPrefixes {
		if fmt.Sprintf("%s/", parts[2*i]) != tokenPrefix {
			return nil, errors.Errorf("invalid prefix %q in request %q", tokenPrefix, name)
		}
		if parts[2*i+1] == "" {
			return nil, errors.Errorf("invalid request %q with empty prefix %q", name, tokenPrefix)
		}
		tokens = append(tokens, parts[2*i+1])
	}
	return tokens, nil
}

func ExtractWorkspaceSettingKeyFromName(name string) (string, error) {
	const prefix = "workspace/settings/"
	if !strings.HasPrefix(name, prefix) {
		return "", errors.Errorf("invalid workspace setting name: expected prefix %q, got %q", prefix, name)
	}

	settingKey := strings.TrimPrefix(name, prefix)
	if settingKey == "" {
		return "", errors.Errorf("invalid workspace setting name: empty setting key in %q", name)
	}

	// Ensure there are no additional path segments
	if strings.Contains(settingKey, "/") {
		return "", errors.Errorf("invalid workspace setting name: setting key cannot contain '/' in %q", name)
	}

	return settingKey, nil
}

// ExtractUserIDFromName returns the uid from a resource name.
func ExtractUserIDFromName(name string) (int32, error) {
	tokens, err := GetNameParentTokens(name, UserNamePrefix)
	if err != nil {
		return 0, err
	}
	id, err := util.ConvertStringToInt32(tokens[0])
	if err != nil {
		return 0, errors.Errorf("invalid user ID %q", tokens[0])
	}
	return id, nil
}

// ExtractMemoUIDFromName returns the memo UID from a resource name.
// e.g., "memos/uuid" -> "uuid".
func ExtractMemoUIDFromName(name string) (string, error) {
	tokens, err := GetNameParentTokens(name, MemoNamePrefix)
	if err != nil {
		return "", err
	}
	id := tokens[0]
	return id, nil
}

// ExtractAttachmentUIDFromName returns the attachment UID from a resource name.
func ExtractAttachmentUIDFromName(name string) (string, error) {
	tokens, err := GetNameParentTokens(name, AttachmentNamePrefix)
	if err != nil {
		return "", err
	}
	id := tokens[0]
	return id, nil
}

// ExtractReactionIDFromName returns the reaction ID from a resource name.
// e.g., "reactions/123" -> 123.
func ExtractReactionIDFromName(name string) (int32, error) {
	tokens, err := GetNameParentTokens(name, ReactionNamePrefix)
	if err != nil {
		return 0, err
	}
	id, err := util.ConvertStringToInt32(tokens[0])
	if err != nil {
		return 0, errors.Errorf("invalid reaction ID %q", tokens[0])
	}
	return id, nil
}

// ExtractInboxIDFromName returns the inbox ID from a resource name.
func ExtractInboxIDFromName(name string) (int32, error) {
	tokens, err := GetNameParentTokens(name, InboxNamePrefix)
	if err != nil {
		return 0, err
	}
	id, err := util.ConvertStringToInt32(tokens[0])
	if err != nil {
		return 0, errors.Errorf("invalid inbox ID %q", tokens[0])
	}
	return id, nil
}

func ExtractIdentityProviderIDFromName(name string) (int32, error) {
	tokens, err := GetNameParentTokens(name, IdentityProviderNamePrefix)
	if err != nil {
		return 0, err
	}
	id, err := util.ConvertStringToInt32(tokens[0])
	if err != nil {
		return 0, errors.Errorf("invalid identity provider ID %q", tokens[0])
	}
	return id, nil
}

func ExtractActivityIDFromName(name string) (int32, error) {
	tokens, err := GetNameParentTokens(name, ActivityNamePrefix)
	if err != nil {
		return 0, err
	}
	id, err := util.ConvertStringToInt32(tokens[0])
	if err != nil {
		return 0, errors.Errorf("invalid activity ID %q", tokens[0])
	}
	return id, nil
}
