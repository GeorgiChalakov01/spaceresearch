module github.com/GeorgiChalakov01/spaceresearch/src/documentprocessor

go 1.24.5

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/GeorgiChalakov01/spaceresearch/backend/app v0.0.0
)

replace github.com/GeorgiChalakov01/spaceresearch/backend/app => ../../backend/app
