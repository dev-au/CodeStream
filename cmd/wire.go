//go:build wireinject
// +build wireinject

package main

import (
	wirePkg "github.com/google/wire"

	"github.com/dev-au/CodeStream/internal/config"
	"github.com/dev-au/CodeStream/internal/infrastructure/http"
	"github.com/dev-au/CodeStream/internal/infrastructure/wire"
)

type App struct {
	Server *http.Server
}

func NewApp(server *http.Server) *App {
	return &App{
		Server: server,
	}
}

func InitializeApp(cfg *config.Config) (*App, func(), error) {
	wirePkg.Build(
		wire.AllProviders,
		NewApp,
	)
	return &App{}, nil, nil
}
