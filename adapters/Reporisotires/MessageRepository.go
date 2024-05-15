package Reporisotires

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"server/entity/Message"
)

type MessageRepository struct {
	Collection *mongo.Database
}

func (r *MessageRepository) GetMessage(id string) Message.Message {
	var message Message.Message
	condition := bson.M{"id": id}
	find := r.Collection.Collection("messages").FindOne(context.TODO(), condition)
	err := find.Decode(&message)
	if err != nil {
		fmt.Println(err)
		return Message.Message{}
	}
	return message
}

func (r *MessageRepository) GetRoomMessages(page int, offset int, roomId string) []Message.Message {
	var messages []Message.Message
	find, err := r.Collection.Collection("messages").Find(context.TODO(), filter.GetFilter())
	err = find.All(context.TODO(), &messages)
	if err != nil {
		return []Message.Message{}
	}
	if err != nil {
		fmt.Println(err)
		return []Message.Message{}
	}
	if err != nil {
		fmt.Println(err)

		return []Message.Message{}
	}
	return messages
}
