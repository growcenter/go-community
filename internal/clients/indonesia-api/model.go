package indonesiaAPI

type (
	RegionMetadata struct {
		TotalRow int `json:"totalRow"`
	}
)

type (
	ProvinceResponse struct {
		Code     int            `json:"code"`
		Status   string         `json:"status"`
		Message  string         `json:"message"`
		Data     []ProvinceData `json:"data"`
		Metadata RegionMetadata `json:"metadata"`
	}

	ProvinceData struct {
		Type       string   `json:"type"`
		Code       string   `json:"code"`
		Indonesian string   `json:"indonesian"`
		English    string   `json:"english"`
		Aliases    []string `json:"aliases"`
	}
)

type (
	CityResponse struct {
		Code     int            `json:"code"`
		Status   string         `json:"status"`
		Message  string         `json:"message"`
		Data     []CityData     `json:"data"`
		Metadata RegionMetadata `json:"metadata"`
	}

	CityData struct {
		Type         string `json:"type"`
		Code         string `json:"code"`
		Name         string `json:"name"`
		ProvinceCode string `json:"provinceCode"`
		ProvinceName string `json:"provinceName"`
	}
)

type (
	DistrictResponse struct {
		Code     int            `json:"code"`
		Status   string         `json:"status"`
		Message  string         `json:"message"`
		Data     []DistrictData `json:"data"`
		Metadata RegionMetadata `json:"metadata"`
	}

	DistrictData struct {
		Type         string `json:"type"`
		Code         string `json:"code"`
		Name         string `json:"name"`
		ProvinceCode string `json:"provinceCode"`
		ProvinceName string `json:"provinceName"`
		CityCode     string `json:"cityCode"`
		CityName     string `json:"cityName"`
	}
)
