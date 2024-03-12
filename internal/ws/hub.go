package ws

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Member struct {
	Id    string   `json:"Id"`
	Roles []string `json:"roles"`
}

type ReadMessage struct {
	MessageId string `json:"MessageId"`
	UserId    string `json:"UserId"`
}

type RoomStatus struct {
	RoomId string `json:"roomId"`
	Status status `json:"status"`
}

type RoomModel struct {
	ID        primitive.ObjectID `json:"_id"`
	RoomId    string             `json:"id"`
	Name      string             `json:"name" `
	Temporary bool               `json:"type" `
	Members   []Member           `json:"members"`
	Message   Message            `json:"message" bson:"last_message"`
	Status    status             `json:"status" `
	Clients   map[string]*Client `json:"clients"`
}
type Room struct {
	_Id       primitive.ObjectID `json:"_id"`
	ID        string             `json:"Id" `
	Name      string             `json:"name" `
	Temporary bool               `json:"type" `
	Members   []Member           `json:"members"`
	Message   Message            `json:"message" bson:"last_message"`
	Status    status             `json:"status" `
	Clients   map[string]*Client `json:"clients"`
}

type Hub struct {
	Users          map[string]*User
	Rooms          map[string]*Room
	Register       chan *Client
	ReadAble       chan *ReadMessage
	Join           chan *User
	Left           chan *User
	Unregister     chan *Client
	Broadcast      chan *Message
	Room           chan *Room
	MessageService MessageService
	RoomService    RoomService
	RoomBroker     RoomBrokerInfrastructure
}

func NewHub(client *mongo.Client) *Hub {
	clientDatabase := client.Database("MessageDB")
	messageRepository := NewMessageRepository(clientDatabase)
	RoomRepository := NewRoomRepository(clientDatabase)
	RoomService := NewRoomService(RoomRepository)
	service := MessageService{
		messageRepository,
	}
	roomChan := make(chan *Room)
	mqBroker := NewRabbitMqBroker(roomChan, messageRepository)

	mqBroker.Consume()

	return &Hub{
		Rooms:          make(map[string]*Room),
		Register:       make(chan *Client),
		ReadAble:       make(chan *ReadMessage),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *Message, 5),
		Join:           make(chan *User),
		Left:           make(chan *User),
		Users:          make(map[string]*User),
		Room:           roomChan,
		MessageService: service,
		RoomService:    RoomService,
	}

}

func (h *Hub) Run() {
	defer func() {
	}()

	for {
		select {
		case messageId := <-h.ReadAble:
			{
				h.MessageService.MessageRead(messageId.MessageId, messageId.UserId)
			}
		case room := <-h.Room:
			{
				h.Rooms[room.ID] = &Room{
					_Id:     room._Id,
					ID:      room.ID,
					Name:    room.Name,
					Members: room.Members,
					Clients: make(map[string]*Client),
				}

			}

		//when user join the chat page
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				r := h.Rooms[cl.RoomID]
				if _, ok := r.Clients[cl.ID]; ok {
					cl.Message = make(chan *Message)
					cl.Status = online

				} else {
					r.Clients[cl.ID] = cl
				}
				r.Clients[cl.ID] = cl
				room := h.Rooms[cl.RoomID]
				room.Status = online
				h.RoomService.changeRoomStatus(*room)
				members := h.Rooms[cl.RoomID].Members

				for _, member := range members {

					if user, ok := h.Users[member.Id]; ok {
						user.rooms <- &RoomStatus{
							RoomId: h.Rooms[cl.RoomID].ID,
							Status: online,
						}
					}
				}
			}
			//when user exit from chat page
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					fmt.Println(cl.Status)

					close(cl.Message)
					delete(h.Rooms[cl.RoomID].Clients, cl.ID)
				}
				if ok := len(h.Rooms[cl.RoomID].Clients) == 0; ok {
					room := h.Rooms[cl.RoomID]
					room.Status = offline
					h.RoomService.changeRoomStatus(*room)
					delete(h.Rooms, cl.ID)
				}
				members := h.Rooms[cl.RoomID].Members

				for _, member := range members {
					if user, ok := h.Users[member.Id]; ok {
						user.rooms <- &RoomStatus{
							RoomId: h.Rooms[cl.RoomID].ID,
							Status: offline,
						}
					}
				}
			}
		//when send message
		case m := <-h.Broadcast:
			if _, ok := h.Rooms[m.RoomID]; ok {
				if m.ID.IsZero() {
					m.Deliver = nil
					m.Read = nil
					m.ID = h.MessageService.MessageRepository.insertMessageInDb(*m).InsertedID.(primitive.ObjectID)

				}
				h.RoomService.UpdateLastMessage(*h.Rooms[m.RoomID], *m)
				for _, cl := range h.Rooms[m.RoomID].Clients {

					if cl.ID != m.ClientID {
						if ok := cl.Status == online; ok {
							m.Deliver = append(m.Deliver, cl.ID)
							h.MessageService.MessageDelivery(m.ID.Hex(), m.Deliver)
							cl.Message <- m

						}
					}

				}
			}
			//when join chat system for show online

		}
	}
}

func (h *Hub) Manager() {
	for {
		select {
		case user, _ := <-h.Join:
			h.Users[user.UserId] = user
		case user, _ := <-h.Left:
			delete(h.Users, user.UserId)
		}
	}
}
