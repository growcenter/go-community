package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type FeatureFlagUsecase interface {
	IsFeatureEnabled(ctx context.Context, key string, communityId string) (bool, error)
}

type featureFlagUsecase struct {
	r pgsql.PostgreRepositories
}

func NewFeatureFlagUsecase(r pgsql.PostgreRepositories) *featureFlagUsecase {
	return &featureFlagUsecase{
		r: r,
	}
}

// GetAllFlags returns all feature flags
func (ffu *featureFlagUsecase) GetAll(ctx context.Context) (response []models.FeatureFlagResponse, err error) {
	flags, err := ffu.r.FeatureFlag.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, flag := range flags {
		response = append(response, models.FeatureFlagResponse{
			Type:        models.TYPE_FEATURE_FLAG,
			Name:        flag.Name,
			Key:         flag.Key,
			Description: flag.Description,
			Enabled:     flag.Enabled,
			Rules:       flag.Rules,
		})
	}

	return response, nil
}

// GetFlag returns a specific feature flag by key
func (ffu *featureFlagUsecase) GetByKey(ctx context.Context, key string) (response *models.FeatureFlagResponse, err error) {
	flag, err := ffu.r.FeatureFlag.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return &models.FeatureFlagResponse{
		Type:        models.TYPE_FEATURE_FLAG,
		Name:        flag.Name,
		Key:         flag.Key,
		Description: flag.Description,
		Enabled:     flag.Enabled,
		Rules:       flag.Rules,
	}, nil
}

// CreateFlag creates a new feature flag
func (ffu *featureFlagUsecase) Create(ctx context.Context, request models.FeatureFlagRequest) (response *models.FeatureFlagResponse, err error) {
	// Normalize the key (lowercase, no spaces)
	request.Key = common.StringTrimSpaceAndLower(strings.ReplaceAll(request.Key, " ", "_"))

	// Check if flag with this key already exists
	exist, err := ffu.r.FeatureFlag.GetByKey(ctx, request.Key)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorDataNotFound
	}

	var rules *models.Rules
	if request.Rules != nil {
		rules = &models.Rules{
			Percentage:   request.Rules.Percentage,
			CommunityIds: request.Rules.CommunityIds,
			Parameters:   request.Rules.Parameters,
		}
	}

	flag := models.FeatureFlag{
		Name:        request.Name,
		Key:         request.Key,
		Description: request.Description,
		Enabled:     request.Enabled,
		Rules:       rules,
	}

	err = ffu.r.FeatureFlag.Create(ctx, flag)
	if err != nil {
		return nil, err
	}

	return &models.FeatureFlagResponse{
		Type:        models.TYPE_FEATURE_FLAG,
		Name:        flag.Name,
		Key:         flag.Key,
		Description: flag.Description,
		Enabled:     flag.Enabled,
		Rules:       flag.Rules,
	}, nil
}

// UpdateFlag updates an existing feature flag
func (ffu *featureFlagUsecase) Update(ctx context.Context, key string, request models.FeatureFlagRequest) (response *models.FeatureFlagResponse, err error) {
	// Normalize the key (lowercase, no spaces)
	request.Key = strings.ToLower(strings.ReplaceAll(request.Key, " ", "_"))
	if common.StringTrimSpaceAndLower(key) != request.Key {
		return nil, models.ErrorInvalidInput
	}

	flag, err := ffu.r.FeatureFlag.GetByKey(ctx, request.Key)
	if err != nil {
		return nil, models.ErrorDataNotFound
	}

	flag.Name = request.Name
	flag.Key = request.Key
	flag.Description = request.Description
	flag.Enabled = request.Enabled
	if request.Rules != nil {
		flag.Rules = request.Rules
	}

	err = ffu.r.FeatureFlag.Update(ctx, &flag)
	if err != nil {
		return nil, err
	}

	return &models.FeatureFlagResponse{
		Type:        models.TYPE_FEATURE_FLAG,
		Name:        flag.Name,
		Key:         flag.Key,
		Description: flag.Description,
		Enabled:     flag.Enabled,
		Rules:       flag.Rules,
	}, nil
}

// ToggleFlag enables or disables a feature flag
func (ffu *featureFlagUsecase) Toggle(ctx context.Context, key string, enabled bool) (err error) {
	return ffu.r.FeatureFlag.Toggle(ctx, key, enabled)
}

// DeleteFlag removes a feature flag
func (ffu *featureFlagUsecase) Delete(ctx context.Context, key string) error {
	return ffu.r.FeatureFlag.Delete(ctx, key)
}

// IsFeatureEnabled checks if a feature flag is enabled for a specific context
func (ffu *featureFlagUsecase) IsFeatureEnabled(ctx context.Context, key string, communityId string) (bool, error) {
	flag, err := ffu.r.FeatureFlag.GetByKey(ctx, key)
	if err != nil {
		return false, err
	}

	// If flag is disabled, return immediately
	if !flag.Enabled {
		return false, nil
	}

	// If no rules, the flag is enabled for everyone
	if flag.Rules == nil {
		return true, nil
	}

	// Check user-specific rules
	if flag.Rules.CommunityIds != nil {
		for _, id := range flag.Rules.CommunityIds {
			if id == communityId {
				return true, nil
			}
		}
	}

	// Check percentage-based rollout
	if flag.Rules.Percentage != nil {
		// In a real implementation, you'd use a consistent hashing function
		// to ensure the same user always gets the same result
		// Here's a simplified version:
		if communityId != "" {
			// Simple hash function for demonstration
			hash := 0
			for _, c := range communityId {
				hash += int(c)
			}
			percentage := hash % 100
			return percentage < *flag.Rules.Percentage, nil
		}
	}

	// Default to enabled if flag is enabled and no rules matched
	return true, nil
}
