package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
)

type User struct {
	Conn             *websocket.Conn
	StatusConnection *websocket.Conn
	online           bool
	UserId           string `json:"UserId"`
	roomStatuses     chan *RoomStatus
	roomList         chan bool
}

func (User *User) WireRooms(h *Hub) {
	defer func() {
		User.StatusConnection.Close()
	}()
	for {
		select {
		case roomStatuses, _ := <-User.roomStatuses:
			if User.online == true {
				err := User.StatusConnection.WriteJSON(roomStatuses)
				if err != nil {
				}
			} else {
				_, _, err := User.StatusConnection.ReadMessage()
				if err != nil {
					break
				}

			}

		}
	}
}

func (User *User) userConnection(h *Hub) {
	defer func() {
		h.Left <- User
		User.Conn.Close()
	}()

	for {
		err := User.Conn.WriteJSON(h.RoomService.GetMyRoom(User.UserId))
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		_, _, err = User.Conn.ReadMessage()
		if err != nil {
			break

		}
	}

}
