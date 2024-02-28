package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MessageRepository struct {
	Mongo MongoDBRepository
	Redis RedisRepository
}

type MongoDBRepository struct {
	Collection *mongo.Database
}

func NewMongoDbRepository(client *mongo.Database) *mongo.Database {
	return client

}

func NewMessageRepository(client *mongo.Database) MessageRepository {

	return MessageRepository{
		Mongo: MongoDBRepository{NewMongoDbRepository(client)},
		Redis: NewRedisRepository(),
	}

}

// InsertInDb Insert In Db For StateFull
func (r MessageRepository) insertMessageInDb(message Message) *mongo.InsertOneResult {
	message.CreatedAt = time.Now()
	return r.Mongo.insertMessage(message)
}

func (r MongoDBRepository) insertMessage(message Message) *mongo.InsertOneResult {
	one, err := r.Collection.Collection("messages").InsertOne(context.TODO(), message)
	if err != nil {
		return nil
	}
	return one
}

func (r MongoDBRepository) GetAllMessages(roomId string) []Message {
	var data []Message
	condition := bson.M{"roomId": roomId}
	opts := options.Find().SetLimit(10)
	cur, _ := r.Collection.Collection("messages").Find(context.TODO(), condition, opts)
	err := cur.All(context.TODO(), &data)
	if err != nil {
		fmt.Println(err)
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

	one := r.Collection.Collection("rooms").FindOne(context.Background(), filter)
	err := one.Decode(&roomResult)
	if err != nil {
		fmt.Println(err)
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
		Addr:     "127.0.0.1:6379",
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
	data := r.Redis.HRandField(r.ctx, key, 100)
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
	return r.Mongo.getRoom(roomId)
}
func (r MessageRepository) MessageDelivery(id string, clientIds []string) (*mongo.UpdateResult, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", _id}}
	update := bson.D{{"$set", bson.D{{"deliver", clientIds}}}}
	result, err := r.Mongo.Collection.Collection("messages").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(result)
	return result, err
}

func (r RedisRepository) GetLen(key string) int64 {
	return r.Redis.LLen(context.TODO(), key).Val()
}

func (r RedisRepository) GetMessage(key string) string {
	return r.Redis.LPop(r.ctx, key).Val()
}
func (r RedisRepository) GetNotDeliverMessages(number int, key string) []Message {
	var messages []Message
	var message Message
	for i := 0; i < number; i++ {
		itemMessage := r.Redis.LPop(context.TODO(), key).Val()
		if itemMessage != "" {
			err := json.Unmarshal([]byte(itemMessage), &message)
			if err != nil {
				fmt.Println(err)
				return nil
			}
		}

		messages = append(messages, message)

	}

	return messages
}
func (r MessageRepository) getNumberNotDelivered(key string) int64 {
	return r.Redis.GetLen(key)
}

func (r MessageRepository) UpdateRoomById(id string, room Room) *mongo.UpdateResult {
	update := bson.D{{"$set", bson.D{{"Writer", room.Writer}, {"Owner", room.Owner}}}}
	filter := bson.D{{"_id", id}}
	byID, err := r.Mongo.Collection.Collection("rooms").UpdateByID(context.TODO(), filter, update)
	if err != nil {
		return nil
	}
	return byID
}
