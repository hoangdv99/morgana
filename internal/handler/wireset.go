package handler

import (
	"github.com/google/wire"
	"github.com/hoangdv99/morgana/internal/handler/consumers"
	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http"
	"github.com/hoangdv99/morgana/internal/handler/jobs"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
	consumers.WireSet,
	jobs.WireSet,
)
