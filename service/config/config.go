package config

import (
	"log"
	"os"
	"time"

	"personal/gosketch/api"
	"personal/gosketch/pixservice"
	"personal/gosketch/pixservice/cache"
	"personal/gosketch/service"
	"personal/gosketch/store/postgresdb"

	"github.com/joho/godotenv"
)

const (
	imageType    = "all" // Accepted image-type values: "all", "photo", "illustration", "vector"
	cacheTimeout = 3 * time.Hour
)

func Run() (*postgresdb.PsqlDatastore, *api.Server, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
		return nil, nil, err
	}

	newDB, err := postgresdb.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf(err.Error())
		return nil, nil, err
	}

	service := service.NewService(newDB, cache.New(cacheTimeout, pixservice.New(os.Getenv("PIXABAY_KEY"), imageType)))
	srv := api.NewServer(service)

	return newDB, srv, err
}
