package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
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
	ID      primitive.ObjectID `json:"_id"`
	RoomId  string             `json:"id"`
	Name    string             `json:"name" `
	Type    Type               `json:"type" `
	Members []Member           `json:"members"`
	Message Message            `json:"message" bson:"last_message"`
	Status  status             `json:"status" `
	Clients map[string]*Client `json:"clients"`
}
type Type string

const (
	PRIVATE Type = "PRIVATE"
	GROUP   Type = "GROUP"
	TEMP    Type = "TEMP"
)

type Room struct {
	_Id     primitive.ObjectID `json:"_id"`
	ID      string             `json:"Id" `
	Name    string             `json:"Name" `
	Type    Type               `json:"Type" `
	Members []Member           `json:"Members"`
	Message Message            `json:"Message" bson:"last_message"`
	Status  status             `json:"Status,omitempty" `
	Clients map[string]*Client `json:"Clients"`
	Pinned  bool               `json:"Pinned" bson:"pinned"`
}

type RoomTemp struct {
	_Id       primitive.ObjectID `json:"_id"`
	ID        string             `json:"Id" `
	Name      string             `json:"Name" `
	Temporary bool               `json:"Type" `
	Members   Member             `json:"Members"`
	Message   Message            `json:"Message" bson:"last_message"`
	Status    status             `json:"Status,omitempty" `
	Clients   map[string]*Client `json:"Clients"`
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

	clientDatabase := client.Database(os.Getenv("CHAT_DB"))
	messageRepository := NewMessageRepository(clientDatabase)
	userRepository := NewUserRepository(client.Database(os.Getenv("MAIN_DB")))
	RoomRepository := NewRoomRepository(clientDatabase)
	RoomService := NewRoomService(RoomRepository, messageRepository, userRepository)
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
				go func() {
					message := h.MessageService.MessageRead(messageId.MessageId, messageId.UserId)
					if user, ok := h.Users[message.ClientID]; ok {
						go func() {
							user.seenMessage <- &SeenNotification{MessageId: message.UniqId, RoomId: message.RoomID}
						}()
					}

				}()

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
				if room._Id.IsZero() {
					SyncedRoom := h.RoomService.SyncUser(*room)
					for _, member := range room.Members {
						if user, ok := h.Users[member.Id]; ok {
							fmt.Println(ok)
							if user.IsConnected {
								go func() {
									user.createRoom <- &SyncedRoom
								}()
							}
						}
					}
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
				for s := range cl.Conn {
					go cl.readerMessage(s, h)
				}
				r.Clients[cl.ID] = cl
				room := h.Rooms[cl.RoomID]
				room.Status = online
				go func() {
					members := h.Rooms[cl.RoomID].Members
					for _, member := range members {
						if user, ok := h.Users[member.Id]; ok {
							if cl.ID != user.UserId {
								go func() {
									user.roomStatuses <- &RoomStatus{
										RoomId: h.Rooms[cl.RoomID].ID,
										Status: online,
									}
								}()
							}

						}
					}

				}()

			}
			//when user exit from chat page
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
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
			go func() {
				m.CreatedAt = time.Now()
				if _, ok := h.Rooms[m.RoomID]; ok {
					if m.ID.IsZero() {
						m.Deliver = nil
						m.Read = nil
						m.ID = h.MessageService.MessageRepository.insertMessageInDb(*m).InsertedID.(primitive.ObjectID)
					}
					h.RoomService.UpdateLastMessage(*h.Rooms[m.RoomID], *m)

					members := h.Rooms[m.RoomID].Members
					var firstname string
					var lastname string
					for _, userID := range members {

						if userID.Id == m.ClientID {

							firstname = userID.FirstName
							lastname = userID.LastName
						}

						if user, ok := h.Users[userID.Id]; ok {
							if _, ok := h.Rooms[m.RoomID].Clients[user.UserId]; !ok {
								if user.UserId != m.ClientID {

									go func() {
										user.pupMessage <- &PupMessage{
											MessageId: m.ID.Hex(),
											RoomId:    m.RoomID,
											Firstname: firstname,
											Lastname:  lastname,
											Content:   m.Content,
										}
									}()

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
			}()

		}
		//when join chat system for show online

	}
}

func (h *Hub) Manager() {
	for {
		select {
		case user, _ := <-h.Join:
			if userExists, ok := h.Users[user.UserId]; ok {
				if h.Users[user.UserId].IsConnected == false {
					h.Users[user.UserId].IsConnected = true
				}
				userExists.Conn = mergeConnection(userExists.Conn, user.Conn)

			} else {
				user.roomStatuses = make(chan *RoomStatus)
				user.pupMessage = make(chan *PupMessage)
				user.chanelNotification = make(chan *SystemMessage)
				user.createRoom = make(chan *Room)
				user.seenMessage = make(chan *SeenNotification)
				go user.WireRooms(h)
				h.Users[user.UserId] = user
			}
			for s := range user.Conn {
				go user.userConnection(h, s)
			}
			go h.OnlineMessage(user.UserId, online)
		case user, _ := <-h.Left:

			go h.OnlineMessage(user.UserId, offline)
			if len(user.Conn) == 0 {
				user.IsConnected = false
			}
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
