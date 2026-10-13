package types

type WeatherRecord struct {
	Latitude                 float64  `json:"latitude"`
	Longitude                float64  `json:"longitude"`
	Date                     string   `json:"date"`
	Timezone                 string   `json:"timezone"`
	PrecipitationSum         *float64 `json:"precipitation_sum"`
	Temperature2mMax         *float64 `json:"temperature_2m_max"`
	Temperature2mMin         *float64 `json:"temperature_2m_min"`
	Et0FaoEvapotranspiration *float64 `json:"et0_fao_evapotranspiration"`
}
