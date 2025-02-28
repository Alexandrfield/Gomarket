package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/handle"
	"github.com/Alexandrfield/Gomarket/internal/market"
	"github.com/Alexandrfield/Gomarket/internal/server"
	"github.com/Alexandrfield/Gomarket/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Printf("Start app test print.")
	log.Printf("Start app test log.")
	logger := common.GetComponentLogger()
	logger.Infof("Start app.")
	config, err := server.ParseFlags()
	if err != nil {
		logger.Fatalf("issue with parse flags")
	}
	storeConfig := storage.Config{DatabasURI: config.DatabaseURI}
	storageServer, err := storage.GetStorage(storeConfig, logger)
	if err != nil {
		logger.Fatalf("cant init storage %s", err)
	}
	logger.Infof("init communicator ... ")
	communicatorExternalService := market.CommunicatorAddServer{Logger: logger,
		Storage: storageServer, AddresMarket: config.AccurakSystemAddres}
	commChan := communicatorExternalService.Init()
	server := handle.ServiceHandler{Logger: logger, Storage: storageServer, BufferOrder: commChan}
	server.Init()
	logger.Infof("init route ... ")
	router := chi.NewRouter()
	router.Post(`/api/user/register`, server.Rgistarte())
	router.Post(`/api/user/login`, server.Login())
	router.Post(`/api/user/orders`, server.Orders())
	router.Get(`/api/user/orders`, server.GetOrders())
	router.Get(`/api/user/balance`, server.GetBalance())

	router.Post(`/api/user/balance/withdraw`, server.Withdraw())
	router.Get(`/api/user/withdrawals`, server.Withdrawals())

	logger.Infof("Server started. config.ServerAddress:%s", config.ServerAddress)
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
