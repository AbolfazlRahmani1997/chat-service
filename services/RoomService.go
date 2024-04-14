package services

import (
	"server/Dtos"
	"server/entity"
	"server/ports"
)

type RoomService struct {
	RoomRepository ports.RoomRepositoryPort
}

func NewRoomService(RoomRepository ports.RoomRepositoryPort) ports.RoomServicePort {

	return RoomService{RoomRepository: RoomRepository}

}
func (receiver RoomService) RetrieveRoom(id string) entity.Room {
	room := receiver.RoomRepository.GetRoom(id)
	return room
}

func (receiver RoomService) FetchAllRooms(filter Dtos.GetAllRoomFilterDto) []entity.Room {

	rooms := receiver.RoomRepository.GetAllRooms(filter)

	return rooms
}

func (receiver RoomService) EditRooms(filter Dtos.UpdateRoomDto) []entity.Room {

	rooms := receiver.RoomRepository.Update(filter)

	return rooms
}
