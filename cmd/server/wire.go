//+build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"pkg.aiocean.dev/polvoservice/internal/repository"
	"pkg.aiocean.dev/polvoservice/internal/server"
	"pkg.aiocean.dev/serviceutil/handler"
	"pkg.aiocean.dev/serviceutil/wireset"
)

func InitializeHandler(ctx context.Context) (*handler.Handler, error) {
	wire.Build(
		repository.FirestoreWireSet,
		wireset.Default,
		server.WireSet,
	)

	return nil, nil
}
