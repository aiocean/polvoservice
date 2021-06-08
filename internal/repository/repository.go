package repository

import (
	"context"

	"cloud.google.com/go/firestore"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
)

type Repository interface {
	GetPackage(ctx context.Context, packageOrn string) (*polvo_v1.Package, error)
	ListPackages(ctx context.Context, projectOrn string) (*firestore.DocumentIterator, error)

	ListVersions(ctx context.Context, packageOrn string) (*firestore.DocumentIterator, error)
	GetVersion(ctx context.Context, versionOrn string) (*polvo_v1.Version, error)

	SetResource(ctx context.Context, resourceOrn string, fields interface{}, opts ...firestore.SetOption) error
	DeleteResource(ctx context.Context, resourceOrn string) error
}
