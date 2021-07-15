package repository

import (
	"context"
	"encoding/json"
	"log"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/google/wire"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
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
		d, err := grpc.Dial("165.22.105.129:9080", grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		r.dgraphClient = dgo.NewDgraphClient(
			api.NewDgraphClient(d),
		)
	}

	return r.dgraphClient, nil
}

func (r *DgraphRepository) SetApplication (ctx context.Context, updatedFields interface{}) (*polvo_v1.Application, error) {

	txn := r.dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	payload, err := json.Marshal(updatedFields)
	if err != nil {
		return nil, err
	}

	payload, err = sjson.SetBytes(payload, "dgraph.type", []string{"Application"})
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{
		SetJson: payload,
	}

	res, err := txn.Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}

	var setedApplication polvo_v1.Application

	if err := json.Unmarshal(res.Json, &setedApplication); err != nil {
		return nil, err
	}

	return &setedApplication, nil
}

