package adapters

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"server/Dtos"
	"server/entity"
)

type RoomMongoRepository struct {
	Collection *mongo.Database
}

func NewRoomRepository(Collection *mongo.Client) RoomMongoRepository {
	return RoomMongoRepository{Collection: Collection.Database("MessageDB")}
}

func (receiver RoomMongoRepository) GetRoom(id string) entity.Room {
	var room entity.Room
	fmt.Println(id)
	condition := bson.M{"id": id}
	find := receiver.Collection.Collection("rooms").FindOne(context.TODO(), condition)
	err := find.Decode(&room)
	if err != nil {
		fmt.Println(err)
		return entity.Room{}
	}
	return room
}

func (receiver RoomMongoRepository) GetAllRooms(filter Dtos.GetAllRoomFilterDto) []entity.Room {
	var room []entity.Room
	find, err := receiver.Collection.Collection("rooms").Find(context.TODO(), filter.GetFilter())
	err = find.All(context.TODO(), &room)
	if err != nil {
		return []entity.Room{}
	}
	if err != nil {
		fmt.Println(err)
		return []entity.Room{}
	}
	if err != nil {
		fmt.Println(err)

		return []entity.Room{}
	}
	return room
}

func (receiver RoomMongoRepository) Update(Update Dtos.UpdateRoomDto) []entity.Room {
	result := receiver.Collection.Collection("rooms").FindOneAndUpdate(context.TODO(), bson.M{"roomId": Update.Id}, Update.Room)
	fmt.Println(result.Err())
	return nil
}
