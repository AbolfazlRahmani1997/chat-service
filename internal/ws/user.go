package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"strconv"
	"sync"
)

type eventType string

type messagePop struct {
	roomId string
	body   string
}

const (
	roomStatus    eventType = "roomStatus"
	incomeMessage eventType = "incomeMessage"
	listRooms     eventType = "listRooms"
	seenMessage   eventType = "seenMessage"
)

type SystemMessage struct {
	EventType eventType `json:"event_type"`
	Content   interface{}
}

type PupMessage struct {
	MessageId string `json:"message_id,omitempty"`
	RoomId    string `json:"room_id,omitempty"`
	Content   string `json:"content"`
}

type User struct {
	Conn               map[string]*websocket.Conn
	StatusConnection   *websocket.Conn
	online             bool
	UserId             string `json:"UserId"`
	roomStatuses       chan *RoomStatus
	chanelNotification chan *SystemMessage
	pupMessage         chan *PupMessage
	roomList           chan bool
}

func (User *User) WireRooms(h *Hub) {
	defer func() {
		fmt.Println("sendMessage To ")
		h.Left <- User
	}()
	var wg sync.WaitGroup
	for {
		select {
		case roomStatuses, ok := <-User.roomStatuses:

			fmt.Println("send message to Room")
			fmt.Println(roomStatuses.RoomId)
			if ok {
				wg.Add(1)
				go User.writeInAll(&wg)
				User.chanelNotification <- &SystemMessage{EventType: roomStatus, Content: roomStatuses}
				wg.Wait()
			}
		case notification, ok := <-User.pupMessage:
			{
				if ok {
					wg.Add(1)
					go User.writeInAll(&wg)
					User.chanelNotification <- &SystemMessage{EventType: incomeMessage, Content: notification}
					wg.Wait()
				}
			}

		}
	}
}

func (User *User) writeInAll(wg *sync.WaitGroup) {
	defer func() {

		wg.Done()
	}()

	select {
	case sysMessage, ok := <-User.chanelNotification:
		if ok {
			for s, conn := range User.Conn {
				err := conn.WriteJSON(sysMessage)
				if err != nil {
					fmt.Println(err)
					delete(User.Conn, s)
					fmt.Println(len(User.Conn))

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
		delete(User.Conn, connectionId)

		if len(User.Conn) == 0 {
			h.Left <- User
		}

	}()
	var messageClient MessageReceive
	var item string
	var eventRequest eventType
	for {
		if eventRequest != "" {

			switch eventRequest {
			case listRooms:
				_, err := strconv.Atoi(item)
				if err != nil {
					fmt.Println("err")
					break
				}
				err = User.Conn[connectionId].WriteJSON(SystemMessage{EventType: listRooms, Content: h.RoomService.GetMyRoom(User.UserId, item)})
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
