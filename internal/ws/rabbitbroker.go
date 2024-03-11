package ws

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type RabbitMqBroker struct {
	Room            chan *Room
	connection      amqp.Connection
	MongoRepository MessageRepository
}

func NewRabbitMqBroker(Room chan *Room, repository MessageRepository) RabbitMqBroker {

	RabbitMqUrl := fmt.Sprintf("amqp://user:5KagXebV8DXNeQWa@rabbit-dev-rabbitmq.rabbit-dev:5672/")
	fmt.Println(RabbitMqUrl)
	conn, err := amqp.Dial(RabbitMqUrl)
	if err != nil {
		fmt.Println(err)
	}
	return RabbitMqBroker{connection: *conn,
		MongoRepository: repository,
		Room:            Room}
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
			receiver.MongoRepository.insertRoomInDb(Room{
				ID:      RoomRequest.Id,
				Name:    RoomRequest.Name,
				Members: RoomRequest.Member,
			})

		}
		wg.Wait()

	}()

}
