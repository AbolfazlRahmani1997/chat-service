package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

type status string

const (
	online   = "online"
	offline  = "offline"
	isTyping = "isTyping"
)

type Client struct {
	Conn        *websocket.Conn
	ReadMessage *websocket.Conn
	Message     chan *Message
	ID          string `json:"id"`
	RoomID      string `json:"roomId"`
	Username    string `json:"username"`
	Status      string `json:"status"`
}

type Message struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Content   string             `json:"Content,omitempty"  bson:"content"`
	RoomID    string             `json:"RoomID,omitempty"  bson:"roomID"`
	Username  string             `json:"Username,omitempty" bson:"username" `
	ClientID  string             `json:"ClientID,omitempty" bson:"clientID"`
	Deliver   []string           `json:"Deliver,omitempty" bson:"Deliver"`
	Read      []string           `json:"Read,omitempty" bson:"Read"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (c *Client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		err := c.Conn.WriteJSON(message)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg := &Message{
			Content:  string(m),
			RoomID:   c.RoomID,
			Username: c.Username,
			ClientID: c.ID,
		}
		hub.Broadcast <- msg
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
