package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Message struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Content      string             `json:"Content,omitempty"  bson:"content"`
	RoomID       string             `json:"RoomID,omitempty"  bson:"roomID"`
	Username     string             `json:"Username,omitempty" bson:"username" `
	ClientID     string             `json:"ClientID,omitempty" bson:"clientID"`
	Deliver      []string           `json:"Deliver,omitempty" bson:"Deliver"`
	Read         []string           `json:"Read,omitempty" bson:"Read"`
	connectionId string
	CreatedAt    time.Time `json:"CreatedAt" bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}
