package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Member struct {
	Id           string   `json:"Id"`
	Roles        []string `json:"roles"`
	FirstName    string   `json:"firstname"`
	LastName     string   `json:"lastname"`
	AvatarUrl    string   `json:"AvatarUrl"bson:"avatar_url"`
	Notification bool     `json:"Notification"`
	Pin          bool     `json:"Pin"`
}

type ReadMessage struct {
	MessageId string `json:"MessageId"`
	UserId    string `json:"UserId"`
}

type RoomStatus struct {
	RoomId  string `json:"roomId"`
	Status  status `json:"status"`
	Message string `json:"Message"`
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
	Status    status             `json:"status,omitempty" `
	Clients   map[string]*Client `json:"clients"`
}

type Hub struct {
	Users          map[string]*User
	Rooms          map[string]*Room
	Register       chan *Client
	ReadAble       chan *ReadMessage
	Join           chan *User
	Left           chan *User
	Evade          chan *User
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
	userRepository := NewUserRepository(client.Database("main"))
	RoomRepository := NewRoomRepository(clientDatabase)
	RoomService := NewRoomService(RoomRepository, messageRepository, userRepository)
	service := MessageService{
		messageRepository,
	}
	roomChan := make(chan *Room)
	//mqBroker := NewRabbitMqBroker(roomChan, messageRepository)
	//
	//mqBroker.Consume()

	return &Hub{
		Rooms:          make(map[string]*Room),
		Register:       make(chan *Client),
		ReadAble:       make(chan *ReadMessage),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *Message, 5),
		Join:           make(chan *User),
		Left:           make(chan *User),
		Evade:          make(chan *User),
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
				if client, ok := r.Clients[cl.ID]; ok {
					cl.Conn = mergeConnection(cl.Conn, client.Conn)
					cl.Message = make(chan *Message)
					cl.Status = online

				} else {
					r.Clients[cl.ID] = cl
				}
				for s, _ := range cl.Conn {
					go cl.readerMessage(s, h)
				}
				r.Clients[cl.ID] = cl
				room := h.Rooms[cl.RoomID]
				room.Status = online
				h.RoomService.changeRoomStatus(*room)
				members := h.Rooms[cl.RoomID].Members

				for _, member := range members {
					if user, ok := h.Users[member.Id]; ok {

						go func() {
							user.roomStatuses <- &RoomStatus{
								RoomId: h.Rooms[cl.RoomID].ID,
								Status: online,
							}
						}()

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
						go func() {
							user.roomStatuses <- &RoomStatus{
								RoomId: h.Rooms[cl.RoomID].ID,
								Status: offline,
							}
						}()
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

				members := h.Rooms[m.RoomID].Members
				for _, userID := range members {
					if userID.Notification != false {
						if user, ok := h.Users[userID.Id]; ok {
							if h.Rooms[m.RoomID].Clients[user.UserId] == nil {
								if user.UserId != m.ClientID {
									go func() {
										user.pupMessage <- &PupMessage{
											MessageId: m.ID.Hex(),
											RoomId:    m.RoomID,
											Content:   m.Content,
										}
									}()

								}
							}

						}
					}

				}
				for _, cl := range h.Rooms[m.RoomID].Clients {

					if ok := cl.Status == online; ok {
						if cl.ID != m.ClientID {
							m.Deliver = append(m.Deliver, cl.ID)
							h.MessageService.MessageDelivery(m.ID.Hex(), m.Deliver)
						}
						cl.Message <- m

					}

				}

			}
		}
		//when join chat system for show online

	}
}

func (h *Hub) Manager() {
	for {
		select {
		case user, _ := <-h.Join:
			if userExists, ok := h.Users[user.UserId]; ok {
				userExists.Conn = mergeConnection(userExists.Conn, user.Conn)

			} else {

				user.roomStatuses = make(chan *RoomStatus)
				user.pupMessage = make(chan *PupMessage)
				user.chanelNotification = make(chan *SystemMessage)
				go user.WireRooms(h)
				h.Users[user.UserId] = user
			}
			for s, _ := range user.Conn {
				go user.userConnection(h, s)
			}
			go h.OnlineMessage(user.UserId, online)
		case user, _ := <-h.Left:
			go h.OnlineMessage(user.UserId, offline)
			delete(h.Users, user.UserId)
		case user, _ := <-h.Evade:
			go h.OnlineMessage(user.UserId, evade)
		}

	}
}

func (h *Hub) OnlineMessage(userId string, status status) {
	rooms := h.RoomService.RoomRepository.GetOlineMyRooms(userId)
	for _, room := range rooms {
		for _, client := range room.Members {
			if client.Id != userId {
				if user, ok := h.Users[client.Id]; ok {
					user.roomStatuses <- &RoomStatus{
						RoomId: room.ID,
						Status: status,
					}
				}
			}

		}
	}

}

func mergeConnection(m1 map[string]*websocket.Conn, m2 map[string]*websocket.Conn) map[string]*websocket.Conn {
	merged := make(map[string]*websocket.Conn)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
