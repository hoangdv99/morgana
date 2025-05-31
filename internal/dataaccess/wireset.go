package dataaccess

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/dataaccess/cache"
	"github.com/hoangdv99/morgana/internal/dataaccess/database"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq"
)

var WireSet = wire.NewSet(
	database.WireSet,
	cache.WireSet,
	mq.WireSet,
)
