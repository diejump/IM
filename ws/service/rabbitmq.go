package service

import (
	"fmt"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"time"
)

var RabbitMQChannel *amqp.Channel
var rabbitMQConn *amqp.Connection
var Queue amqp.Queue

func InitRabbitMQ(host, port string) {
	rabbitMQConn, _ = amqp.Dial("amqp://guest:guest@" + host + ":" + port + "/")
	// 创建RabbitMQ通道
	RabbitMQChannel, _ = rabbitMQConn.Channel()
	Queue, _ = RabbitMQChannel.QueueDeclare( //声明一个队列
		"chatMQ",
		false,
		false,
		false,
		false,
		nil,
	)

}

func SendMessageToMQ(message []byte) {
	err := RabbitMQChannel.Publish(
		"",
		Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message, //数据库信息
		},
	)
	if err != nil {
		fmt.Println(err)
	}
}

func ConsumeMessage() <-chan amqp.Delivery {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := RabbitMQChannel.Consume(
		Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return msgs
}

func MQClose() {
	RabbitMQChannel.Close()
	rabbitMQConn.Close()
}
