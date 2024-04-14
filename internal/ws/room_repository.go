package ws

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type RoomMongoRepository struct {
	MongoDbRepository *mongo.Database
}

func NewRoomRepository(client *mongo.Database) RoomMongoRepository {
	return RoomMongoRepository{MongoDbRepository: NewMongoDbRepository(client)}
}

func (r RoomMongoRepository) getById(roomId string) Room {
	var room Room
	filter := bson.M{
		"id": roomId,
	}
	cur := r.MongoDbRepository.Collection("rooms").FindOne(context.TODO(), filter)
	err := cur.Decode(&room)
	if err != nil {
		return Room{}
	}

	return room
}

func (r *RoomMongoRepository) GetMyRooms(userId string, page string) []Room {
	var rooms []Room
	filter := bson.M{
		"members.id": userId,
	}
	limit, _ := strconv.Atoi(page)
	l := int64(10)
	skip := int64(limit*10 - 10)
	findOptions := options.FindOptions{Skip: &skip, Limit: &l}
	opts := findOptions.SetSort(bson.D{{"last_message.created_at", -1}})

	cur, err := r.MongoDbRepository.Collection("rooms").Find(context.TODO(), filter, opts)
	err = cur.All(context.TODO(), &rooms)
	if err != nil {
		fmt.Println(err)
		return rooms
	}
	return rooms
}
func (r *RoomMongoRepository) GetOlineMyRooms(userId string) []Room {
	var rooms []Room
	filter := bson.M{
		"members.id": userId,
	}
	findOptions := options.FindOptions{}
	opts := findOptions.SetSort(bson.D{{"last_message.created_at", -1}})

	cur, err := r.MongoDbRepository.Collection("rooms").Find(context.TODO(), filter, opts)
	err = cur.All(context.TODO(), &rooms)
	if err != nil {
		fmt.Println(err)
		return rooms
	}
	return rooms
}

func (r RoomMongoRepository) insert(room Room) interface{} {
	result, _ := r.MongoDbRepository.Collection("rooms").InsertOne(context.TODO(), room)
	return result
}

func (r RoomMongoRepository) lastMessage(id string, message Message) bool {

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", _id}}
	message.CreatedAt = time.Now()

	update := bson.D{{"$set", bson.D{{"last_message", message}}}}
	_, err := r.MongoDbRepository.Collection("rooms").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}

	return true
}

func (r RoomMongoRepository) update(room Room) *mongo.SingleResult {
	fliter := bson.M{"id": room.ID}
	update := bson.D{{"$set", bson.D{{"status", room.Status}}}}
	result := r.MongoDbRepository.Collection("rooms").FindOneAndUpdate(context.Background(), fliter, update)

	return result
}
func (r RoomMongoRepository) updateMember(room Room) *mongo.SingleResult {

	fliter := bson.M{"id": room.ID}
	update := bson.D{{"$set", bson.D{{"members", room.Members}}}}
	result := r.MongoDbRepository.Collection("rooms").FindOneAndUpdate(context.Background(), fliter, update)
	return result
}
