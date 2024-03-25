package ws

type RoomService struct {
	RoomRepository    RoomMongoRepository
	MemberRepository  UserRepository
	MessageRepository MessageRepository
}

type RoomResponse struct {
	Id                string   `json:"Id,omitempty"`
	Name              string   `json:"Name"`
	Members           []Member `json:"Members"`
	NotDeliverMessage int64    `json:"CountMessages"`
	LastMessage       Message  `json:"Message"`
}

func (r RoomService) GetMyRoom(userId string) []RoomResponse {

	var Rooms []RoomResponse
	room := r.RoomRepository.GetMyRooms(userId)
	for _, room := range room {
		roomSync := r.SyncUser(room)
		notDelivered := r.MessageRepository.Mongo.GetMessageNotCountDelivery(room.ID, userId)
		Rooms = append(Rooms, RoomResponse{Id: roomSync.ID, Name: roomSync.Name, Members: roomSync.Members, NotDeliverMessage: notDelivered, LastMessage: room.Message})
	}

	return Rooms

}

func (r RoomService) UpdateLastMessage(room Room, message Message) {
	r.RoomRepository.lastMessage(room._Id.Hex(), message)
}

func NewRoomService(RoomRepository RoomMongoRepository, MessageRepository MessageRepository, UserRepository UserRepository) RoomService {
	return RoomService{
		RoomRepository:    RoomRepository,
		MemberRepository:  UserRepository,
		MessageRepository: MessageRepository,
	}
}
func (receiver RoomService) SyncUser(room Room) Room {
	var newMember []Member
	for _, m := range room.Members {
		member := receiver.MemberRepository.getUser(m.Id)
		m.FirstName = member.FirstName
		m.LastName = member.LastName
		m.AvatarUrl = member.AvatarUrl
		newMember = append(newMember, m)

	}
	room.Members = newMember
	go receiver.RoomRepository.updateMember(room)
	return room
}

func (r RoomService) changeRoomStatus(room Room) {
	r.RoomRepository.update(room)
}
