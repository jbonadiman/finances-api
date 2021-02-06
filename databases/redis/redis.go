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

const TimeOut = 1 * time.Second

var singleton *DB

func GetDB() (*DB, error) {
	if singleton == nil {
		singleton = &DB{
			Connection: utils.Connection{
				Host:     environment.LambdaHost,
				Password: environment.LambdaPassword,
				Port:     environment.LambdaPort},
			client:     nil,
		}
		singleton.ConnectionString = formatConnectionString(singleton)

		opt, err := redis.ParseURL(singleton.ConnectionString)
		if err != nil {
			return nil, err
		}

		singleton.client = redis.NewClient(opt)
	}

	return singleton, nil
}

func formatConnectionString(db *DB) string {
	return fmt.Sprintf(
		"redis://%v:%v@%v:%v",
		db.User,
		db.Password,
		db.Host,
		db.Port)
}

func (db *DB) GetTokenFromCache() (*oauth2.Token, error) {
	log.Println("attempting to retrieve token from cache...")
	wg := sync.WaitGroup{}

	var (
		accessToken,
		refreshToken,
		// expiry,
		tokenType string
	)

	log.Println("getting token from cache...")

	wg.Add(3)
	go db.getValue(&wg, "token:AccessToken", &accessToken)
	go db.getValue(&wg, "token:RefreshToken", &refreshToken)
	go db.getValue(&wg, "token:TokenType", &tokenType)

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

func (db *DB) getValue(wg *sync.WaitGroup, key string, variable *string) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	*variable = db.client.Get(ctx, key).Val()
}

func (db *DB) CompareKeys(apiKey string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	return db.client.Get(ctx, "apiKey").Val() == apiKey
}

func (db *DB) SetValue(key string, value interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	db.client.Set(
		ctx,
		key,
		value,
		0,
	)
}
