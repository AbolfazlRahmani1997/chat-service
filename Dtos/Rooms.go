package Dtos

import (
	"go.mongodb.org/mongo-driver/bson"
	"server/entity"
)

type GetAllRoomDto struct {
}

type GetAllRoomFilterDto struct {
	Id       string `json:"Id"`
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
	Id   string      `json:"Id,omitempty"`
	Room entity.Room `json:"Room,omitempty"`
}

func (receiver UpdateRoomDto) GetUpdate() {
	filters := bson.M{}

	if receiver.Room.Name != "" {
		filters["name"] = receiver.Room.Name
	}
	if receiver.Room.Status != "" {
		filters["status"] = receiver.Room.Status

	}

}

type MessageEntity struct {
}

type RoomResource struct {
}

type AllRoomRequest struct {
}
