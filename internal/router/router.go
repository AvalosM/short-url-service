package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/AvalosM/short-url-service/internal/handlers"
)

func NewRouter(config *Config, shortURLHandler *handlers.ShortURLHandler) http.Handler {
	r := chi.NewRouter()

	// Mount the routers
	r.Mount("/public", createPublicRouter(shortURLHandler))
	r.Mount("/private", createPrivateRouter(shortURLHandler))

	if config.SwaggerEnabled {
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.InstanceName("swagger")))
	}

	return r
}

func createPublicRouter(shortURLHandler *handlers.ShortURLHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: set public middlewares (CORS, Rate Limiting, etc.)
	r.Use(middleware.RealIP)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/short-urls", func(r chi.Router) {
			r.Get("/{shortURLId}", shortURLHandler.RedirectToLongURL)
		})
	})

	return r
}

func createPrivateRouter(shortURLHandler *handlers.ShortURLHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: set private middlewares (Auth)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/short-urls", func(r chi.Router) {
			r.Post("/", shortURLHandler.CreateShortURL)
			r.Delete("/{shortURLId}", shortURLHandler.DeleteShortURL)
			r.Get("/{shortURLId}/metrics", shortURLHandler.GetShortURLMetrics)
		})
	})

	return r
}
