package server

import (
	"flag"
	"os"
)

func ParseFlags() (Config, error) {
	config := Config{}
	if envServerAdderess := os.Getenv("RUN_ADDRESS"); envServerAdderess != "" {
		config.ServerAddress = envServerAdderess
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		config.DatabaseURI = envDatabaseURI
	}
	if envAccurakSystemAddres := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccurakSystemAddres != "" {
		config.AccurakSystemAddres = envAccurakSystemAddres
	}
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080",
		"address and port to run server [default:localhost:8080]") // RUN_ADDRESS
	flag.StringVar(&config.DatabaseURI, "d", "",
		"uri for database [default:]") // RUN_ADDRESS
	flag.StringVar(&config.AccurakSystemAddres, "r", "localhost:8091",
		"addres for connect accures sysem [default:localhost:8091]") // RUN_ADDRESS
	flag.Parse()
	// config.DatabaseURI = "host=localhost port=5430 user=gopher password=qwerty dbname=postgres_db sslmode=disable"
	return config, nil
}
