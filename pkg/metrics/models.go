package metrics

import "time"

// Metrics are the metrics for a short URL over a period of time
type Metrics struct {
	ShortURLId   string
	Visits       int64
	UniqueVisits int64
	From         time.Time
	To           time.Time
}

// Collector is used to collect metrics for a short URL before flushing them to the database
type Collector struct {
	ShortURLId string
	Visits     int64
	Visitors   map[string]struct{}
}

// UniqueVisits returns the number of unique visitors for the short URL
func (m *Collector) UniqueVisits() int64 {
	return int64(len(m.Visitors))
}

// Request represents a request to collect metrics for a short URL
type Request struct {
	ShortURLId string
	VisitorId  string
}
