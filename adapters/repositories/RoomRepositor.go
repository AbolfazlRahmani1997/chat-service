package repositories

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"os"
	"server/Dtos"
	"server/entity/Room"
)

type RoomMongoRepository struct {
	Collection *mongo.Database
}

func NewRoomRepository(Collection *mongo.Client) RoomMongoRepository {
	return RoomMongoRepository{Collection: Collection.Database(os.Getenv("CHAT_DB"))}
}

func (receiver RoomMongoRepository) GetRoom(id string) Room.Room {
	var room Room.Room
	fmt.Println(room.GetTableName())

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
	skip := int64((page * offset) - offset)
	opts := options.Find().SetSkip(skip).SetLimit(int64(offset))
	find, err :=
		receiver.Collection.Collection("rooms").Find(context.TODO(), filter.GetFilter(), opts)
	if err != nil {
		fmt.Println(err)
	}
	err = find.All(context.TODO(), &room)

	if err != nil {
		fmt.Println(err)

		return []Room.Room{}
	}
	return room
}

func (receiver RoomMongoRepository) Update(Update Dtos.UpdateRoomDto) []Room.Room {
	result := receiver.Collection.Collection("rooms").FindOneAndUpdate(context.TODO(), bson.M{"roomId": Update.Id}, Update.Room)
	fmt.Println(result.Err())
	return nil
}
