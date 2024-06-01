package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/optclblast/blk/internal/logger"
)

// router object
type router struct {
	*chi.Mux
	log *slog.Logger

	walletsController WalletsController
}

// NewRouter returns a new http.Handler object that can power your server
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

	r.Get("/most-changed", r.handle(
		r.walletsController.MostChangedWalletAddress,
		"most-changed",
	))

	return r
}

// handleMw adds content type headers
func handleMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

type handleFunc func(w http.ResponseWriter, req *http.Request) (any, error)

// handle is a helper functions that makes it easier to work with http handlers
func (s *router) handle(
	h handleFunc,
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

			s.responseError(w, err)

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

			s.responseError(w, err)

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

func (s *router) responseError(
	w http.ResponseWriter,
	e error,
) {
	apiErr := mapError(e)

	out, err := json.Marshal(apiErr)
	if err != nil {
		return
	}

	w.WriteHeader(apiErr.Code)

	if _, err := w.Write(out); err != nil {
		s.log.Error("error write error to connection", logger.Err(err))
	}
}
