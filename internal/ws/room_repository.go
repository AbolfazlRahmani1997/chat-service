package ws

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
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
	cur, _ := r.MongoDbRepository.Collection("rooms").Find(context.TODO(), filter)
	err := cur.All(context.TODO(), &room)
	if err != nil {
		return Room{}
	}
	return room
}

func (receiver RoomMongoRepository) GetMyRooms(userId string) []Room {
	var rooms []Room
	filter := bson.M{
		"members.id": userId,
	}

	cur, _ := receiver.MongoDbRepository.Collection("rooms").Find(context.TODO(), filter)
	err := cur.All(context.TODO(), &rooms)
	if err != nil {
		return rooms
	}
	return rooms
}

func (r RoomMongoRepository) insert(room Room) interface{} {
	result, _ := r.MongoDbRepository.Collection("rooms").InsertOne(context.TODO(), room)
	return result
}

func (r RoomMongoRepository) update(room Room) *mongo.UpdateResult {
	fliter := bson.D{{"$set", bson.D{{"Status", room.Status}}}}
	result, err := r.MongoDbRepository.Collection("rooms").UpdateByID(context.TODO(), room.ID, fliter)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return result
}
