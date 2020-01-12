module grader/qprocessor

go 1.13

require (
	github.com/google/uuid v1.1.1
	github.com/satori/go.uuid v1.2.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	project/api v0.0.0-00010101000000-000000000000
	project/systemstop v0.0.0-00010101000000-000000000000
)

replace project/api => ../api

replace project/systemstop => ../systemstop
