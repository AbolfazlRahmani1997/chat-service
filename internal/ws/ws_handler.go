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
	room := h.hub.MessageService.MessageRepository.GetRoomById(req.ID)
	if room.ID != "" {
		h.hub.Rooms[room.ID] = &Room{
			ID:      room.ID,
			Name:    room.Name,
			Clients: make(map[string]*Client),
		}
	} else {
		h.hub.Rooms[req.ID] = &Room{
			ID:      req.ID,
			Name:    req.Name,
			Clients: make(map[string]*Client),
		}
		h.hub.MessageService.MessageRepository.Mongo.InsertRoom(*h.hub.Rooms[req.ID])
	}

	c.JSON(http.StatusOK, req)

}

func (h *Handler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	room := h.hub.MessageService.MessageRepository.GetRoomById(roomID)

	h.hub.RoomService.SyncUser(room)
	if _, ok := h.hub.Rooms[roomID]; !ok {
		h.hub.Room <- &room
	}
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

	h.hub.Register <- cl

	go cl.writeMessage()

	messages := h.hub.MessageService.MessageRepository.Mongo.GetMessageNotDelivery(roomID, clientID)
	if ok := len(messages) != 0; ok {
		for i := len(messages) - 1; i >= 0; i-- {
			h.hub.Broadcast <- &Message{
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
func (h *Handler) GetRooms(c *gin.Context) {
	userId := c.Param("userId")
	room := h.hub.RoomService.GetMyRoom(userId)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
	}
	err = conn.WriteJSON(room)
	if err != nil {
		fmt.Println(err)
		return
	}
	user := &User{
		Conn:         conn,
		UserId:       userId,
		online:       false,
		roomStatuses: make(chan *RoomStatus),
	}
	h.hub.Join <- user
	user.userConnection(h.hub)

}
func (h *Handler) SyncRoom(c *gin.Context) {
	userId := c.Param("userId")
	if _, ok := h.hub.Users[userId]; ok {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
		}
		h.hub.Users[userId].online = true
		h.hub.Users[userId].StatusConnection = conn
		h.hub.Users[userId].WireRooms(h.hub)
	} else {
		c.JSON(404, "not found")
	}

}
func (h *Handler) ReadMessage(c *gin.Context) {
	userId := c.Query("userId")
	roomId := c.Param("roomId")
	clients := h.hub.Rooms[roomId].Clients
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	client := clients[userId]
	client.ReadMessage = conn
	client.seenMessage(h.hub)
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
