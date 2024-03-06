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
	Rooms          map[string]*Room
	Register       chan *Client
	ReadAble       chan *ReadMessage
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
	mqBroker := NewRabbitMqBroker(roomChan)

	mqBroker.Consume()

	return &Hub{
		Rooms:          make(map[string]*Room),
		Register:       make(chan *Client),
		ReadAble:       make(chan *ReadMessage),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *Message, 5),
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
				h.MessageService.MessageRepository.insertRoomInDb(*room)
			}

		//when user join the chat page
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomID]; ok {

				r := h.Rooms[cl.RoomID]
				if _, ok := r.Clients[cl.ID]; ok {
					cl.Message = make(chan *Message)
					cl.Status = online

				} else {
					fmt.Println(cl.ID)
					r.Clients[cl.ID] = cl
				}
				r.Clients[cl.ID] = cl
				fmt.Println(len(r.Clients))
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
					delete(h.Rooms, cl.ID)
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
