package shorturl

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/AvalosM/short-url-service/pkg/logging"
)

const (
	charset          = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base             = uint64(len(charset))
	shortURLIdLength = 6
)

// Storage short url persistent storage
type Storage interface {
	CreateShortURL(ctx context.Context, id string, longURL string) error
	DeleteShortURL(ctx context.Context, id string) error
	GetLongURL(ctx context.Context, id string) (string, bool, error)
}

// Cache short url cache
type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, duration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Logger ...
type Logger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// Manager short URL manager
type Manager struct {
	config  *Config
	storage Storage
	cache   Cache
	logger  Logger
}

// NewManager creates a new short URL manager
func NewManager(config *Config, storage Storage, cache Cache, logger Logger) (*Manager, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if storage == nil {
		return nil, errors.New("storage cannot be nil")
	}
	if cache == nil {
		return nil, errors.New("cache cannot be nil")
	}

	return &Manager{
		config:  config,
		storage: storage,
		cache:   cache,
		logger:  logger,
	}, nil
}

// GetLongURL retrieves the long URL for the given short URL id
func (m *Manager) GetLongURL(ctx context.Context, shortURLId string) (string, error) {
	longURL, found, err := m.cache.Get(ctx, shortURLId)
	if err != nil {
		m.logger.Error("failed to get long URL from cache", logging.ShortURLIdKey, shortURLId, logging.ErrorKey, err)
	}
	if found {
		return longURL, nil
	}

	longURL, found, err = m.storage.GetLongURL(ctx, shortURLId)
	if err != nil {
		m.logger.Error("failed to get long URL from storage", logging.ShortURLIdKey, shortURLId, logging.ErrorKey, err)

		return "", fmt.Errorf("failed to get long URL from storage: %w", err)
	}
	if !found {
		m.logger.Debug("short URL not found", logging.ShortURLIdKey, shortURLId)

		return "", ErrShortURLNotFound
	}

	go func(ctx context.Context) {
		if err := m.cache.Set(ctx, shortURLId, longURL, time.Duration(m.config.ShortURLCacheTTLInSeconds)*time.Second); err != nil {
			m.logger.Error("failed to set long URL in cache", logging.ShortURLIdKey, shortURLId, logging.ErrorKey, err)
		}
	}(context.WithoutCancel(ctx))

	return longURL, nil
}

// CreateShortURL creates a short URL id for the given long URL
func (m *Manager) CreateShortURL(ctx context.Context, longURL string) (string, error) {
	err := validateLongURL(longURL)
	if err != nil {
		m.logger.Info("invalid long URL", logging.LongURLKey, longURL, logging.ErrorKey, err)

		return "", ErrInvalidLongURL
	}

	id, err := m.GenerateShortURLId(ctx, longURL)
	if err != nil {
		if errors.Is(err, ErrShortURLExists) {
			return id, nil
		}

		return "", err
	}

	if err := m.storage.CreateShortURL(ctx, id, longURL); err != nil {
		m.logger.Error("failed to create short URL in storage", logging.ShortURLIdKey, id, logging.LongURLKey, longURL, logging.ErrorKey, err)

		return "", fmt.Errorf("failed to create short URL in storage: %w", err)
	}

	return id, nil
}

func validateLongURL(longURL string) error {
	if longURL == "" {
		return errors.New("long URL cannot be empty")
	}
	if !strings.HasPrefix(longURL, "https://") {
		return errors.New("long URLs must start with https://")
	}

	return nil
}

// DeleteShortURL deletes the short URL with the given id
func (m *Manager) DeleteShortURL(ctx context.Context, shortURLId string) error {
	if shortURLId == "" {
		return errors.New("short URL ID cannot be empty")
	}

	// Remove from storage
	if err := m.storage.DeleteShortURL(ctx, shortURLId); err != nil {
		m.logger.Error("failed to delete short URL from storage", logging.ShortURLIdKey, shortURLId, logging.ErrorKey, err)

		return fmt.Errorf("failed to delete short URL from storage: %w", err)
	}

	// Remove from cache
	if err := m.cache.Delete(ctx, shortURLId); err != nil {
		m.logger.Error("failed to delete short URL from cache", logging.ShortURLIdKey, shortURLId, logging.ErrorKey, err)

		return fmt.Errorf("failed to delete short URL from cache: %w", err)
	}

	return nil
}

// GenerateShortURLId generates a unique short URL ID for the given long URL
func (m *Manager) GenerateShortURLId(ctx context.Context, longURL string) (string, error) {
	if longURL == "" {
		return "", errors.New("long URL cannot be empty")
	}

	for offset := 0; offset < m.config.MaxShortURLIdRetries; offset++ {
		id, err := m.GenerateIdWithOffset(longURL, uint(offset))
		if err != nil {
			m.logger.Error("failed to generate short URL ID with offset", logging.LongURLKey, longURL, logging.ErrorKey, err)

			return "", fmt.Errorf("failed to generate short URL ID with offset: %w", err)
		}

		storedLongURL, found, err := m.storage.GetLongURL(ctx, id)
		if err != nil {
			m.logger.Error("error checking existing short URL", logging.ShortURLIdKey, id, logging.ErrorKey, err)

			return "", fmt.Errorf("error checking existing short URL: %w", err)
		}
		if !found {
			return id, nil
		}
		if storedLongURL == longURL {
			return id, ErrShortURLExists
		}

		m.logger.Debug("collision detected for short URL", logging.ShortURLIdKey, id, logging.LongURLKey, longURL)
	}

	m.logger.Error("failed to generate unique short URL", logging.LongURLKey, longURL)

	return "", fmt.Errorf("failed to generate unique short URL")
}

// GenerateIdWithOffset creates an id with the given long URL and offset
func (m *Manager) GenerateIdWithOffset(longURL string, offset uint) (string, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(longURL))
	if err != nil {
		return "", fmt.Errorf("failed to hash long URL: %w", err)
	}

	hashValue := h.Sum64()

	// https://en.wikipedia.org/wiki/Quadratic_probing
	// m = 2^64
	// h(k,i) = h(k) + i/2 + i^2/2
	quadraticOffset := (offset + (offset * offset)) / 2
	hashValue += uint64(quadraticOffset)

	var id strings.Builder
	id.Grow(shortURLIdLength)
	for range shortURLIdLength {
		id.WriteByte(charset[hashValue%base])
		hashValue /= base
	}

	return id.String(), nil
}
