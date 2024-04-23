package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

type status string

const (
	online  status = "online"
	offline status = "offline"
	evade   status = "evade"
)

type Client struct {
	Conn          map[string]*websocket.Conn
	ReadMessage   *websocket.Conn
	Message       chan *Message
	ChanelMessage chan *Message
	ID            string `json:"id"`
	RoomID        string `json:"roomId"`
	Username      string `json:"username"`
	Status        status `json:"status"`
}

type Message struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Content      string             `json:"Content,omitempty"  bson:"content"`
	UniqId       string             `json:"UniqId"`
	RoomID       string             `json:"RoomID,omitempty"  bson:"roomID"`
	Username     string             `json:"Username,omitempty" bson:"username" `
	ClientID     string             `json:"ClientID,omitempty" bson:"clientID"`
	Deliver      []string           `json:"Deliver" bson:"Deliver"`
	Read         []string           `json:"Read" bson:"Read"`
	connectionId string
	CreatedAt    time.Time `json:"CreatedAt"bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func (c *Client) writeMessage() {
	defer func() {

	}()

	for {
		message, ok := <-c.Message
		c.writeInAll(message)
		if !ok {
			return
		}

	}
}

// write in all connection
func (c *Client) writeInAll(m *Message) {
	for i, conn := range c.Conn {
		if m.connectionId != i {
			err := conn.WriteJSON(m)
			if err != nil {

				break

			}
		}

	}
}

type messageClient struct {
	Ulid    string `json:"ulid"`
	Content string `json:"content"`
}

func (c *Client) readerMessage(index string, hub *Hub) {
	defer func() {
		err := c.Conn[index].Close()
		if err != nil {
			return
		}
		delete(c.Conn, index)
		fmt.Println("closed \t" + index)
		if len(c.Conn) == 0 {
			fmt.Println("Unregister \t" + index)
			hub.Unregister <- c
		}
	}()
	var messageDeliverClient messageClient
	for {
		_, message, err := c.Conn[index].ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		err = json.Unmarshal(message, &messageDeliverClient)
		if err != nil {
			fmt.Println(err)
			return
		}
		if messageDeliverClient.Ulid != "ping" {
			msg := &Message{
				Content:      messageDeliverClient.Content,
				UniqId:       messageDeliverClient.Ulid,
				connectionId: index,
				RoomID:       c.RoomID,
				Username:     c.Username,
				ClientID:     c.ID,
			}
			c.ChanelMessage <- msg
			systemMessage := SystemMessage{EventType: deliverMessage, Content: messageDeliverClient.Ulid}
			err = c.Conn[index].WriteJSON(systemMessage)

		} else {
			msg := &Message{
				Content:      messageDeliverClient.Content,
				UniqId:       messageDeliverClient.Ulid,
				connectionId: index,
				RoomID:       c.RoomID,
				Username:     c.Username,
				ClientID:     c.ID,
			}
			c.ChanelMessage <- msg
			systemMessage := SystemMessage{EventType: deliverMessage, Content: messageDeliverClient.Ulid}
			err = c.Conn[index].WriteJSON(systemMessage)
		}
		if err != nil {

			break
		}

	}
}

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
	}()

	for {
		m, ok := <-c.ChanelMessage
		if !ok {
			break
		}
		hub.Broadcast <- m
	}
}

func (c *Client) seenMessage(hub *Hub) {
	defer func() {
		c.ReadMessage.Close()
	}()

	for {
		_, m, err := c.ReadMessage.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg := &ReadMessage{
			MessageId: string(m),
			UserId:    c.ID,
		}

		hub.ReadAble <- msg
	}
}
