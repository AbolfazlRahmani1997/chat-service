package ws

type RoomService struct {
	RoomRepository RoomMongoRepository
}

func (r RoomService) GetMyRoom(userId string) []Room {
	return r.RoomRepository.GetMyRooms(userId)

}

func NewRoomService(RoomRepository RoomMongoRepository) RoomService {
	return RoomService{
		RoomRepository: RoomRepository,
	}
}

func (r RoomService) changeRoomStatus(room Room) {
	r.RoomRepository.update(room)
}
