package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub         *Hub
	UserHandler map[string]UserRequest
	MessageService
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub:         h,
		UserHandler: make(map[string]UserRequest),
	}
}

type CreateRoomReq struct {
	ID     string   `json:"Id"`
	Name   string   `json:"Name"`
	Member []Member `json:"Member"`
	Owner  []string `json:"Owner"`
	Writer []string `json:"Writer"`
}

func (Handler *Handler) UpdateUser(dto UserDto) {
	Handler.hub.RoomService.MemberRepository.UpdateUser(dto.UserId, dto)
}

// Store UpdateUserPool in this Pool
func (Handler *Handler) UpdateUserPool() {
	ticker := time.NewTicker((60 * 60) * time.Second)
	for {
		select {
		case <-ticker.C:
			for i, _ := range Handler.UserHandler {
				if Handler.UserHandler[i].Time.Before(time.Now()) {
					delete(Handler.UserHandler, i)
				}
			}
		}
	}
}

func (Handler *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room := Handler.hub.MessageService.MessageRepository.GetRoomById(req.ID)
	if room.ID != "" {
		Handler.hub.Rooms[room.ID] = &Room{
			ID:      room.ID,
			Name:    room.Name,
			Clients: make(map[string]*Client),
		}
	} else {
		Handler.hub.Rooms[req.ID] = &Room{
			ID:      req.ID,
			Name:    req.Name,
			Clients: make(map[string]*Client),
		}
		Handler.hub.MessageService.MessageRepository.Mongo.InsertRoom(*Handler.hub.Rooms[req.ID])
	}

	c.JSON(http.StatusOK, req)

}

func (Handler *Handler) JoinRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	roomID := c.Param("roomId")
	room := Handler.hub.MessageService.MessageRepository.GetRoomById(roomID)

	token := c.Query("token")
	token = fmt.Sprintf("%s", token)
	userAuthed, err := Handler.getUser(token)
	if err != nil {
		conn.Close()
		return
	}
	clientID := strconv.Itoa(userAuthed.Id)

	page := c.Query("page")

	userOwner, _, roles := hasAccess(clientID, room.Members, []string{"Owner", "Writer"})
	if !(userOwner) {
		c.JSON(http.StatusForbidden, "Access Deny ")
		err := conn.Close()
		if err != nil {
			return
		}
		return
	}

	go Handler.hub.RoomService.SyncUser(room)
	if _, ok := Handler.hub.Rooms[roomID]; !ok {
		Handler.hub.Room <- &room
	}
	var cl *Client
	var connectionPool map[string]*websocket.Conn
	connectionPool = make(map[string]*websocket.Conn)
	connectionPool[ulid.Make().String()] = conn

	cl = &Client{
		Conn:          connectionPool,
		ID:            clientID,
		RoomID:        roomID,
		ChanelMessage: make(chan *Message),
		Status:        online,
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

	Handler.hub.Register <- cl

	go cl.writeMessage()
	messagesRoom := Handler.hub.MessageService.MessageRepository.Mongo.GetRoomMessage(roomID, page)
	err = conn.WriteJSON(messagesRoom)
	if err != nil {
		return
	}
	messages := Handler.hub.MessageService.MessageRepository.Mongo.GetMessageNotDelivery(roomID, clientID)
	if ok := len(messages) != 0; ok {
		for i := 0; i <= len(messages)-1; i++ {
			messages[i].Deliver = append(messages[i].Deliver, clientID)
			_, err := Handler.hub.MessageService.MessageRepository.MessageDelivery(messages[i].ID.Hex(), messages[i].Deliver)
			if err != nil {
				return
			}

			if err != nil {
				return
			}
		}
	}
	cl.readMessage(Handler.hub)

}
func (Handler *Handler) GetRooms(c *gin.Context) {
	token := c.Query("token")
	token = fmt.Sprintf("%s", token)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	userAuthed, err := Handler.getUser(token)
	if err != nil {
		conn.Close()
		c.JSON(403, "cant authorization")
		return
	}
	userId := strconv.Itoa(userAuthed.Id)
	room := Handler.hub.RoomService.GetMyRoom(userId, "1")
	err = conn.WriteJSON(room)
	if err != nil {

		return
	}
	var connectionPool map[string]*websocket.Conn
	connectionPool = make(map[string]*websocket.Conn)
	connectionPool[ulid.Make().String()] = conn
	user := &User{
		Conn:   connectionPool,
		UserId: userId,
		online: true,
	}
	Handler.hub.Join <- user

}
func (Handler *Handler) SyncRoom(c *gin.Context) {
	token := c.Query("token")
	token = fmt.Sprintf("%s", token)
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	userAuthed, err := Handler.getUser(token)

	if err != nil {
		conn.Close()
		return
	}
	userId := strconv.Itoa(userAuthed.Id)
	if _, ok := Handler.hub.Users[userId]; !ok {
		c.JSON(404, "not found")
	} else {
		Handler.hub.Users[userId].online = true
		Handler.hub.Users[userId].StatusConnection = conn
		Handler.hub.Users[userId].WireRooms(Handler.hub)
	}

}
func (Handler *Handler) ReadMessage(c *gin.Context) {
	userId := c.GetString("userId")
	roomId := c.Param("roomId")
	clients := Handler.hub.Rooms[roomId].Clients
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	client := clients[userId]
	client.ReadMessage = conn
	client.seenMessage(Handler.hub)
}

