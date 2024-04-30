package ws

import (
	"fmt"
	"strconv"
)

type RoomService struct {
	RoomRepository    RoomMongoRepository
	MemberRepository  UserRepository
	MessageRepository MessageRepository
}

type RoomResponse struct {
	Id                string   `json:"Id,omitempty"`
	Name              string   `json:"Name"`
	Members           []Member `json:"Members"`
	Status            status   `json:"Status"`
	NotDeliverMessage int64    `json:"CountMessages"`
	LastMessage       Message  `json:"Message"`
}

func (r RoomService) GetMyRoom(userId string, page string) []RoomResponse {
	var Rooms []RoomResponse

	pageA, _ := strconv.Atoi(page)
	room := r.RoomRepository.GetMyPinRooms(userId)
	if (len(room) < 4) && pageA == 1 {
		roomList := r.RoomRepository.GetMyRooms(userId, pageA, 10-len(room))
		room = append(room, roomList...)
	} else {
		room = r.RoomRepository.GetMyRooms(userId, pageA, 10-len(room))
	}
	for _, room := range room {
		roomSync := r.SyncUser(room)
		notDelivered := r.MessageRepository.Mongo.GetMessageNotCountDelivery(room.ID, userId)
		Rooms = append(Rooms, RoomResponse{Id: roomSync.ID, Name: roomSync.Name, Members: roomSync.Members, NotDeliverMessage: notDelivered, LastMessage: room.Message, Status: room.Status})
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

type SpecificationRoom struct {
	Notification bool `json:"Notification"`
	Pin          bool `json:"Pin"`
}

func (receiver RoomService) updateRoomSpecification(id string, userId string, notification SpecificationRoom) []Member {
	var NewMember []Member
	fmt.Println(userId)
	room := receiver.RoomRepository.getById(id)
	member := room.Members
	var LastStatus SpecificationRoom
	for _, m := range member {

		if m.Id == userId {
			if notification.Notification == true {
				m.Notification = !m.Notification
				LastStatus.Notification = m.Notification
			}
			if notification.Pin == true {
				if !m.Pin {
					room := receiver.RoomRepository.GetMyPinRooms(userId)
					if len(room) < 4 {
						m.Pin = !m.Pin
					}
				} else {
					m.Pin = !m.Pin
				}

				LastStatus.Pin = m.Pin
			}

		}
		NewMember = append(NewMember, m)
	}

	room.Members = NewMember
	receiver.RoomRepository.updateMember(room)
	return NewMember
}

func (r RoomService) changeRoomStatus(room Room) {
	r.RoomRepository.update(room)
}
