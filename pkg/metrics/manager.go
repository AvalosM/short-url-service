package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AvalosM/meli-interview-challenge/pkg/logging"
)

// Storage short url persistent storage
type Storage interface {
	CreateMetrics(ctx context.Context, metrics map[string]*Collector) error
	GetMetrics(ctx context.Context, shortURLId string, from, to time.Time) (*Metrics, bool, error)
}

// Logger ...
type Logger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// Manager metrics manager
type Manager struct {
	config      *Config
	storage     Storage
	collectors  map[string]*Collector
	requestChan chan Request
	stopChan    chan struct{}
	logger      Logger
}

// NewManager creates a new metrics manager
func NewManager(config *Config, storage Storage, logger Logger) (*Manager, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if storage == nil {
		return nil, errors.New("storage cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	return &Manager{
		config:      config,
		storage:     storage,
		collectors:  make(map[string]*Collector),
		requestChan: make(chan Request, config.RequestChannelSize),
		stopChan:    make(chan struct{}),
		logger:      logger,
	}, nil
}

// Start starts the metrics manager request consumer
func (m *Manager) Start() func() {
	go func() {
		ticker := time.NewTicker(time.Duration(m.config.MetricsIntervalInMS) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.logger.Debug("flushing metrics")
				m.flushMetrics()
			case request := <-m.requestChan:
				m.logger.Debug("processing request")
				m.processRequest(request)
			case <-m.stopChan:
				m.logger.Debug("flushing metrics")
				m.flushMetrics()

				return
			}
		}
	}()

	m.logger.Info("metrics manager started")

	return func() {
		close(m.stopChan)
	}
}

func (m *Manager) flushMetrics() {
	err := m.storage.CreateMetrics(context.Background(), m.collectors)
	if err != nil {
		m.logger.Error("creating metrics in storage", logging.ErrorKey, err)
	}

	clear(m.collectors)
}

func (m *Manager) processRequest(request Request) {
	m.logger.Debug("processing request")

	collector, found := m.collectors[request.ShortURLId]
	if !found {
		collector = &Collector{
			ShortURLId: request.ShortURLId,
			Visits:     1,
			Visitors:   map[string]struct{}{request.VisitorId: {}},
		}
		m.collectors[request.ShortURLId] = collector

		return
	}

	collector.Visits++
	collector.Visitors[request.VisitorId] = struct{}{}
}

// Stop stops the metrics manager and flushes any remaining metrics
func (m *Manager) Stop() {
	m.logger.Info("stopping metrics manager")
	close(m.stopChan)
}

// RecordShortURLRequestAsync records a short URL request asynchronously
func (m *Manager) RecordShortURLRequestAsync(id string, ip string) {
	go m.RecordShortURLRequest(id, ip)
}

// RecordShortURLRequest records a short URL request
func (m *Manager) RecordShortURLRequest(id string, ip string) {
	select {
	case m.requestChan <- Request{
		ShortURLId: id,
		VisitorId:  ip,
	}:
	case <-time.After(time.Millisecond * time.Duration(m.config.RecordRequestTimeoutInMS)):
		m.logger.Warn("timeout while recording short URL request")
	case <-m.stopChan:
		m.logger.Warn("metrics manager is stopping, cannot record request")
	}
}

// GetShortURLMetrics retrieves metrics for a short URL within a specified time range
func (m *Manager) GetShortURLMetrics(ctx context.Context, id string, from, to time.Time) (*Metrics, error) {
	metrics, found, err := m.storage.GetMetrics(ctx, id, from, to)
	if err != nil {
		m.logger.Error("failed to get metrics from storage", logging.ShortURLIdKey, id, logging.ErrorKey, err)

		return nil, fmt.Errorf("getting metrics from storage: %w", err)
	}
	if !found {
		return &Metrics{
			ShortURLId:   id,
			Visits:       0,
			UniqueVisits: 0,
			From:         from,
			To:           to,
		}, nil
	}

	return metrics, nil
}
