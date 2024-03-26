package app

import (
	"github.com/pluhe7/gophermart/internal/model"
)

type App struct {
	Server  *Server
	Session model.Session
}

func NewApp(server *Server) *App {
	return &App{
		Server: server,
	}
}
