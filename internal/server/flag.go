package server

import (
	"flag"
	"fmt"
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

	res := strings.Split(config.AccurakSystemAddres, ":")
	fmt.Printf("old:config.AccurakSystemAddres: %s\n", config.AccurakSystemAddres)
	config.AccurakSystemAddres = fmt.Sprintf("127.0.0.1:%s", res[1])
	fmt.Printf("new:config.AccurakSystemAddres: %s\n", config.AccurakSystemAddres)
	return config, nil
}
