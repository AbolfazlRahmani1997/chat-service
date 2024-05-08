package ports

import (
	"server/Dtos"
	"server/entity/Room"
)

type RoomRepositoryPort interface {
	GetRoom(id string) Room.Room
	GetAllRooms(page int, offset int, filter Dtos.GetAllRoomFilterDto) []Room.Room
	Update(Update Dtos.UpdateRoomDto) []Room.Room
}

type RoomServicePort interface {
	RetrieveRoom(id string) Room.Room
	FetchAllRooms(page int, offset int, filter Dtos.GetAllRoomFilterDto) []Room.Room
	EditRooms(filter Dtos.UpdateRoomDto) []Room.Room
}
