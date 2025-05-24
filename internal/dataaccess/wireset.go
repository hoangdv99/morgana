package dataaccess

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/dataaccess/database"
)

var WireSet = wire.NewSet(
	database.WireSet,
)
