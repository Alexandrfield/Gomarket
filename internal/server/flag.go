package server

import (
	"flag"
	"os"
	"strings"
)

func ParseFlags() (Config, error) {
	config := Config{}
	var tempServerAdderess string
	if tempServerAdderess = os.Getenv("RUN_ADDRESS"); tempServerAdderess != "" {
		config.ServerAddress = tempServerAdderess
	}
	var tempDatabaseURI string
	if tempDatabaseURI = os.Getenv("DATABASE_URI"); tempDatabaseURI != "" {
		config.DatabaseURI = tempDatabaseURI
	}
	var tempAccurakSystemAddres string
	if tempAccurakSystemAddres = os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); tempAccurakSystemAddres != "" {
		config.AccurakSystemAddres = tempAccurakSystemAddres
	}

	flag.StringVar(&config.ServerAddress, "a", tempServerAdderess,
		"address and port to run server ")
	flag.StringVar(&config.DatabaseURI, "d", tempDatabaseURI,
		"uri for database [default:]")
	flag.StringVar(&config.AccurakSystemAddres, "r", tempAccurakSystemAddres,
		"addres for connect accures sysem ")
	flag.Parse()

	// костыль для тестов. Наткнулся на ошибку, когда localhost неправильно резоливлся в сети докера
	res := strings.Split(config.AccurakSystemAddres, ":")
	config.AccurakSystemAddres = "127.0.0.1:" + res[2]
	return config, nil
}
