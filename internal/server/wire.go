//+build wireinject

package server

import (
	"context"

	"github.com/google/wire"
	"pkg.aiocean.dev/polvoservice/internal/repository"
	"pkg.aiocean.dev/serviceutil/wireset"
)

func initializeServer(ctx context.Context) (*Server, error) {
	wire.Build(
		repository.FirestoreWireSet,
		wireset.Default,
		WireSet,
	)
	return nil, nil
}
