package ws

import (
	"github.com/gorilla/websocket"
)

type User struct {
	Conn             *websocket.Conn
	StatusConnection *websocket.Conn
	online           bool
	UserId           string `json:"UserId"`
	rooms            chan *RoomStatus
}

func (User *User) WireRooms(h *Hub) {
	defer func() {
		h.Left <- User
		User.StatusConnection.Close()
	}()
	for {
		select {
		case room, _ := <-User.rooms:
			if User.online == true {
				err := User.StatusConnection.WriteJSON(room)
				if err != nil {

					return
				}
			}

		}

	}
}
