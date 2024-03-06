package ws

type RoomService struct {
	RoomRepository RoomMongoRepository
}

func (receiver RoomService) GetMyRoom(userId string) []Room {
	return receiver.RoomRepository.GetMyRooms(userId)

}

func NewRoomService(RoomRepository RoomMongoRepository) RoomService {
	return RoomService{
		RoomRepository: RoomRepository,
	}
}
