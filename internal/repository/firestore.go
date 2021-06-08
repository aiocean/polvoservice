package repository

import (
	"context"
	"path/filepath"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
)

var firestoreClient *firestore.Client

func getFirestoreClient(ctx context.Context) (_ *firestore.Client, err error) {
	if firestoreClient == nil {
		firestoreClient, err = firestore.NewClient(ctx, "aio-shopify-services")
		if err != nil {
			return nil, err
		}
	}
	return firestoreClient, nil
}

var FirestoreWireSet = wire.NewSet(
	NewFirestoreRepository,
	wire.Bind(new(Repository), new(*FirestoreRepository)),
)

type FirestoreRepository struct{}

func (f *FirestoreRepository) SetResource(ctx context.Context, resourceOrn string, fields interface{}, opts ...firestore.SetOption) error {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get firestore client")
	}

	relatedResourceName := filepath.Join(strings.Split(resourceOrn, "/")[1:]...)

	if _, err := client.Doc(relatedResourceName).Set(ctx, fields, opts...); err != nil {
		return errors.Wrap(err, "failed to create version")
	}

	return nil
}

func (f *FirestoreRepository) DeleteResource(ctx context.Context, resourceOrn string) error {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(resourceOrn, "/")[1:]...)

	doc := client.Doc(orrn)
	if _, err := doc.Delete(ctx); err != nil {
		return errors.Wrap(err, "failed to delete doc")
	}

	return nil
}
func (f *FirestoreRepository) GetVersion(ctx context.Context, versionOrn string) (*polvo_v1.Version, error) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(versionOrn, "/")[1:]...)

	snapshot, err := client.Doc(orrn).Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, status.Error(codes.NotFound, "package not found")
	}

	var foundVersion polvo_v1.Version

	if err := snapshot.DataTo(&foundVersion); err != nil {
		return nil, err
	}

	foundVersion.Orn = versionOrn

	return &foundVersion, nil
}

func (f *FirestoreRepository) GetPackage(ctx context.Context, packageOrn string) (*polvo_v1.Package, error) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(packageOrn, "/")[1:]...)

	snapshot, err := client.Doc(orrn).Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, status.Error(codes.NotFound, "package not found")
	}

	var foundPackage polvo_v1.Package

	if err := snapshot.DataTo(&foundPackage); err != nil {
		return nil, err
	}

	foundPackage.Orn = packageOrn

	return &foundPackage, nil
}

func (f *FirestoreRepository) ListPackages(ctx context.Context, projectOrn string) (*firestore.DocumentIterator, error) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(projectOrn, "/")[1:]...) + "/packages"

	cols := client.Collection(orrn)
	if cols == nil {
		return nil, status.Error(codes.NotFound, "can not find packages")
	}

	packagesRef := cols.Documents(ctx)
	return packagesRef, nil
}

func (f *FirestoreRepository) ListVersions(ctx context.Context, packageOrn string) (*firestore.DocumentIterator, error) {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "'failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(packageOrn, "/")[1:]...) + "/versions"

	cols := client.Collection(orrn)
	if cols == nil {
		return nil, status.Error(codes.NotFound, "can not find packages")
	}

	colRef := cols.Documents(ctx)
	return colRef, nil
}

func (f *FirestoreRepository) DeletePackage(ctx context.Context, packageOrn string) error {
	client, err := getFirestoreClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get firestore client")
	}

	orrn := filepath.Join(strings.Split(packageOrn, "/")[1:]...)

	doc := client.Doc(orrn)
	if _, err := doc.Delete(ctx); err != nil {
		return errors.Wrap(err, "failed to delete doc")
	}

	return nil
}

func NewFirestoreRepository() (*FirestoreRepository, error) {
	return &FirestoreRepository{}, nil
}
