package ws

import (
	"context"
	_ "context"
	"encoding/json"
	"fmt"
	_ "fmt"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct {
	MongoDBRepository
	RedisRepository
}

type MongoDBRepository struct {
	Collection *mongo.Database
}

func NewMongoDbRepository(client *mongo.Database) *mongo.Database {
	return client

}

func NewMessageRepository(client *mongo.Database) MessageRepository {

	return MessageRepository{
		MongoDBRepository: MongoDBRepository{NewMongoDbRepository(client)},
		RedisRepository:   NewRedisRepository(),
	}

}

func (r MongoDBRepository) InsertMessage(message Message) *mongo.InsertOneResult {
	one, err := r.Collection.Collection("messages").InsertOne(context.TODO(), message)
	if err != nil {
		return nil
	}
	return one
}

func (r MongoDBRepository) InsertRoom(room Room) *mongo.InsertOneResult {
	one, err := r.Collection.Collection("rooms").InsertOne(context.TODO(), room)
	if err != nil {
		return nil
	}
	return one
}

type RedisRepository struct {
	Redis *redis.Client
	ctx   context.Context
}

func NewRedisRepository() RedisRepository {
	return RedisRepository{Redis: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}),
		ctx: context.Background(),
	}
}

func (r RedisRepository) SetData(key string, value interface{}) {
	r.Redis.Set(r.ctx, key, value, 0)
}
func (r RedisRepository) GetData(key string) *redis.MapStringStringCmd {
	data := r.Redis.HRandField(r.ctx, key, 10)
	fmt.Println(data.Val())
	return r.Redis.HGetAll(r.ctx, key)
}

func (r RedisRepository) SetMessage(roomId string, messageId string, message Message) *redis.BoolCmd {
	out, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	return r.Redis.HMSet(r.ctx, roomId, messageId, string(out))
}
