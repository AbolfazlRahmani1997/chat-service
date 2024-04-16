package ports

type MessageRepositoryPort interface {
	GetMessage(id string)
	InsertMessage()
	UpdateMessage()
}
