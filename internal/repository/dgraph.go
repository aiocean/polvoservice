package repository

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
)

var DgraphWireSet = wire.NewSet(
	NewDgraphRepository,
	wire.Bind(new(Repository), new(*DgraphRepository)),
)

func NewDgraphRepository () (*DgraphRepository, error) {
	return &DgraphRepository{}, nil
}

type DgraphRepository struct {
	dgraphClient *dgo.Dgraph
	UnimplementedRepository
}

func (r *DgraphRepository) getDgraphClient () (*dgo.Dgraph, error) {
	if r.dgraphClient == nil {
		d, err := grpc.Dial(os.Getenv("DGRAPH_ADDRESS"), grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		r.dgraphClient = dgo.NewDgraphClient(
			api.NewDgraphClient(d),
		)
	}

	return r.dgraphClient, nil
}

func (r *DgraphRepository) GetPackage(ctx context.Context, name string) (*polvo_v1.Package, error) {

	query := `{
		  items(func: eq(dgraph.type, "Package")) @filter(eq(name, "` + name + `")){
			uid
			name
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query:      query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, err
	}

	items := gjson.GetBytes(requestResult.Json, "items.0")
	if !items.Exists() {
		return nil, errors.New("package " + name + " not found")
	}

	pkg := &polvo_v1.Package{
		Name:       items.Get("name").String(),
	}

	return pkg, nil
}

func (r *DgraphRepository) GetVersion(ctx context.Context, packageName, versionName string) (*polvo_v1.Version, error) {

	query := `{
		  package(func: eq(name, "` + packageName + `")) @filter(eq(dgraph.type, "Package")){
			versions @filter(eq(name, "` + versionName + `")) {
				uid
				name
				manifest_url
			}
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query:      query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(requestResult.Json))
	items := gjson.GetBytes(requestResult.Json, "package.0.versions.0")
	if !items.Exists() {
		return nil, errors.New("version not found")
	}

	pkg := &polvo_v1.Version{
		Name:       items.Get("name").String(),
		ManifestUrl: items.Get("manifest_url").String(),
	}

	return pkg, nil
}

func (r *DgraphRepository) GetHeaviestVersion(ctx context.Context, packageName string) (*polvo_v1.Version, error) {

	query := `{
		  package(func: eq(name, "` + packageName + `")) @filter(eq(dgraph.type, "Package")){
			versions @facets(orderdesc: weight) (first: 1) {
				uid
				name
				manifest_url
			}
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query:      query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, err
	}

	items := gjson.GetBytes(requestResult.Json, "package.0.versions.0")
	if !items.Exists() {
		return nil, errors.New("version not found")
	}

	pkg := &polvo_v1.Version{
		Name:       items.Get("name").String(),
		ManifestUrl: items.Get("manifest_url").String(),
	}

	return pkg, nil
}

func (r *DgraphRepository) CreatePackage(ctx context.Context, pkg *polvo_v1.Package) (*polvo_v1.Package, error) {
	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get db client", err)
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	setStatment := `
_:package <dgraph.type> "Package" .
_:package <name> "` + pkg.GetName() + `" .
_:package <maintainer> "` + pkg.GetMaintainer() + `" .
_:package <created_at> "` + time.Now().Format(time.RFC3339) + `" .`

	request := &api.Request{
		Query:      `query {pkg as var(func: eq(name, "` + pkg.GetName() + `")) @filter(eq(dgraph.type, "Package"))}`,
		Mutations:  []*api.Mutation{
			{
				SetNquads: []byte(setStatment),
				Cond:       "@if(eq(len(pkg), 0))",
			},
		},
	}

	mutateResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mutate data", err)
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit data", err)
	}

	if len(mutateResult.Uids) == 0 {
		return nil, status.Errorf(codes.Aborted, "application was not created, maybe it was exists")
	}

	return pkg, nil
}

func (r *DgraphRepository) UpdateVersion(ctx context.Context, packageName, versionName string, updatedFields map[string]interface{}) (*polvo_v1.Version, error) {
	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get db client", err)
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	updateNquads := ``

	if manifestUrl, ok := updatedFields["ManifestUrl"]; ok {
		updateNquads += `uid(versionUid) <manifest_url> "` + manifestUrl.(string) + `" .` + "\n"
	}

	definePackageUid := ``
	if weight, ok := updatedFields["Weight"]; ok {
		definePackageUid = "packageUid as uid\n"
		updateNquads += `uid(packageUid) <versions> uid(versionUid) (weight=` + strconv.FormatInt(int64(weight.(uint32)), 10) + `) .` + "\n"
	}

	if versionName, ok := updatedFields["Name"]; ok {
		updateNquads += `uid(versionUid) <name> "` + versionName.(string) + `" .` + "\n"
	}

	request := &api.Request{
		Query:      `{
						var(func: eq(name, ` + packageName + `)) @filter(eq(dgraph.type, "Package")) {`+
						definePackageUid +
						`versions @filter(eq(name, `+ versionName +`)) {
								versionUid as uid
							}
						}
					}`,
		Mutations:  []*api.Mutation{
			{
				SetNquads: []byte(updateNquads),
				Cond:       "@if(eq(len(versionUid), 1))",
			},
		},
	}

	if _, err := txn.Do(ctx, request); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mutate data", err)
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit data", err)
	}

	savedVersion, err := r.GetVersion(ctx, packageName, versionName)
	if err != nil {
		return nil, err
	}

	return savedVersion, nil
}

func (r *DgraphRepository) CreateVersion(ctx context.Context, packageName string, version *polvo_v1.Version) (*polvo_v1.Version, error) {
	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get db client", err)
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	request := &api.Request{
		Query:      `{
						var(func: eq(dgraph.type, "Package")) @filter(eq(name, ` + packageName + `)) {
							packageUid as uid
							versions @filter(eq(name, `+ version.GetName() +`)) {
								versionUid as uid
							}
						}
					}`,
			Mutations:  []*api.Mutation{
			{
				SetNquads: []byte(` _:version <dgraph.type> "Version" .
									_:version <name> "` + version.GetName() + `" .
									_:version <manifest_url> "` + version.GetManifestUrl() + `" .
									_:version <created_at> "` + time.Now().Format(time.RFC3339) + `" .
									uid(packageUid) <versions> _:version (weight=0) .`,
							),
				Cond:       "@if(eq(len(versionUid), 0) AND eq(len(packageUid), 1))",
			},
		},
	}

	mutateResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mutate data", err)
	}

	if len(mutateResult.Uids) == 0 {
		return nil, status.Errorf(codes.Aborted, "version was not created, maybe it was exists")
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit data", err)
	}

	return version, nil
}


func (r *DgraphRepository) DeleteVersion(ctx context.Context, packageName, versionName string) error {
	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return err
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	request := &api.Request{
		Query: `
{
	 var(func: eq(dgraph.type, "Package")) @filter(eq(name,"`+ packageName +`")){
		packageUid as uid
	  versions @filter(eq(name, "` + versionName + `")) {
				versionUid as uid
	  }
	}
 }`,
		Mutations:  []*api.Mutation{
			{
				DelNquads: []byte(`uid(versionUid) * * .
								   uid(packageUid) <versions> uid(versionUid) .`),
			},
		},
	}

	if _, err := txn.Do(ctx, request); err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	if err := txn.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	return nil
}

func (r *DgraphRepository)DeletePackage(ctx context.Context, name string) error {
	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return err
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	request := &api.Request{
		Query: `{
					var(func: eq(dgraph.type, "Package")) @filter(eq(name, ` + name + `)) {
						packageUid as uid
						versions {
							versionUid as uid
						}
					}
				}`,
		Mutations:  []*api.Mutation{
			{
				DelNquads: []byte(`uid(packageUid) * * .
									uid(versionUid) * * .`),
			},
		},
	}

	if _, err := txn.Do(ctx, request); err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	if err := txn.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	return nil
}

func (r *DgraphRepository) ListPackages(ctx context.Context) ([]*polvo_v1.Package, error)  {

	query := `{
		  items(func: eq(dgraph.type, "Package")){
			uid
			name
			maintainer
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query:      query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, err
	}

	var packages []*polvo_v1.Package

	rawPackages := gjson.GetBytes(requestResult.Json, "items")
	if !rawPackages.Exists() {
		return nil, errors.New("no item found")
	}
	rawPackages.ForEach(func(key, value gjson.Result) bool {
		pkg := &polvo_v1.Package{
			Name: value.Get("name").String(),
			Maintainer: value.Get("maintainer").String(),
		}
		packages = append(packages, pkg)

		return true
	})

	return packages, nil
}

func (r *DgraphRepository) IsPackageExists(ctx context.Context, name string) (bool, error) {
	query := `{
		  packages(func: eq(name, ` + name + `)) @filter(eq(dgraph.type, "Package")){
			uid
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return false, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query: query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return false, err
	}

	packageCount := gjson.GetBytes(requestResult.Json, "packages.#")

	return packageCount.Int() > 0, nil
}

func (r *DgraphRepository) IsVersionExists(ctx context.Context, packageName string, versionName string) (bool, error) {
	query := `{
		  packages(func: eq(name, "` + packageName + `")) @filter(eq(dgraph.type, "Package")){
			versions @filter(eq(name, "` + versionName + `")) {
				uid
			}
		  }
		}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return false, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query: query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return false, err
	}

	versionUid := gjson.GetBytes(requestResult.Json, "packages.0.versions.0.uid")

	return versionUid.Exists(), nil
}

func (r *DgraphRepository) ListVersions(ctx context.Context, packageName string, option ...ListVersionsOptions) ([]*polvo_v1.Version, error) {

	query := `
{
  package(func: eq(dgraph.type, "Package")) @filter(eq(name,"` + packageName + `")) {
	uid
	versions @facets(orderdesc: weight, weight: weight) {
		uid
		name
		manifest_url
	}
  }
}`

	dgraphClient, err := r.getDgraphClient()
	if err != nil {
		return nil, err
	}

	txn := dgraphClient.NewReadOnlyTxn()

	request := &api.Request{
		Query:      query,
	}

	requestResult, err := txn.Do(ctx, request)
	if err != nil {
		return nil, err
	}

	var versions []*polvo_v1.Version

	rawVersions := gjson.GetBytes(requestResult.Json, "package.0.versions")
	if !rawVersions.Exists() {
		return nil, errors.New("not found")
	}

	rawVersions.ForEach(func(key, value gjson.Result) bool {
		version := &polvo_v1.Version{
			Name:        value.Get("name").String(),
			ManifestUrl: value.Get("manifest_url").String(),
			Weight: uint32(value.Get("weight").Uint()),
		}
		versions = append(versions, version)

		return true
	})

	return versions, nil
}
