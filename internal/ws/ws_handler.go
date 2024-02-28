package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *Hub
	MessageService
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

type CreateRoomReq struct {
	ID     string   `json:"Id"`
	Name   string   `json:"Name"`
	Member []Member `json:"Member"`
	Owner  []string `json:"Owner"`
	Writer []string `json:"Writer"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room := h.hub.MessageRepository.GetRoomById(req.ID)
	if room.ID != "" {
		h.hub.Rooms[room.ID] = &Room{
			ID:      room.ID,
			Name:    room.Name,
			Owner:   room.Owner,
			Writer:  room.Writer,
			Clients: make(map[string]*Client),
		}
	} else {
		h.hub.Rooms[req.ID] = &Room{
			ID:      req.ID,
			Name:    req.Name,
			Owner:   req.Owner,
			Writer:  req.Writer,
			Clients: make(map[string]*Client),
		}
		h.hub.MessageRepository.Mongo.InsertRoom(*h.hub.Rooms[req.ID])
	}

	c.JSON(http.StatusOK, req)

}

func (h *Handler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	room := h.hub.MessageRepository.GetRoomById(roomID)
	h.hub.Room <- &room
	clientID := c.Query("userId")
	username := c.Query("username")

	userOwner, _, roles := hasAccess(clientID, room.Members, []string{"Owner", "Writer"})
	if !(userOwner) {
		c.JSON(http.StatusForbidden, "Access Deny ")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	cl := &Client{
		Conn:     conn,
		ID:       clientID,
		RoomID:   roomID,
		Username: username,
		Status:   online,
	}
	if roles == nil {
		cl.Message = make(chan *Message, 3)

		err := conn.WriteJSON([]string{"templates", "Salam Chetori", "ساعت ازاد کن "})
		if err != nil {
			return
		}

	} else {
		cl.Message = make(chan *Message)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		return
	}

	err = conn.WriteJSON(h.hub.MessageService.MessageRepository.Mongo.GetAllMessages(roomID))
	if err != nil {
		fmt.Print(err)
		return
	}

	h.hub.Register <- cl
	go cl.writeMessage()
	messages := h.hub.MessageRepository.Redis.GetNotDeliverMessages(int(h.hub.MessageRepository.Redis.GetLen(roomID+"."+cl.ID)), roomID+"."+cl.ID)
	if ok := len(messages) != 0; ok {
		for i := len(messages) - 1; i >= 0; i-- {
			h.hub.Broadcast <- &Message{
				_Id:       messages[i]._Id,
				ID:        messages[i].ID,
				Content:   messages[i].Content,
				RoomID:    messages[i].RoomID,
				Username:  messages[i].Username,
				ClientID:  messages[i].ClientID,
				Deliver:   messages[i].Deliver,
				Read:      messages[i].Read,
				CreatedAt: messages[i].CreatedAt,
				UpdatedAt: messages[i].CreatedAt,
			}
			if err != nil {
				return
			}
		}
	}

	cl.readMessage(h.hub)

}

type RoomRes struct {
	ID            string   `json:"id" `
	Name          string   `json:"name"`
	NumberMessage int64    `json:"NumberMessage"`
	Writer        []string `json:"Writer"`
	Owner         []string `json:"Owner"`
}

func (h *Handler) GetRooms(c *gin.Context) {
	rooms := make([]RoomRes, 0)
	userId := c.Param("userId")
	for _, r := range h.hub.Rooms {
		WriterStatus, Role, _ := hasAccess(userId, r.Members, []string{"Owner", "Writer"})
		fmt.Println(Role)
		if WriterStatus {
			fmt.Println(r.ID + "." + userId)
			rooms = append(rooms, RoomRes{
				ID:            r.ID,
				Name:          r.Name,
				NumberMessage: h.hub.MessageRepository.Redis.GetLen(r.ID + "." + userId),
				Writer:        r.Writer,
				Owner:         r.Owner,
			})
		}

	}

	c.JSON(http.StatusOK, rooms)
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) GetClients(c *gin.Context) {
	var clients []ClientRes
	roomId := c.Param("roomId")
	if _, ok := h.hub.Rooms[roomId]; !ok {
		clients = make([]ClientRes, 0)
		c.JSON(http.StatusOK, clients)
	}

	for _, c := range h.hub.Rooms[roomId].Clients {
		clients = append(clients, ClientRes{
			ID:       c.ID,
			Username: c.Username,
		})
	}

	c.JSON(http.StatusOK, clients)
}
func hasAccess(val interface{}, array interface{}, access interface{}) (exists bool, index int, role interface{}) {
	exists = false
	index = -1
	role = nil
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			arrayData := s.Index(i)

			if reflect.DeepEqual(arrayData.Field(0).String(), val.(string)) == true {
				remissions := reflect.ValueOf(arrayData.Field(1).Interface())
				for j := 0; j < remissions.Len(); j++ {
					exist, _ := InArray(arrayData.Field(1).Index(j).String(), access)
					if exist {

						role = arrayData.Field(1).Interface()
					}

				}
				index = i
				exists = true
				return

			}
		}
	}

	return
}

func InArray(needle interface{}, haystack interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(haystack)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}
