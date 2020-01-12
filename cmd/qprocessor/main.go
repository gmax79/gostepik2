package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"project/api"
	"project/systemstop"

	"github.com/streadway/amqp"
)

// Settings - parameters to start
type Settings struct {
	Host     string   `json:"host"`
	RabbitMQ string   `json:"rabbitmq"`
	Graders  []string `json:"graders"`
}

func readSettings() (*Settings, error) {
	settings := &Settings{}
	data, err := ioutil.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(data, &settings)
	}
	return settings, err
}

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Println("Queue processor error:", err)
			os.Exit(1)
		}
	}()

	var settings *Settings
	if settings, err = readSettings(); err != nil {
		return
	}
	var rabbitConn *amqp.Connection
	if rabbitConn, err = amqp.Dial(settings.RabbitMQ); err != nil {
		return
	}
	defer rabbitConn.Close()

	var rabbitChan *amqp.Channel
	if rabbitChan, err = rabbitConn.Channel(); err != nil {
		return
	}
	defer rabbitChan.Close()

	if _, err = rabbitChan.QueueDeclare(
		"grader", // queue name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		return
	}

	rmqchan, err := rabbitChan.Consume(
		"grader", // queue name
		"",       // consumer
		false,    // auto ask
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil)      // arguments
	if err != nil {
		return
	}

	apiserver := &httpAPI{}

	stop := systemstop.Subscribe()
	go func() {
		defer stop.Done()
		time.Sleep(time.Second) // wait start web server
		for {
			select {
			case msg := <-rmqchan:
				if len(msg.Body) == 0 {
					fmt.Println("[QProcessor] empty body from rmq ???")
					msg.Ack(false)
					continue
				}
				mq := &api.MQmessage{}
				if err := json.Unmarshal(msg.Body, mq); err != nil {
					fmt.Printf("[QProcessor] Got invalid blob: %v\n", err)
				} else {
					if err := apiserver.SendTaskToGrader(mq); err != nil {
						fmt.Println(err)
					}
				}
				msg.Ack(false)
			case <-stop.Signal():
				return
			}
		}
	}()

	fmt.Println("Queue processor started at", settings.Host)
	err = apiserver.Serve(settings.Host, settings.Graders)
	systemstop.Wait()
	fmt.Println("Queue processor stopped")
}
