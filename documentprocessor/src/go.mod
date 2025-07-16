module github.com/GeorgiChalakov01/spaceresearch/src/documentprocessor

go 1.24.5

require (
	github.com/GeorgiChalakov01/spaceresearch/backend/app v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

replace github.com/GeorgiChalakov01/spaceresearch/backend/app => ../../backend/app
