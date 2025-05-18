package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB
var redisClientInstance *redis.Client

// GetDB retrieves the database connection instance
func GetDB() *gorm.DB {
	return dbInstance
}

// GetRedisClient retrieves the redis connection instance
func GetRedisClient() *redis.Client {
	return redisClientInstance
}

// ConnectDB connects to the database
func ConnectDB() {
	con := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))
	db, err := gorm.Open(mysql.Open(con), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	fmt.Println("Connected to database")
	dbInstance = db
}

// connect redis
func ConnectRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:        os.Getenv("REDIS_ADDR"),
		Password:    os.Getenv("REDIS_PASSWORD"),
		DialTimeout: time.Second * 20,
		DB:          0,
		Username:    os.Getenv("REDIS_USER"),
	})
	//fmt.Println("Connecting to redis", os.Getenv("REDIS_USER"))
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Println("error connecting to redis instance, app will not start", err)
		panic("failed to connect redis")
	}
	//fmt.Println("Connected to redis")
	redisClientInstance = client
}
