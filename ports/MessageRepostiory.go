package ports

type MessageRepositoryPort interface {
	GetMessage(id string)
	GetRoomMessage(roomId string)
}

type MessageServicePort interface {
	GetServiceMessage(id string)
	GetALLMessage(roomId string)
}
