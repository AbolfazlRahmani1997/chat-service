package Reporisotires

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"server/Dtos/DtoReporitories"
	"server/entity/Message"
)

type MessageRepository struct {
	Collection *mongo.Database
}

func NewMessageRepository(client *mongo.Client) *MessageRepository {
	return &MessageRepository{client.Database("MessageDB")}
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

func (r *MessageRepository) GetRoomMessages(page int, offset int, filter DtoReporitories.MessageFilterDto) []Message.Message {
	var messages []Message.Message
	l := offset
	skip := int64((page * offset) - offset)
	var query []bson.M
	query = r.queryBuilding(filter)
	skipStage := bson.M{"$skip": skip}
	limitStage := bson.M{"$limit": l}
	query = append(query, skipStage)
	query = append(query, limitStage)
	find, err := r.Collection.Collection("messages").Find(context.TODO(), query)
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

func (r *MessageRepository) queryBuilding(filter DtoReporitories.MessageFilterDto) []bson.M {
	var query []bson.M
	if filter.RoomId != "" {
		query = append(query, bson.M{"roomId": filter.RoomId})
	}
	if filter.ClientId != "" {
		query = append(query, bson.M{"clientId": filter.ClientId})
	}

	return query
}
