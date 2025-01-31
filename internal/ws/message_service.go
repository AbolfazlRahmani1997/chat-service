package ws

import (
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
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
	receiver.MessageRepository.Mongo.insertMessage(*message)
	return receiver.MessageRepository.Redis.SetMessage(roomId, messageId, *message).Val(), nil
}

// GetMessage Get Data From Redis For Paper ChatRoom Message
func (receiver MessageService) GetMessage(roomId string) []interface{} {
	return receiver.MessageRepository.Redis.GetData(roomId).Val()
}

func (receiver MessageService) MessageDelivery(id string, clientIds []string) {

	_, err := receiver.MessageRepository.MessageDelivery(id, clientIds)
	if err != nil {

		return
	}
}

func (receiver MessageService) MessageRead(id string, clientId string) Message {
	messages := receiver.MessageRepository.getMessageById(id)
	clientIds := messages.Read
	clientIds = append(clientIds, clientId)
	_, err := receiver.MessageRepository.MessageRead(id, clientIds)
	if err != nil {
		return Message{}
	}
	return messages

}

func (r MessageRepository) insertRoomInDb(room Room) *mongo.InsertOneResult {
	return r.Mongo.InsertRoom(room)
}

func (r MessageRepository) getRoom(roomId string) Room {
	return r.GetRoomById(roomId)
}

func (r MessageRepository) updateRoom(roomId string, room Room) {
	r.UpdateRoomById(roomId, room)
}
func (r MessageRepository) getAllMessages(roomId string, userId string) []Message {
	messages := r.Mongo.GetAllMessages(roomId, userId)
	return messages
}
func (r MessageService) getRoomMessages(roomId string, limit string) []Message {
	_, err := strconv.Atoi(limit)
	if err != nil {
		return nil
	}

	messages := r.MessageRepository.Mongo.GetRoomMessage(roomId, limit)
	return messages
}
func (receiver MessageService) getNotDeliverMessage(number int, key string) []Message {
	return receiver.MessageRepository.Redis.GetNotDeliverMessages(number, key)
}
