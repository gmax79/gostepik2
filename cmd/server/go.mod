module project/server

go 1.13

replace project/api => ../api

replace project/systemstop => ../systemstop

replace project/storage => ../storage

require (
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	project/api v0.0.0-00010101000000-000000000000
)
