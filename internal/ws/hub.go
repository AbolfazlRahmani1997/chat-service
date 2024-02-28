package ws

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// type RoomType string
//
// const (
//
//	Temporary RoomType = "temporary"
//	Permanent RoomType = "permanent"
//
// )
type Member struct {
	Id    string   `json:"Id"`
	Roles []string `json:"roles"`
}

type Room struct {
	_Id       primitive.ObjectID `bson:"_id"`
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Temporary bool               `json:"type"`
	Members   []Member           `json:"members"`
	Owner     []string           `json:"owner,omitempty" bson:"Owner"`
	Writer    []string           `json:"Writer,omitempty" bson:"Writer"`
	Clients   map[string]*Client `json:"clients"`
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	Room       chan *Room
	MessageService
	RoomBroker RoomBrokerInfrastructure
}

func NewHub(client *mongo.Client) *Hub {
	clientDatabase := client.Database("MessageDB")
	messageRepository := NewMessageRepository(clientDatabase)
	service := MessageService{
		messageRepository,
	}
	roomChan := make(chan *Room)
	mqBroker := NewRabbitMqBroker(roomChan)

	mqBroker.Consume()

	return &Hub{
		Rooms:          make(map[string]*Room),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *Message, 5),
		Room:           roomChan,
		MessageService: service,
	}

}

func (h *Hub) Run() {
	defer func() {
	}()

	for {
		select {
		case room := <-h.Room:
			{
				existRoom := h.MessageRepository.GetRoomById(room.ID)
				if existRoom.ID != "" {
					if len(existRoom.Owner) != len(room.Owner) || (len(existRoom.Writer) != len(room.Writer)) {
						h.MessageRepository.UpdateRoomById(room._Id.String(), *room)
					}
				} else {
					h.MessageRepository.insertRoomInDb(*room)
				}
				h.Rooms[room.ID] = &Room{
					ID:      room.ID,
					Name:    room.Name,
					Owner:   room.Owner,
					Writer:  room.Writer,
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
			}
			//when user exit from chat page
		case cl := <-h.Unregister:

			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					fmt.Println(cl.Status)
					cl.Status = offline
					close(cl.Message)
				}
			}
		//when send message
		case m := <-h.Broadcast:
			if _, ok := h.Rooms[m.RoomID]; ok {
				m.Deliver = nil
				m.Read = nil
				m._Id = h.MessageRepository.insertMessageInDb(*m).InsertedID.(primitive.ObjectID)
				m.ID = m._Id.Hex()
				for _, cl := range h.Rooms[m.RoomID].Clients {
					if cl.ID != m.ClientID {
						if ok := cl.Status == online; ok {
							m.Deliver = append(m.Deliver, cl.ID)
							h.MessageDelivery(m.ID, m.Deliver)
							cl.Message <- m
						} else {
							message, e := json.Marshal(m)
							if e != nil {
								fmt.Println(e)
							}
							_, err := h.MessageRepository.Redis.Redis.LPush(h.MessageRepository.Redis.ctx, m.RoomID+"."+cl.ID, string(message)).Result()
							if err != nil {
								fmt.Println(err)
								return
							}

						}

					}

				}
			}
			//when join chat system for show online

		}
	}
}
