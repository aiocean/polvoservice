package repository

import (
	"context"

	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
)

type Repository interface {
	GetPackage(ctx context.Context, packageOrn string) (*polvo_v1.Package, error)
	ListPackages(ctx context.Context, projectOrn string) ([]*polvo_v1.Package, error)
	SetPackage(ctx context.Context, projectOrn string, updatedFields interface{}) (*polvo_v1.Package, error)

	ListVersions(ctx context.Context, packageOrn string) ([]*polvo_v1.Version, error)
	SetVersion(ctx context.Context, packageOrn string, updatedFields interface{}) (*polvo_v1.Version, error)
	GetVersion(ctx context.Context, versionOrn string) (*polvo_v1.Version, error)

	GetApplication (ctx context.Context, applicationOrn string) (*polvo_v1.Application, error)
	SetApplication (ctx context.Context, updatedFields interface{}) (*polvo_v1.Application, error)
}

type UnimplementedRepository struct {
}

func (u UnimplementedRepository) GetPackage(ctx context.Context, packageOrn string) (*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) ListPackages(ctx context.Context, projectOrn string) ([]*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) SetPackage(ctx context.Context, projectOrn string, updatedFields interface{}) (*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) ListVersions(ctx context.Context, packageOrn string) ([]*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) SetVersion(ctx context.Context, packageOrn string, updatedFields interface{}) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) GetVersion(ctx context.Context, versionOrn string) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) GetApplication(ctx context.Context, applicationOrn string) (*polvo_v1.Application, error) {
	panic("implement me")
}

func (u UnimplementedRepository) SetApplication(ctx context.Context, updatedFields interface{}) (*polvo_v1.Application, error) {
	panic("implement me")
}
