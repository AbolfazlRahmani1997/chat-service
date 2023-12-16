package ws

type MessageService struct {
	MessageRepository RedisRepository
}

func NewMessageService(repository RedisRepository) MessageService {
	return MessageService{
		repository,
	}
}
