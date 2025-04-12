package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type ConfigDBUsecase interface {
	GetLocationsByCampusCode(ctx context.Context, campusCode string) ([]models.GetLocationsByCampusCodeResponse, error)
	IsLocationExist(ctx context.Context, location string) (bool, error)
}

type configDBUsecase struct {
	r   pgsql.PostgreRepositories
	cfg *config.Configuration
}

func NewConfigDBUsecase(r pgsql.PostgreRepositories, cfg config.Configuration) *configDBUsecase {
	return &configDBUsecase{
		r:   r,
		cfg: &cfg,
	}
}

func (cu *configDBUsecase) GetLocationsByCampusCode(ctx context.Context, campusCode string) (response []models.GetLocationsByCampusCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	config, err := cu.r.Config.GetByKey(ctx, "campusLocationValue")
	if err != nil {
		return nil, err
	}

	if config.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	var areaLocations map[string][]string
	if err := json.Unmarshal([]byte(config.Value), &areaLocations); err != nil {
		return nil, err
	}

	_, campusExist := cu.cfg.Campus[common.StringTrimSpaceAndLower(campusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	campusCode = common.StringTrimSpaceAndLower(campusCode)
	locations, ok := areaLocations[common.StringTrimSpaceAndLower(campusCode)]
	if !ok {
		errCampusCode := fmt.Errorf("Area '%s' not found", campusCode)
		return nil, errCampusCode
	}

	var responses []models.GetLocationsByCampusCodeResponse
	for _, location := range locations {
		responses = append(responses, models.GetLocationsByCampusCodeResponse{
			Type: "location",
			Name: location,
		})
	}

	return responses, nil
}

func (cu *configDBUsecase) IsLocationExist(ctx context.Context, location string) (exist bool, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	config, err := cu.r.Config.GetByKey(ctx, "campusLocationValue")
	if err != nil {
		return false, err
	}

	if config.ID == 0 {
		return false, models.ErrorDataNotFound
	}

	var areaLocations map[string][]string
	if err := json.Unmarshal([]byte(config.Value), &areaLocations); err != nil {
		return false, err
	}

	for _, locations := range areaLocations {
		for _, loc := range locations {
			if common.StringTrimSpaceAndLower(loc) == common.StringTrimSpaceAndLower(location) {
				return true, nil
			}
		}
	}

	return false, nil
}
