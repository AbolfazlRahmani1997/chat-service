package ws

import (
	"github.com/gorilla/websocket"
)

type User struct {
	Conn             *websocket.Conn
	StatusConnection *websocket.Conn
	UserId           string `json:"UserId"`
	rooms            chan *RoomStatus
}

func (User *User) WireRooms(h *Hub) {
	defer func() {
		h.Left <- User
		User.Conn.Close()
	}()
	for {
		select {
		case room, _ := <-User.rooms:
			err := User.StatusConnection.WriteJSON(room)
			if err != nil {
				return
			}

		}

	}
}
