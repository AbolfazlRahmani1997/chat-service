package Dtos

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"server/entity/Room"
)

type GetAllRoomDto struct {
}

type GetAllRoomFilterDto struct {
	Id       string `json:"Id,"`
	MemberId string `json:"MemberId"`
}

func (receiver GetAllRoomFilterDto) GetFilter() bson.M {
	filters := bson.M{}

	if receiver.MemberId != "" {
		filters["members.id"] = receiver.MemberId
	}
	if receiver.Id != "" {
		filters["id"] = receiver.Id

	}
	return filters
}

type UpdateRoomDto struct {
	Id   string    `json:"Id,omitempty"`
	Room Room.Room `json:"Room,omitempty"`
}

func (receiver UpdateRoomDto) GetUpdate() bson.D {
	filters := bson.D{}
	fmt.Println(receiver.Room.Name)
	if receiver.Room.Name != "" {
		filters = append(filters, bson.E{Key: "name", Value: receiver.Room.Name})
	}
	if receiver.Room.Status != "" {

	}
	return filters

}

type MessageEntity struct {
}

type RoomResource struct {
}

type AllRoomRequest struct {
}
