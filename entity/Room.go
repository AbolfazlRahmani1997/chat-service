package entity

import "server/enums"

type Room struct {
	ID        string       `json:"Id" `
	Name      string       `json:"name" `
	Temporary bool         `json:"type" `
	Members   []Member     `json:"members"`
	Message   Message      `json:"message" bson:"last_message"`
	Status    enums.Status `json:"status,omitempty" `
}

type Member struct {
	Id        string   `json:"Id"`
	Roles     []string `json:"roles"`
	FirstName string   `json:"firstname"`
	LastName  string   `json:"lastname"`
	AvatarUrl string   `json:"AvatarUrl"bson:"avatar_url"`
}
