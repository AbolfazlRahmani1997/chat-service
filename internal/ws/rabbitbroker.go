package ws

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"sync"
)

type RabbitMqBroker struct {
	Room            chan *Room
	connection      amqp.Connection
	MongoRepository MessageRepository
}

func NewRabbitMqBroker(Room chan *Room, repository MessageRepository) RabbitMqBroker {

	RabbitMqUrl := fmt.Sprintf("amqp://%s:%s@%s:5672/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"))

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

	mes, err := ch.Consume("chat-service-room", "", true, false, false, false, nil)

	if err != nil {
		fmt.Println(err)
	}

	var wg sync.WaitGroup
	go func() {
		defer wg.Done()
		wg.Add(1)
		for d := range mes {
			var RoomRequest RoomBrokerDto
			err = json.Unmarshal(d.Body, &RoomRequest)
			if err != nil {
				fmt.Println(err)
			}
			room := Room{
				ID:      RoomRequest.Id,
				Name:    RoomRequest.Name,
				Members: RoomRequest.Member,
				Type:    RoomRequest.Type,
			}
			receiver.MongoRepository.insertRoomInDb(room)
			receiver.Room <- &room

		}
		wg.Wait()

	}()

}
