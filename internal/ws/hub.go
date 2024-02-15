package ws

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Room struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	TeacherId string             `json:"teacherId"`
	StudentId string             `json:"studentId"`
	Clients   map[string]*Client `json:"clients"`
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	MessageService
}

func NewHub(client *mongo.Client) *Hub {
	clientDatabase := client.Database("MessageDB")
	messageRepository := NewMessageRepository(clientDatabase)
	service := MessageService{
		messageRepository,
	}
	return &Hub{
		Rooms:          make(map[string]*Room),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *Message, 5),
		MessageService: service,
	}
}

func (h *Hub) Run() {
	defer func() {
	}()

	for {
		select {
		//when user join the chat page
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				r := h.Rooms[cl.RoomID]
				if _, ok := r.Clients[cl.ID]; !ok {

					r.Clients[cl.ID] = cl
				} else {
					cl.Message = make(chan *Message)
				}
			}
			//when user exit from chat page
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
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
						if cl.Status == online {
							m.Deliver = append(m.Deliver, m.ClientID)
							h.MessageDelivery(m.ID, m.Deliver)
							cl.Message <- m
						} else {
							fmt.Println(m.RoomID + "." + cl.ID)

							message, e := json.Marshal(m)
							if e != nil {
								fmt.Println(e)
							}

							_, err := h.MessageRepository.Redis.Redis.LPush(h.MessageRepository.Redis.ctx, m.RoomID+"."+cl.ID, string(message)).Result()
							if err != nil {
								fmt.Println(err)
								return
							}
							//var ta Message
							//Dm := h.MessageRepository.Redis.Redis.LPop(h.MessageRepository.Redis.ctx, m.RoomID+"."+cl.ID).Val()
							//err = json.Unmarshal([]byte(Dm), &ta)
							//if err != nil {
							//	fmt.Println(err)
							//	return
							//}

						}

					}

				}
			}
			//when join chat system for show online

		}
	}
}
