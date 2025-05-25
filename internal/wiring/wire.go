//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/dataaccess"
	"github.com/hoangdv99/morgana/internal/handler"
	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/logic"
	"github.com/hoangdv99/morgana/internal/utils"
)

var WireSet = wire.NewSet(
	configs.WireSet,
	dataaccess.WireSet,
	logic.WireSet,
	handler.WireSet,
	utils.WireSet,
)

func InitializeGRPCServer(configFilePath configs.ConfigFilePath) (grpc.Server, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
