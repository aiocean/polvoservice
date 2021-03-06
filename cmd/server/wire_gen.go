// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

import (
	"context"
	"pkg.aiocean.dev/polvoservice/internal/repository"
	"pkg.aiocean.dev/polvoservice/internal/server"
	"pkg.aiocean.dev/serviceutil/handler"
	"pkg.aiocean.dev/serviceutil/healthserver"
	"pkg.aiocean.dev/serviceutil/interceptor"
	"pkg.aiocean.dev/serviceutil/logger"
)

// Injectors from wire.go:

func InitializeHandler(ctx context.Context) (*handler.Handler, error) {
	zapLogger, err := logger.NewLogger(ctx)
	if err != nil {
		return nil, err
	}
	dgraphRepository, err := repository.NewDgraphRepository()
	if err != nil {
		return nil, err
	}
	serverServer := server.NewServer(zapLogger, dgraphRepository)
	streamServerInterceptor := interceptor.NewStreamServerInterceptor(zapLogger)
	unaryServerInterceptor := interceptor.NewUnaryServerInterceptor(zapLogger)
	healthserverServer := healthserver.NewHealthServer()
	handlerHandler := handler.NewHandler(ctx, zapLogger, serverServer, streamServerInterceptor, unaryServerInterceptor, healthserverServer)
	return handlerHandler, nil
}
