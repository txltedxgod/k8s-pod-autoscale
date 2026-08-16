package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MetricResponse struct {
	QueueDepth   int     `json:"queue_depth"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

type Collector struct {
	endpoint   string
	httpClient *http.Client
}

func NewCollector(endpoint string) *Collector {
	return &Collector{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Collector) FetchCurrentMetric(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to reach metric endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code from metric service: %d", resp.StatusCode)
	}

	var m MetricResponse
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return 0, fmt.Errorf("error decoding metric json: %w", err)
	}

	return m.QueueDepth, nil
}
