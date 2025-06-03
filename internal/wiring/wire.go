//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/app"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/dataaccess"
	"github.com/hoangdv99/morgana/internal/handler"
	"github.com/hoangdv99/morgana/internal/logic"
	"github.com/hoangdv99/morgana/internal/utils"
)

var WireSet = wire.NewSet(
	app.WireSet,
	configs.WireSet,
	dataaccess.WireSet,
	logic.WireSet,
	handler.WireSet,
	utils.WireSet,
)

func InitializeStandaloneServer(configFilePath configs.ConfigFilePath) (*app.StandaloneServer, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
