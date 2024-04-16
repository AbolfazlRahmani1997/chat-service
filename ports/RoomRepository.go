package ports

import (
	"server/Dtos"
	"server/entity"
)

type RoomRepositoryPort interface {
	GetRoom(id string) entity.Room
	GetAllRooms(filter Dtos.GetAllRoomFilterDto) []entity.Room
	Update(Update Dtos.UpdateRoomDto) []entity.Room
}

type RoomServicePort interface {
	RetrieveRoom(id string) entity.Room
	FetchAllRooms(filter Dtos.GetAllRoomFilterDto) []entity.Room
	EditRooms(filter Dtos.UpdateRoomDto) []entity.Room
}
