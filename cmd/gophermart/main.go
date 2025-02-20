package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/Alexandrfield/Gomarket/internal/handle"
	"github.com/Alexandrfield/Gomarket/internal/market"
	"github.com/Alexandrfield/Gomarket/internal/server"
	"github.com/Alexandrfield/Gomarket/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Cant not initializate zap logger.err:%w", err)
	}
	defer func() { _ = zapLogger.Sync() }()
	logger := zapLogger.Sugar()
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Recover. Panic occurred. err: %w", err)
			debug.PrintStack()
		}
	}()

	logger.Infof("Start app.")
	config, err := server.ParseFlags()
	if err != nil {
		logger.Fatalf("issue with parse flags")
	}
	//TODO: Remove this logs
	logger.Debugf("DatabaseURI:%s; ServerAddress:%s; AccurakSystemAddres:%s;",
		config.DatabaseURI, config.ServerAddress, config.AccurakSystemAddres)

	storeConfig := storage.Config{DatabasURI: config.DatabaseURI}
	storageServer, err := storage.GetStorage(storeConfig, logger)
	if err != nil {
		logger.Fatalf("cant init storage %s", err)
	}
	communicatorExternalService := market.CommunicatorAddServer{Logger: logger, Storage: storageServer, AddresMarket: config.AccurakSystemAddres}
	commChan := communicatorExternalService.Init()
	server := handle.ServiceHandler{Storage: storageServer, BufferOrder: commChan}
	server.Init()
	router := chi.NewRouter()
	router.Post(`/api/user/register/`, server.Rgistarte())
	router.Post(`/api/user/login/`, server.Login())
	router.Post(`/api/user/orders/`, server.Orders())
	router.Get(`/api/user/orders/`, server.GetOrders())
	router.Post(`/api/user/register/`, server.Rgistarte())
	router.Get(`/api/user/balance/`, server.GetBalance())

	router.Post(`/api/user/withdraw/`, server.Withdraw())
	router.Get(`/api/user/withdrawals/`, server.Withdrawals())

	logger.Info("Server started")
	go func() {
		err = http.ListenAndServe(config.ServerAddress, router)
		if err != nil {
			logger.Errorf("Unexpected error. err:%s", err)
		}
	}()
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-osSignals

	logger.Infof("Stop app.")
}
