package v1

import (
	"context"
	"os"

	"github.com/google/wire"
	"github.com/pkg/errors"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
	"pkg.aiocean.dev/serviceutil/grpcutils"
)

const (
	ServiceName   = "polvo"
	BaseDomain    = "aiocean.services"
	ServiceDomain = ServiceName + "." + BaseDomain
	port          = "443"
)

var WireSet = wire.NewSet(
	NewClient,
	DefaultConfig,
)

type Config struct {
	Address string
}

func DefaultConfig() *Config {
	return &Config{
		Address: ServiceDomain + ":" + port,
	}
}

func ConfigFromEnv() *Config {
	if address, hasBaseDomain := os.LookupEnv("ADDRESS"); hasBaseDomain {
		return &Config{
			Address: address,
		}
	}

	if baseDomain, hasBaseDomain := os.LookupEnv("BASE_DOMAIN"); hasBaseDomain {
		return &Config{
			Address: ServiceName + "." + baseDomain + ":" + os.Getenv("PORT"),
		}
	}

	return &Config{
		Address: ":" + os.Getenv("PORT"),
	}
}

func NewClient(ctx context.Context, config *Config) (polvo_v1.PolvoServiceClient, error) {

	conn, err := grpcutils.NewGrpcConnection(ctx, config.Address)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to dial")
	}

	client := polvo_v1.NewPolvoServiceClient(conn)
	return client, nil
}
