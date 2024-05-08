package Message

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"server/transform/api"
	"server/transform/api/message"
	"time"
)

type Message struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Content        string             `json:"Content,omitempty"  bson:"content"`
	RoomID         string             `json:"RoomID,omitempty"  bson:"roomID"`
	Username       string             `json:"Username,omitempty" bson:"username" `
	ClientID       string             `json:"ClientID,omitempty" bson:"clientID"`
	Deliver        []string           `json:"Deliver,omitempty" bson:"Deliver"`
	Read           []string           `json:"Read,omitempty" bson:"Read"`
	connectionId   string
	CreatedAt      time.Time `json:"CreatedAt" bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	collectionName string
}

func (Message *Message) init() {
	Message.collectionName = os.Getenv("MESSAGE_COLLECTION")
}
func (Message *Message) ToTransformer() api.Transformer {
	return &message.Message{}

}

func (Message *Message) GetTableName() string {
	return Message.collectionName
}

func (Message *Message) ToCollectionTransformer() api.Transformer {
	return &message.Message{}

}
