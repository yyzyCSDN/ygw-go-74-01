package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/service"
)

type Server struct {
	http    *http.Server
	service *service.Service
	webDir  string
}

func NewServer(cfg Config) (*Server, error) {
	svc, err := service.BuildService(service.Config{DataDir: cfg.DataDir})
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	registerHandlers(mux, svc, cfg.WebDir)
	address := cfg.Addr + ":" + strconv.Itoa(cfg.Port)
	return &Server{
		http:    &http.Server{Addr: address, Handler: mux},
		service: svc,
		webDir:  cfg.WebDir,
	}, nil
}

func (s *Server) Start() error {
	err := s.service.Recover()
	if err != nil && !errors.Is(err, differential.ErrSnapshotMissing) {
		return err
	}
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.http.Shutdown(ctx)
	if persistErr := s.service.Persist(); persistErr != nil && err == nil {
		err = persistErr
	}
	return err
}
