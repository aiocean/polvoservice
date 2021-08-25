package repository

import (
	"context"

	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
)

type ListVersionsOptions struct {
	Limit *uint
	OrderByWeight *bool
}

type Repository interface {
	GetPackage(ctx context.Context, name string) (*polvo_v1.Package, error)
	ListPackages(ctx context.Context) ([]*polvo_v1.Package, error)
	CreatePackage(ctx context.Context, pkg *polvo_v1.Package) (*polvo_v1.Package, error)
	UpdatePackage(ctx context.Context, name string, updatedFields map[string]interface{}) (*polvo_v1.Package, error)
	DeletePackage(ctx context.Context, name string) error
	IsPackageExists(ctx context.Context, name string) (bool, error)

	ListVersions(ctx context.Context, pkdUid string, option ...ListVersionsOptions) ([]*polvo_v1.Version, error)
	GetVersion(ctx context.Context, packageName, versionName string) (*polvo_v1.Version, error)
	GetHeaviestVersion(ctx context.Context, packageName string) (*polvo_v1.Version, error)
	CreateVersion(ctx context.Context, packageName string, version *polvo_v1.Version) (*polvo_v1.Version, error)
	UpdateVersion(ctx context.Context, packageName, versionName string, updatedFields map[string]interface{}) (*polvo_v1.Version, error)
	DeleteVersion(ctx context.Context, packageName, versionName string) error
	IsVersionExists(ctx context.Context, packageName string, versionName string) (bool, error)
}

type UnimplementedRepository struct {
}

func (u UnimplementedRepository) GetPackage(ctx context.Context, name string) (*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) ListPackages(ctx context.Context) ([]*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) CreatePackage(ctx context.Context, pkg *polvo_v1.Package) (*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) UpdatePackage(ctx context.Context, name string, updatedFields map[string]interface{}) (*polvo_v1.Package, error) {
	panic("implement me")
}

func (u UnimplementedRepository) DeletePackage(ctx context.Context, name string) error {
	panic("implement me")
}

func (u UnimplementedRepository) IsPackageExists(ctx context.Context, name string) (bool, error) {
	panic("implement me")
}

func (u UnimplementedRepository) ListVersions(ctx context.Context, pkdUid string, option ...ListVersionsOptions) ([]*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) GetVersion(ctx context.Context, packageName, versionName string) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) GetHeaviestVersion(ctx context.Context, packageName string) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) CreateVersion(ctx context.Context, packageName string, version *polvo_v1.Version) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) UpdateVersion(ctx context.Context, packageName, versionName string, updatedFields map[string]interface{}) (*polvo_v1.Version, error) {
	panic("implement me")
}

func (u UnimplementedRepository) DeleteVersion(ctx context.Context, packageName, versionName string) error {
	panic("implement me")
}

func (u UnimplementedRepository) IsVersionExists(ctx context.Context, packageName string, versionName string) (bool, error) {
	panic("implement me")
}

