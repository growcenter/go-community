package contextc

import (
	"context"
	"go-community/internal/pkg/errorc"
)

// ExtractCommunityID safely extracts the community_id from context
// Returns an error if the value is missing or not a string
func ExtractCommunityID(ctx context.Context) (string, error) {
	value := ctx.Value("community_id")
	if value == nil {
		return "", errorc.Error(errorc.ErrorUnauthorized, "community_id not found in context")
	}

	communityID, ok := value.(string)
	if !ok {
		return "", errorc.Error(errorc.ErrorInternalServer, "community_id has invalid type in context")
	}

	if communityID == "" {
		return "", errorc.Error(errorc.ErrorUnauthorized, "community_id is empty in context")
	}

	return communityID, nil
}

// ExtractString safely extracts a string value from context by key
// Returns an error if the value is missing or not a string
func ExtractString(ctx context.Context, key string) (string, error) {
	value := ctx.Value(key)
	if value == nil {
		return "", errorc.Error(errorc.ErrorUnauthorized, "%s not found in context", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", errorc.Error(errorc.ErrorInternalServer, "%s has invalid type in context", key)
	}

	if str == "" {
		return "", errorc.Error(errorc.ErrorUnauthorized, "%s is empty in context", key)
	}

	return str, nil
}

// ExtractStringSlice safely extracts a []string value from context by key
// Returns an error if the value is missing or not a []string
func ExtractStringSlice(ctx context.Context, key string) ([]string, error) {
	value := ctx.Value(key)
	if value == nil {
		return nil, errorc.Error(errorc.ErrorUnauthorized, "%s not found in context", key)
	}

	slice, ok := value.([]string)
	if !ok {
		return nil, errorc.Error(errorc.ErrorInternalServer, "%s has invalid type in context", key)
	}

	return slice, nil
}
