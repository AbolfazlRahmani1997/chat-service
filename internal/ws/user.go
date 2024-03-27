package ws

import (
	"encoding/json"
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

type MessageReceive struct {
	Page string `json:"Page,omitempty"`
}

func (User *User) userConnection(h *Hub) {
	defer func() {
		h.Left <- User
		User.Conn.Close()
	}()
	var messageClient MessageReceive
	var page string
	for {

		if page != "" {
			err := User.Conn.WriteJSON(h.RoomService.GetMyRoom(User.UserId, page))
			if err != nil {
				fmt.Println("Error reading message:", err)
				break
			}
		}

		var message []byte
		_, message, err := User.Conn.ReadMessage()
		if err != nil {
			break

		}
		if len(message) > 0 {
			err = json.Unmarshal(message, &messageClient)
			if err != nil {
				fmt.Println(err)
			}
			page = messageClient.Page
		} else {
			page = ""
		}

	}

}
