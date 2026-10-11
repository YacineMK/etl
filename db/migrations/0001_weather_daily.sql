CREATE TABLE IF NOT EXISTS weather_daily (
    latitude                   DOUBLE PRECISION NOT NULL,
    longitude                  DOUBLE PRECISION NOT NULL,
    date                       DATE NOT NULL,
    timezone                   TEXT NOT NULL,
    precipitation_sum          DOUBLE PRECISION,
    temperature_2m_max         DOUBLE PRECISION,
    temperature_2m_min         DOUBLE PRECISION,
    et0_fao_evapotranspiration DOUBLE PRECISION,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (latitude, longitude, date)
);
