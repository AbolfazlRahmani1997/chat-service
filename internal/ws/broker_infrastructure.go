package ws

type RoomBrokerInfrastructure interface {
	Consume()
}

type RoomBrokerDto struct {
	Id     string   `json:"Id"`
	Name   string   `json:"Name"`
	Member []Member `json:"Member"`
	Writer []string `json:"Writer"`
	Owner  []string `json:"Owner"`
}
