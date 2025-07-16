package rabbitmq

import (
	"os"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect() (*amqp.Connection, error) {
	user := os.Getenv("RABBITMQ_DEFAULT_USER")
	password := os.Getenv("RABBITMQ_DEFAULT_PASS")
	port := os.Getenv("RABBITMQ_PORT")
	containerName := os.Getenv("RABBITMQ_CONTAINER_NAME")
	endpoint := containerName + ":" + port

	connectionString := "amqp://" + user + ":" + password + "@" + endpoint + "/"

	fmt.Printf("\nAttempting to connect to %s", connectionString)

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
