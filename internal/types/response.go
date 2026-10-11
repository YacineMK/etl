package types

type WeatherResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`

	Daily struct {
		Time             []string   `json:"time"`
		Precipitation    []*float64 `json:"precipitation_sum"`
		TemperatureMax   []*float64 `json:"temperature_2m_max"`
		TemperatureMin   []*float64 `json:"temperature_2m_min"`
		Evapotranspiration []*float64 `json:"et0_fao_evapotranspiration"`
	} `json:"daily"`
}