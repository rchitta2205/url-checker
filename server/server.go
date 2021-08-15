package main

import (
	"context"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"url-checker/application"
	"url-checker/datastore"
)

func main() {
	dbClient, err := mongo.Connect(context.TODO(), dbOptions())
	if err != nil {
		log.Fatal(err)
	}
	defer dbClient.Disconnect(context.TODO())
	cacheClient := redis.NewClient(cacheOptions())
	defer cacheClient.Close()

	store := datastore.NewMongo(dbClient, cacheClient)
	app := application.NewApp(store)
	err = app.Serve()
	log.Println("Error", err)
}

func dbOptions() *options.ClientOptions {
	var host = "db"
	var port = "27017"
	return options.Client().ApplyURI(
		"mongodb://" + host + ":" + port,
	)
}

func cacheOptions() *redis.Options {
	var host = "cache"
	var port = "6379"
	return &redis.Options{
		Addr: host + ":" + port,
	}
}
