package wire

import (
	pingController "github.com/dev-au/CodeStream/internal/adapter/http/ping"
	roomController "github.com/dev-au/CodeStream/internal/adapter/http/room"
	"github.com/dev-au/CodeStream/internal/config"
	httpServer "github.com/dev-au/CodeStream/internal/infrastructure/http"
	"github.com/google/wire"
)

var ConfigSet = wire.NewSet(
	ProvideAppConfig,
	ProvideHTTPConfig,
)

func ProvideAppConfig(cfg *config.Config) *config.AppConfig   { return &cfg.Application }
func ProvideHTTPConfig(cfg *config.Config) *config.HTTPConfig { return &cfg.HTTP }

var DatabaseSet = wire.NewSet()

var RepositorySet = wire.NewSet()

var UseCaseSet = wire.NewSet()

var ControllerSet = wire.NewSet(
	pingController.NewController,
	roomController.NewController,
)

var ServerSet = wire.NewSet(
	httpServer.NewServer,
)

var AllProviders = wire.NewSet(
	DatabaseSet,
	ConfigSet,
	RepositorySet,
	UseCaseSet,
	ControllerSet,
	ServerSet,
)
