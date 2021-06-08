package main

import (
	"context"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	handler, err := InitializeHandler(ctx)
	if err != nil {
		panic(err)
	}

	handler.Serve()
}
