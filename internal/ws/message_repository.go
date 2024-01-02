package ws

import (
	"context"
	_ "context"
	"encoding/json"
	"fmt"
	_ "fmt"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
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

func (r MongoDBRepository) GetAllMessages(roomId string) []Message {
	var data []Message
	condition := bson.M{"roomid": roomId}
	cur, _ := r.Collection.Collection("messages").Find(context.TODO(), condition)
	err := cur.All(context.TODO(), &data)
	if err != nil {
		return nil
	}
	return data
}

func (r MongoDBRepository) InsertRoom(room Room) *mongo.InsertOneResult {
	one, err := r.Collection.Collection("rooms").InsertOne(context.TODO(), room)
	if err != nil {
		return nil
	}
	return one
}

func (r MongoDBRepository) getRoom(roomId string) Room {

	var roomResult Room
	filter := bson.M{"id": roomId}
	one := r.Collection.Collection("rooms").FindOne(context.TODO(), filter)
	err := one.Decode(&roomResult)
	if err != nil {
		fmt.Println(err.Error())
		return Room{}
	}
	return roomResult
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
func (r RedisRepository) GetData(key string) *redis.SliceCmd {
	data := r.Redis.HRandField(r.ctx, key, 1000)
	return r.Redis.HMGet(r.ctx, key, data.Val()...)
}

func (r RedisRepository) SetMessage(roomId string, messageId string, message Message) *redis.BoolCmd {
	out, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}
	return r.Redis.HMSet(r.ctx, roomId, messageId, string(out))
}

func (r MessageRepository) GetRoomById(roomId string) Room {
	return r.MongoDBRepository.getRoom(roomId)
}
