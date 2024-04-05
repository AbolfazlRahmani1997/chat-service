package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type eventType string

type messagePop struct {
	roomId string
	body   string
}

const (
	roomStatus    eventType = "roomStatus"
	incomeMessage eventType = "getMessage"
	listRooms     eventType = "listRooms"
	seenMessage   eventType = "seenMessage"
)

type SystemMessage struct {
	EventType eventType `json:"event_type"`
	Content   interface{}
}

type User struct {
	Conn             map[string]*websocket.Conn
	StatusConnection *websocket.Conn
	online           bool
	UserId           string `json:"UserId"`
	roomStatuses     chan *RoomStatus
	chanelMessage    chan *SystemMessage
	roomList         chan bool
}

func (User *User) WireRooms(h *Hub) {
	defer func() {
		h.Evade <- User
	}()
	var wg sync.WaitGroup
	for {
		select {
		case roomStatuses, ok := <-User.roomStatuses:

			if ok {
				wg.Add(1)
				go User.writeInAll(&wg)
				User.chanelMessage <- &SystemMessage{EventType: roomStatus, Content: roomStatuses}
				wg.Wait()
			}
		}
	}
}

func (User *User) writeInAll(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	select {
	case sysMessage, ok := <-User.chanelMessage:
		fmt.Println("Test")
		fmt.Println(sysMessage.EventType)
		if ok {

			for s, conn := range User.Conn {
				err := conn.WriteJSON(sysMessage)
				if err != nil {
					err := conn.Close()
					if err != nil {
						return
					}
					delete(User.Conn, s)

					return
				}
			}
		}
	}

}

type MessageReceive struct {
	RequestType eventType `json:"event_type"`
	Item        string    `json:"item,omitempty"`
}

func (User *User) userConnection(h *Hub, connectionId string) {
	defer func() {
		User.Conn[connectionId].Close()
	}()
	var messageClient MessageReceive
	var item string
	var eventRequest eventType
	for {
		if eventRequest != "" {

			switch eventRequest {
			case listRooms:
				err := User.Conn[connectionId].WriteJSON(SystemMessage{EventType: listRooms, Content: h.RoomService.GetMyRoom(User.UserId, item)})
				if err != nil {
					fmt.Println("error reading message:", err)
					break
				}
			case seenMessage:
				h.ReadAble <- &ReadMessage{MessageId: item, UserId: User.UserId}
			}

		}

		var message []byte
		_, message, err := User.Conn[connectionId].ReadMessage()
		if err != nil {
			break

		}
		if len(message) > 0 {
			err = json.Unmarshal(message, &messageClient)
			if err != nil {
				fmt.Println(err)
				break
			}
			eventRequest = messageClient.RequestType
			item = messageClient.Item
		} else {
			item = ""
		}

	}

}
