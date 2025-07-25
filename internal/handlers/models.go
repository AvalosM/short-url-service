package handlers

import (
	"time"

	"github.com/AvalosM/meli-interview-challenge/pkg/metrics"
)

// ShortURLRequest ...
type ShortURLRequest struct {
	LongURL string `json:"long_url"`
}

// ShortURLMetricsRequest ...
type ShortURLMetricsRequest struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ShortURLMetricsResponse ...
type ShortURLMetricsResponse struct {
	Visits       int64 `json:"visits"`
	UniqueVisits int64 `json:"unique_visits"`
}

// NewShortURLMetricsResponse creates a new ShortURLMetricsResponse from the given metrics
func NewShortURLMetricsResponse(metrics *metrics.Metrics) *ShortURLMetricsResponse {
	return &ShortURLMetricsResponse{
		Visits:       metrics.Visits,
		UniqueVisits: metrics.UniqueVisits,
	}
}
