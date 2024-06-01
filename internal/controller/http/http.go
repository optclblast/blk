package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/optclblast/blk/internal/logger"
)

type router struct {
	*chi.Mux
	log *slog.Logger

	walletsController WalletsController
}

func NewRouter(
	log *slog.Logger,
	walletsController WalletsController,
) http.Handler {
	r := &router{
		Mux:               chi.NewRouter(),
		log:               log,
		walletsController: walletsController,
	}

	r.Use(middleware.Recoverer)
	r.Use(handleMw)

	r.Get("/most-changed", r.handle(r.walletsController.MostChangedWalletAddress, "most-changed"))

	return r
}

// handleMw adds content type headers
func handleMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

// handle is a helper functions that makes it easier to work woth http handlers
func (s *router) handle(
	h func(w http.ResponseWriter, req *http.Request) (any, error),
	method_name string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := h(w, r)
		if err != nil {
			s.log.Error(
				"http error",
				slog.String("method_name", method_name),
				logger.Err(err),
			)

			responseError(w, err)

			return
		}

		out, err := json.Marshal(resp)
		if err != nil {
			s.log.Error(
				"error marshal response",
				slog.String("method_name", method_name),
				logger.Err(err),
				slog.Any("object", resp),
			)

			responseError(w, err)

			return
		}

		w.WriteHeader(http.StatusOK)

		if _, err = w.Write(out); err != nil {
			s.log.Error(
				"error write http response",
				slog.String("method_name", method_name),
				logger.Err(err),
			)
		}
	}
}

func responseError(w http.ResponseWriter, e error) {
	apiErr := mapError(e)

	out, err := json.Marshal(apiErr)
	if err != nil {
		return
	}

	w.WriteHeader(apiErr.Code)
	w.Write(out)
}
