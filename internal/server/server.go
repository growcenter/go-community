package server

import (
	"context"
	"go-community/internal/config"
	"net/http"
	"strconv"
)

type Server struct {
	server *http.Server
}

func New(config *config.Configuration, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:           ":" + strconv.Itoa(config.Application.Port),
			Handler:        handler,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}