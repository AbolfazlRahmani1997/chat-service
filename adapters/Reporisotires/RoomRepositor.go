package Reporisotires

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"server/Dtos"
	"server/entity/Room"
)

type RoomMongoRepository struct {
	Collection *mongo.Database
}

func NewRoomRepository(Collection *mongo.Client) RoomMongoRepository {
	return RoomMongoRepository{Collection: Collection.Database("MessageDB")}
}

func (receiver RoomMongoRepository) GetRoom(id string) Room.Room {
	var room Room.Room
	condition := bson.M{"id": id}
	find := receiver.Collection.Collection("rooms").FindOne(context.TODO(), condition)
	err := find.Decode(&room)
	if err != nil {
		fmt.Println(err)
		return Room.Room{}
	}
	return room
}

func (receiver RoomMongoRepository) GetAllRooms(page int, offset int, filter Dtos.GetAllRoomFilterDto) []Room.Room {
	var room []Room.Room
	find, err := receiver.Collection.Collection("rooms").Find(context.TODO(), filter.GetFilter())
	err = find.All(context.TODO(), &room)
	if err != nil {
		return []Room.Room{}
	}
	if err != nil {
		fmt.Println(err)
		return []Room.Room{}
	}
	if err != nil {
		fmt.Println(err)

		return []Room.Room{}
	}
	return room
}

func (receiver RoomMongoRepository) Update(Update Dtos.UpdateRoomDto) []Room.Room {

	result := receiver.Collection.Collection("rooms").FindOneAndUpdate(context.TODO(), bson.M{"id": Update.Id}, bson.D{{"$set", Update.GetUpdate()}})
	fmt.Println(result.Err())
	return nil
}
