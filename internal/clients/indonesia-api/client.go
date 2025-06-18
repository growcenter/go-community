package indonesiaAPI

import (
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"

	"github.com/go-resty/resty/v2"
)

type Client interface {
	// GetProvinces(filterCodes []string) ([]ProvinceData, error)
	GetCities(campucCode string) ([]CityData, error)
	GetDistricts(cityCode string) ([]DistrictData, error)
}

type client struct {
	cfg config.Configuration
}

func NewClient(cfg config.Configuration) Client {
	return &client{
		cfg: cfg,
	}
}

// func (c *client) GetProvinces(filterCodes []string) ([]ProvinceData, error) {
// 	client := resty.New()
// 	url := fmt.Sprintf("%s/provinces.json", c.cfg.Clients.IndonesiaApi.Url)

// 	var response ProvinceResponse
// 	request := client.R().
// 		SetHeader("Accept", "application/json").
// 		SetResult(&response)

// 	// Add filter if codes are provided
// 	if len(filterCodes) > 0 {
// 		request.SetQueryParam("codes", strings.Join(filterCodes, ","))
// 	}

// 	_, err := request.Get(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get provinces: %w", err)
// 	}

// 	// Return all if no filter
// 	if len(filterCodes) == 0 {
// 		return response.Data, nil
// 	}

// 	// Filter results
// 	var result []ProvinceData
// 	codeMap := make(map[string]struct{}, len(filterCodes))
// 	for _, code := range filterCodes {
// 		codeMap[code] = struct{}{}
// 	}

// 	for _, province := range response.Data {
// 		if _, exists := codeMap[province.Code]; exists {
// 			result = append(result, province)
// 		}
// 	}

// 	return result, nil
// }

func (c *client) GetCities(campusCode string) ([]CityData, error) {
	client := resty.New()
	var combinedResults []CityData

	_, campusExist := c.cfg.Campus[common.StringTrimSpaceAndLower(campusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	provinceCodes, provinceExist := c.cfg.Clients.IndonesiaApi.AllowedProvince[common.StringTrimSpaceAndLower(campusCode)]
	if !provinceExist {
		return nil, models.ErrorDataNotFound
	}

	for _, code := range provinceCodes {
		url := fmt.Sprintf("%s/cities/%s.json", c.cfg.Clients.IndonesiaApi.Url, code)

		var res CityResponse
		_, err := client.R().
			SetHeader("Accept", "application/json").
			SetResult(&res).
			Get(url)

		if err != nil {
			return nil, fmt.Errorf("failed to get cities for province %v: %w", string(code), err)
		}

		combinedResults = append(combinedResults, res.Data...)
		if len(combinedResults) == 0 {
			return nil, fmt.Errorf("no data received")
		}
	}

	return combinedResults, nil
}

func (c *client) GetDistricts(cityCode string) ([]DistrictData, error) {
	client := resty.New()
	var combinedResults []DistrictData

	url := fmt.Sprintf("%s/districts/%s.json", c.cfg.Clients.IndonesiaApi.Url, cityCode)

	var res DistrictResponse
	_, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&res).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to get districts for city %v: %w", string(cityCode), err)
	}

	combinedResults = append(combinedResults, res.Data...)

	return combinedResults, nil
}
