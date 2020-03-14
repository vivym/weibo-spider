package db

import (
	"github.com/Kamva/mgm/v2"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupDB(config Config) error {
	return mgm.SetDefaultConfig(nil, config.DBName, options.Client().ApplyURI(config.URI))
}
