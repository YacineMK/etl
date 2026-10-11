
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/YacineMK/nasa-space-apps/internal/types"
)



func FetchWeather(
	ctx context.Context,
	baseURI string,
	job types.RangeJob,
) (*types.WeatherResponse, error) {
	start, err := time.Parse(time.RFC3339, job.StartDate)
	if err != nil {
		return nil, fmt.Errorf("parse start date: %w", err)
	}

	end, err := time.Parse(time.RFC3339, job.EndDate)
	if err != nil {
		return nil, fmt.Errorf("parse end date: %w", err)
	}

	apiEnd := end.AddDate(0, 0, -1)
	if apiEnd.Before(start) {
		return nil, fmt.Errorf("invalid date range")
	}

	baseURL, err := url.Parse(baseURI)
	if err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}

	query := baseURL.Query()
	query.Set("latitude", fmt.Sprintf("%f", job.Latitude))
	query.Set("longitude", fmt.Sprintf("%f", job.Longitude))
	query.Set("start_date", start.Format("2006-01-02"))
	query.Set("end_date", apiEnd.Format("2006-01-02"))
	query.Set(
		"daily",
		"precipitation_sum,temperature_2m_max,temperature_2m_min,et0_fao_evapotranspiration",
	)
	query.Set("timezone", "Africa/Algiers")
	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL.String(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf(
			"Open-Meteo returned %s: %s",
			resp.Status,
			string(body),
		)
	}

	var weather types.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &weather, nil
}