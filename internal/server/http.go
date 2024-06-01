// server package contains core network layer components
package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

const (
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

// Http Server object
type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

// Notify return an error channel, that will contain error if server died
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Graceful shutdown
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// New returns a new Server object
func New(
	handler http.Handler,
	addr string,
	opts ...Option,
) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		Addr:         addr,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: defaultShutdownTimeout,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	s.start()

	return s
}

// Options configures Server
type Option func(s *Server)

// Port sets a specific port for Server to listen to
func Port(port string) Option {
	return func(s *Server) {
		s.server.Addr = net.JoinHostPort("", port)
	}
}

// ReadTimeout sets a specific connection read timeout
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WriteTimeout sets a specific connection write timeout
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// ShutdownTimeout sets a specific connection shutdown timeout
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
