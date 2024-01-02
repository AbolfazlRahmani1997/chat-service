package ws

import (
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

func (r RoomMongoRepository) insert(room Room) interface{} {
	result, _ := r.MongoDbRepository.Collection("rooms").InsertOne(context.TODO(), room)
	return result
}

func (r RoomMongoRepository) update(room Room) *mongo.UpdateResult {
	result, _ := r.MongoDbRepository.Collection("rooms").UpdateByID(context.TODO(), room.ID, room)
	return result
}
