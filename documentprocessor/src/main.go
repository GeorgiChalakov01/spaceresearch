package main

import (
	"fmt"

	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/rabbitmq"
)


func main() {
	// Connect to RabbitMQ
	conn, err := rabbitmq.Connect()
	if err != nil {
		fmt.Println("Could not connect to RabbitMQ. Error:\n%v", err)
		return
	}
	defer conn.Close()

	// Create a RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("\nCould not create a RabbitMQ channel. Error:\n%v\n", err)
		return
	}
	defer ch.Close()

	// Create a RabbitMQ queue
	q, err := ch.QueueDeclare(
		"documents",	// name
		true,		// durable
		false,		// delete when unused
		false,		// exclusive
		false,		// no-wait
		nil,		// arguments
	)
	if err != nil {
		fmt.Printf("\nCould not create a RabbitMQ queue. Error:\n%v\n", err)
		return
	}


	msgs, err := ch.Consume(
		q.Name,		// queue
		"",		// consumer
		true,		// auto-ack
		false,		// exclusive
		false,		// no-local
		false,		// no-wait
		nil,		// args
	)
	if err != nil {
		fmt.Printf("\nFailed to register a consumer. Error:\n%v\n", err)
		return
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			fmt.Printf("\nReceived a message: %s\n", d.Body)
		}
	}()

	fmt.Printf("\n [*] Waiting for messages. To exit press CTRL+C\n")
	<-forever
}
