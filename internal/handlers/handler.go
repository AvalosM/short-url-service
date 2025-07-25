package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/AvalosM/short-url-service/pkg/logging"
	"github.com/AvalosM/short-url-service/pkg/metrics"
	"github.com/AvalosM/short-url-service/pkg/shorturl"
)

// ShortURLManager short url manager
type ShortURLManager interface {
	GetLongURL(ctx context.Context, shortURLId string) (string, error)
	CreateShortURL(ctx context.Context, longURL string) (string, error)
	DeleteShortURL(ctx context.Context, shortURLId string) error
}

// MetricsManager metrics manager
type MetricsManager interface {
	RecordShortURLRequestAsync(id string, ip string)
	GetShortURLMetrics(ctx context.Context, id string, from, to time.Time) (*metrics.Metrics, error)
}

// Logger ...
type Logger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// ShortURLHandler handles short URL http requests
type ShortURLHandler struct {
	shortURLManager ShortURLManager
	metricsManager  MetricsManager
	logger          Logger
}

// NewShortURLHandler creates a new ShortURLHandler
func NewShortURLHandler(shortURLManager ShortURLManager, metricsManager MetricsManager, logger Logger) (*ShortURLHandler, error) {
	if shortURLManager == nil {
		return nil, errors.New("short URL manager cannot be nil")
	}
	if metricsManager == nil {
		return nil, errors.New("metrics manager cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	return &ShortURLHandler{
		shortURLManager: shortURLManager,
		metricsManager:  metricsManager,
		logger:          logger,
	}, nil
}

// CreateShortURL godoc
//
//	@Summary      Create a short URL
//	@Description  Create a short URL for the given long URL
//	@Tags         short-url, private
//	@Accept       json
//	@Produce      json
//	@Param        ShortURLRequest  body string true "Long URL to be shortened"
//	@Success      201 {string} string "Short URL id"
//	@Failure      400 {string} string "Invalid long URL"
//	@Failure      500 {string} string "Internal server error"
//	@Router       /private/v1/short-urls/create [post]
func (h *ShortURLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var request ShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	shortURLId, err := h.shortURLManager.CreateShortURL(ctx, request.LongURL)
	if err != nil {
		http.Error(w, "failed to create short URL", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(shortURLId)); err != nil {
		h.logger.Error("failed to write response", logging.ErrorKey, err)

		return
	}
}

// DeleteShortURL godoc
//
//	@Summary      Delete a short URL
//	@Description  Delete a short URL by its id
//	@Tags         short-url, private
//	@Accept       json
//	@Produce      json
//	@Param        shortURLId  path string true "Short URL id to be deleted"
//	@Success      200 {string} string "Short URL deleted successfully"
//	@Failure      400 {string} string "Invalid short URL id"
//	@Failure      500 {string} string "Internal server error"
//	@Router       /private/v1/short-urls/{shortURLId} [delete]
func (h *ShortURLHandler) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	shortURLId := chi.URLParam(r, "shortURLId")
	if shortURLId == "" {
		http.Error(w, "short URL id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.shortURLManager.DeleteShortURL(ctx, shortURLId); err != nil {
		http.Error(w, "failed to delete short URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RedirectToLongURL godoc
//
//	@Summary      Redirect to long URL
//	@Description  Redirect to the long URL for the given short URL id
//	@Tags         short-url, public
//	@Accept       json
//	@Produce      json
//	@Param        longURL  query string true "Long URL to be shortened"
//	@Success      302 {string} string "Short URL ID"
//	@Failure      400 {string} string "Invalid long URL"
//	@Failure      404 {string} string "Short URL not found"
//	@Failure      500 {string} string "Internal server error"
//	@Router       /public/v1/short-urls/{shortURLId} [get]
func (h *ShortURLHandler) RedirectToLongURL(w http.ResponseWriter, r *http.Request) {
	shortURLId := chi.URLParam(r, "shortURLId")
	if shortURLId == "" {
		http.Error(w, "short URL id is required", http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	longURL, err := h.shortURLManager.GetLongURL(ctx, shortURLId)
	if err != nil {
		switch {
		case errors.Is(err, shorturl.ErrShortURLNotFound):
			// TODO: return a custom error page instead of a generic 404
			http.Error(w, "", http.StatusNotFound)

			return
		default:
			http.Error(w, "failed to retrieve long URL", http.StatusInternalServerError)

			return
		}
	}

	h.metricsManager.RecordShortURLRequestAsync(shortURLId, r.RemoteAddr)

	http.Redirect(w, r, longURL, http.StatusFound)
}

// GetShortURLMetrics godoc
//
//	@Summary      Get short URL metrics
//	@Description  Get metrics for a short URL within a specified time range
//	@Tags         short-url, private
//	@Accept       json
//	@Produce      json
//	@Param        shortURLId  path string true "Short URL id to get metrics for"
//	@Param        from        query string true "Start time for metrics (RFC3339 format)"
//	@Param        to          query string true "End time for metrics (RFC3339 format)"
//	@Success      200 {object} metrics.Metrics "Short URL metrics"
//	@Failure      400 {string} string "Invalid request parameters"
//	@Failure      404 {string} string "Metrics not found"
//	@Failure      500 {string} string "Internal server error"
//	@Router       /private/v1/short-urls/{shortURLId}/metrics [get]
func (h *ShortURLHandler) GetShortURLMetrics(w http.ResponseWriter, r *http.Request) {
	shortURLId := chi.URLParam(r, "shortURLId")
	if shortURLId == "" {
		http.Error(w, "short URL id is required", http.StatusBadRequest)

		return
	}

	var request ShortURLMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	metricsResult, err := h.metricsManager.GetShortURLMetrics(ctx, shortURLId, request.From, request.To)
	if err != nil {
		http.Error(w, "failed to retrieve metrics", http.StatusInternalServerError)

		return
	}

	response, err := json.Marshal(metricsResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		h.logger.Error("failed to write response", logging.ErrorKey, err)

		return
	}
}
