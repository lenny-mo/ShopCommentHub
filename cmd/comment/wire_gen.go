// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"comment/internal/biz"
	"comment/internal/conf"
	"comment/internal/data"
	"comment/internal/server"
	"comment/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	db := data.NewDB(confData)
	client := data.NewRedis(confData)
	database := data.NewMongo(confData)
	writer := data.NewKafka(confData)
	esWithIndex := data.NewEsWithIndex(confData)
	dataData, cleanup, err := data.NewData(db, client, database, writer, esWithIndex, logger)
	if err != nil {
		return nil, nil, err
	}
	customerRepo := data.NewCustomerRepo(dataData, logger)
	customerUsecase := biz.NewCustomerUsecase(customerRepo, logger)
	merchantRepo := data.NewMerchantRepo(dataData, logger)
	merchantUsecase := biz.NewMerchantUsecase(merchantRepo, logger)
	eventBusRepo := data.NewEventBusRepo(dataData, logger)
	busUseCase := biz.NewBusUseCase(eventBusRepo, logger)
	readModelRepo := data.NewReadModelRepo(dataData, logger)
	readModelUsecase := biz.NewReadModelUsecase(readModelRepo, logger)
	commentServiceService := service.NewCommentServiceService(customerUsecase, merchantUsecase, busUseCase, readModelUsecase)
	grpcServer := server.NewGRPCServer(confServer, commentServiceService, logger)
	httpServer := server.NewHTTPServer(confServer, commentServiceService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
