package ws

type RoomService struct {
	RoomRepository   RoomMongoRepository
	MemberRepository UserRepository
}

func (r RoomService) GetMyRoom(userId string) []Room {
	room := r.RoomRepository.GetMyRooms(userId)
	for _, room := range room {
		r.SyncUser(room)
	}

	return room

}

func (r RoomService) UpdateLastMessage(room Room, message Message) {
	r.RoomRepository.lastMessage(room._Id.Hex(), message)
}

func NewRoomService(RoomRepository RoomMongoRepository, UserRepository UserRepository) RoomService {
	return RoomService{
		RoomRepository:   RoomRepository,
		MemberRepository: UserRepository,
	}
}
func (receiver RoomService) SyncUser(room Room) {
	var newMember []Member
	for _, m := range room.Members {
		member := receiver.MemberRepository.getUser(m.Id)
		m.FirstName = member.FirstName
		m.LastName = member.LastName
		newMember = append(newMember, m)

	}
	room.Members = newMember
	receiver.RoomRepository.updateMember(room)

}

func (r RoomService) changeRoomStatus(room Room) {
	r.RoomRepository.update(room)
}
