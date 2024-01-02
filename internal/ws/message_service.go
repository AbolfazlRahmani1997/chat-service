package ws

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageService struct {
	MessageRepository
}

// NewMessageService build Service Usage In Hub
func NewMessageService(repository MessageRepository) MessageService {
	return MessageService{
		repository,
	}
}

// SetMessage Insert Data in Redis And MongoDb (stateless data ,stateFull Data )
func (receiver MessageService) SetMessage(roomId string, messageId string, message *Message) (bool, error) {
	receiver.MessageRepository.MongoDBRepository.InsertMessage(*message)
	return receiver.MessageRepository.SetMessage(roomId, messageId, *message).Val(), nil
}

// GetMessage Get Data From Redis For Paper ChatRoom Message
func (receiver MessageService) GetMessage(roomId string) []interface{} {
	return receiver.MessageRepository.GetData(roomId).Val()
}

// InsertInDb Insert In Db For StateFull
func (r MessageRepository) insertMessageInDb(message Message) *mongo.InsertOneResult {
	return r.InsertMessage(message)
}

func (r MessageRepository) insertRoomInDb(room Room) *mongo.InsertOneResult {
	return r.InsertRoom(room)
}

func (r MessageRepository) getRoom(roomId string) Room {
	return r.GetRoomById(roomId)
}