func (Handler *Handler) UpdatePin(c *gin.Context) {
	token := c.GetHeader("Authorization")
	user, err := Handler.getUser(token)
	if err != nil {
		c.JSON(403, "cant authorization")
		return
	}
	var spefic SpecificationRoom
	spefic.Pin = true
	roomId := c.Param("roomId")
	userId := strconv.Itoa(user.Id)

	newMember := Handler.hub.RoomService.updateRoomSpecification(roomId, userId, spefic)
	if _, ok := Handler.hub.Rooms[roomId]; ok {
		Handler.hub.Rooms[roomId].Members = newMember
	}
	for _, member := range newMember {
		if member.Id == userId {
			spefic.Pin = member.Pin
			spefic.Notification = member.Notification
		}
	}
	c.JSON(200, spefic)

}
func (Handler *Handler) UpdateNotification(c *gin.Context) {
	token := c.GetHeader("Authorization")
	user, err := Handler.getUser(token)
	if err != nil {
		c.JSON(403, "cant authorization")
		return
	}
	var spefic SpecificationRoom
	spefic.Notification = true
	roomId := c.Param("roomId")
	userId := strconv.Itoa(user.Id)
	newMember := Handler.hub.RoomService.updateRoomSpecification(roomId, userId, spefic)
	if _, ok := Handler.hub.Rooms[roomId]; ok {
		Handler.hub.Rooms[roomId].Members = newMember
	}

	for _, member := range newMember {
		if member.Id == userId {
			spefic.Pin = member.Pin
			spefic.Notification = member.Notification
		}
	}
	c.JSON(200, spefic)
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (Handler *Handler) GetClients(c *gin.Context) {
	var clients []ClientRes
	roomId := c.Param("roomId")
	if _, ok := Handler.hub.Rooms[roomId]; !ok {
		clients = make([]ClientRes, 0)
		c.JSON(http.StatusOK, clients)
	}
	for _, c := range Handler.hub.Rooms[roomId].Clients {
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

type UserRequest struct {
	Id        int       `json:"id"`
	Avatar    string    `json:"avatar"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	UserName  string    `json:"username"`
	Time      time.Time `json:"created_at"`
}

func (Handler *Handler) getUser(token string) (UserRequest, error) {

	var user UserRequest

	if userRequest, ok := Handler.UserHandler[token]; ok {
		return userRequest, nil
	}
	client := &http.Client{}
	gateway := fmt.Sprintf("%s/api/user", "http://dev.oteacher.org")
	request, err := http.NewRequest("GET", gateway, nil)
	request.Header.Set("Authorization", token)
	if err != nil {
		return user, err
	}
	res, err := client.Do(request)
	if res.StatusCode != 200 {
		fmt.Println(res.StatusCode)
		return user, errors.New("error from server")
	}

	if err != nil {
		return user, err
	}

	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}
	Handler.UserHandler[token] = user
	Handler.UpdateUser(UserDto{UserId: strconv.Itoa(user.Id), UserName: user.UserName, FirstName: user.FirstName, LastName: user.LastName, AvatarUrl: user.Avatar})
	return user, nil

}
