-- name: CopyWeatherDaily :copyfrom
INSERT INTO weather_daily (
    latitude, longitude, date, timezone,
    precipitation_sum, temperature_2m_max, temperature_2m_min,
    et0_fao_evapotranspiration
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
