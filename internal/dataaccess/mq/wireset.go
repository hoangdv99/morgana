package mq

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq/consumer"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq/producer"
)

var WireSet = wire.NewSet(
	producer.WireSet,
	consumer.WireSet,
)
