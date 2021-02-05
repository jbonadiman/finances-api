package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finances-api/environment"
	"github.com/jbonadiman/finances-api/utils"
)

type DB struct {
	utils.Connection
	client *redis.Client
}

var redisSingleton *DB

func (db *DB) GetClient() (*redis.Client, error) {
	connectionStr := db.GetConnectionString()

	opt, err := redis.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opt), nil
}

func (db *DB) GetConnectionString() string {
	return fmt.Sprintf(
		"redis://%v:%v@%v:%v",
		db.User,
		db.Password,
		db.Host,
		db.Port)
}

func GetDB() *DB {
	

	if redisSingleton == nil {
		redisSingleton = &DB{
			Connection: utils.Connection{
				Host:     environment.LambdaHost,
				Password: environment.LambdaPassword,
				Port:     environment.LambdaPort},
			client:     nil,
		}

		redisSingleton.ConnectionString = redisSingleton.GetConnectionString()

		connectionStr := db.GetConnectionString()

	}

	return redisSingleton
}

func GetTokenFromCache() (*oauth2.Token, error) {
	log.Println("attempting to retrieve token from cache...")

	redisClient, err := GetDB().GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	wg := sync.WaitGroup{}

	var (
		accessToken,
		refreshToken,
		// expiry,
		tokenType string
	)

	log.Println("getting token from cache...")

	wg.Add(3)
	go func() {
		accessToken = redisClient.Get(
			ctx,
			"token:AccessToken",
		).Val()

		wg.Done()
	}()

	go func() {
		refreshToken = redisClient.Get(
			ctx,
			"token:RefreshToken",
		).Val()

		wg.Done()
	}()

	go func() {
		tokenType = redisClient.Get(
			ctx,
			"token:TokenType",
		).Val()

		wg.Done()
	}()

	// go func() {
	// 	expiry = redisClient.Get(
	// 		ctx,
	// 		"token:Expiry",
	// 	).Val()
	//
	// 	wg.Done()
	// }()

	wg.Wait()

	if accessToken != "" && tokenType != "" && refreshToken != "" {
		log.Println("retrieved token from cache successfully")

		token := oauth2.Token{
			AccessToken:  accessToken,
			TokenType:    tokenType,
			RefreshToken: refreshToken,
			Expiry:       time.Time{},
		}

		return &token, nil
	}

	return nil, nil
}

func getValue(wg *sync.WaitGroup, key string) {
	tokenType = redisClient.Get(
		ctx,
		"token:TokenType",
	).Val()

	wg.Done()
}
