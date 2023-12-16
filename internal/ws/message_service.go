package ws

type MessageService struct {
	MessageRepository RedisRepository
}

func NewMessageService(repository RedisRepository) MessageService {
	return MessageService{
		repository,
	}
}

func (receiver MessageService) SetMessage(roomId string, messageId string, message *Message) bool {
	return receiver.MessageRepository.SetMessage(roomId, messageId, *message).Val()
}

func (receiver MessageService) GetMessage(roomId string) map[string]string {
	return receiver.MessageRepository.GetData(roomId).Val()
}
