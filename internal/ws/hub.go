package ws

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
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
	for {
		select {
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				r := h.Rooms[cl.RoomID]

				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl
				}
			}
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					if len(h.Rooms[cl.RoomID].Clients) != 0 {
						h.Broadcast <- &Message{
							Content:  "user left the chat",
							RoomID:   cl.RoomID,
							Username: cl.Username,
						}
					}

					delete(h.Rooms[cl.RoomID].Clients, cl.ID)
					close(cl.Message)
				}
			}

		case m := <-h.Broadcast:
			if _, ok := h.Rooms[m.RoomID]; ok {
				for _, cl := range h.Rooms[m.RoomID].Clients {

					h.SetMessage(m.RoomID, time.Now().Format("2006-01-02 15:04:05.002"), m)
					if cl.Username != m.Username {
						cl.Message <- m
					}

				}
			}
		}
	}
}
