package datastore

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"url-checker/datamodel"
)

const (
	NoExpiration = 0
)

type DB interface {
	GetMalware(string) *datamodel.Malware
}

type Store struct {
	collection *mongo.Collection
	cache      *redis.Client
}

type UrlMeta struct {
	Risk     string
	Category string
}

func NewMongo(dbClient *mongo.Client, cacheClient *redis.Client) DB {
	collection := dbClient.Database("malware").Collection("malware")
	return Store{
		collection: collection,
		cache:      cacheClient,
	}
}

func (m Store) GetMalware(url string) *datamodel.Malware {
	var metaObj = &UrlMeta{}
	var err error

	// Check if the URL exists in the cache and fetch if it does
	err = m.getCache(url, metaObj)
	if err == nil {
		log.Println("Returning Cached Results.")
		return &datamodel.Malware{
			Url:      url,
			Risk:     metaObj.Risk,
			Category: metaObj.Category,
		}
	}

	// Try to fetch URL from the database
	var query = bson.D{{"url", url}}
	var malware *datamodel.Malware
	err = m.collection.FindOne(context.TODO(), query).Decode(&malware)
	if err != nil {
		log.Println("No URL found in DB: ", err.Error())
		malware = &datamodel.Malware{
			Url:      url,
			Risk:     "Unknown",
			Category: "Unknown",
		}
	}

	// Caching the results
	metaObj.Risk = malware.Risk
	metaObj.Category = malware.Category
	err = m.setCache(malware.Url, metaObj)
	if err != nil {
		log.Println("Cache Write Error: ", err.Error())
	}
	return malware
}

func (m Store) setCache(key string, value interface{}) error {
	p, err := json.Marshal(value)
	if err != nil {
		log.Println("Failed to serialize meta data.")
		return err
	}
	return m.cache.Set(key, p, NoExpiration).Err()
}

func (m Store) getCache(key string, dest interface{}) error {
	p, err := m.cache.Get(key).Bytes()
	if err != nil {
		log.Println("No URL found in Cache: ", err.Error())
		return err
	}
	return json.Unmarshal(p, dest)
}
