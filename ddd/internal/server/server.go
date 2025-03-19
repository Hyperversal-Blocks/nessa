package server

import "github.com/hblocks/nessa/internal/config"

type Server interface {
	Start() error
}

type server struct {
}

func (s *server) Start() error {
	return nil
}

func NewServer(conf *config.ServerConfig) Server {
	return &server{}
}
