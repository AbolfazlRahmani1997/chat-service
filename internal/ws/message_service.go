package ws

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageService struct {
	MessageRepository MessageRepository
}

// NewMessageService build Service Usage In Hub
func NewMessageService(repository MessageRepository) MessageService {
	return MessageService{
		repository,
	}
}

// SetMessage Insert Data in Redis And MongoDb (stateless data ,stateFull Data )
func (receiver MessageService) SetMessage(roomId string, messageId string, message *Message) (bool, error) {
	receiver.MessageRepository.MongoDBRepository.insertMessage(*message)
	return receiver.MessageRepository.Redis.SetMessage(roomId, messageId, *message).Val(), nil
}

// GetMessage Get Data From Redis For Paper ChatRoom Message
func (receiver MessageService) GetMessage(roomId string) []interface{} {
	return receiver.MessageRepository.Redis.GetData(roomId).Val()
}

func (receiver MessageService) MessageDelivery(id string, clientIds []string) {

	_, err := receiver.MessageRepository.MessageDelivery(id, clientIds)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func (r MessageRepository) insertRoomInDb(room Room) *mongo.InsertOneResult {
	return r.InsertRoom(room)
}

func (r MessageRepository) getRoom(roomId string) Room {
	return r.GetRoomById(roomId)
}
func (r MessageRepository) getAllMessages(roomId string) []Message {
	messages := r.GetAllMessages(roomId)
	return messages
}
func (receiver MessageService) getNotDeliverMessage(number int, key string) []Message {
	return receiver.MessageRepository.Redis.GetNotDeliverMessages(number, key)
}
