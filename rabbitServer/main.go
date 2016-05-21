package main

import (
	"fmt"
	"github.com/streadway/amqp"
)

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"logs", // name
		false,  // durable
		false,  // delete when usused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			receiveMessage(d.Body)
		}
	}()

	fmt.Printf(" [*] Waiting for messages. To exit press CTRL+C\n\n")
	<-forever

}

func receiveMessage(msg []byte) {
	fmt.Printf("msg: %s\n", msg)
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Sprintf("%s: %s\n", msg, err)
		panic(fmt.Sprintf("%s: %s\n", msg, err))
	}
}
