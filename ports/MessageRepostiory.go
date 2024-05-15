package ports

type MessageRepositoryPort interface {
	GetMessage(id string)
	GetRoomMessages(page int, offset int, roomId string)
}

type MessageServicePort interface {
	GetServiceMessage(id string)
	GetALLMessage(roomId string)
}
