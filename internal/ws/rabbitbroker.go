package ws

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type RabbitMqBroker struct {
	Room       chan *Room
	connection amqp.Connection
}

func NewRabbitMqBroker(Room chan *Room) RabbitMqBroker {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.48.4:5672/")
	if err != nil {
		fmt.Println(err)
	}
	return RabbitMqBroker{connection: *conn, Room: Room}
}
func (receiver *RabbitMqBroker) Consume() {
	ch, _ := receiver.connection.Channel()

	mesg, err := ch.Consume("chat-service-room", "", true, false, false, false, nil)

	if err != nil {
		fmt.Println(err)
	}

	var wg sync.WaitGroup
	go func() {

		defer wg.Done()
		wg.Add(1)
		for d := range mesg {
			var RoomRequest RoomBrokerDto
			err = json.Unmarshal(d.Body, &RoomRequest)
			if err != nil {
				fmt.Println(err)
			}

			receiver.Room <- &Room{
				ID:      RoomRequest.Id,
				Name:    RoomRequest.Name,
				Owner:   RoomRequest.Owner,
				Writer:  RoomRequest.Writer,
				Members: RoomRequest.Member,
			}

		}
		wg.Wait()

	}()

}
