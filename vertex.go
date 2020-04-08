package vertex

import (
	"log"
	"fmt"
	"github.com/streadway/amqp"
)
/*
func main(){
        InitVertex(1, 2, 1)
        for {
		datatype, data, err := ReceiveData(true)
		failOnError(err, "Errors")
		log.Printf("%s\t%s", datatype, data)
	}
}
*/
var queueName string
var Key string
var Host string
var ExchangeName string
var Buffer int = -1
var Vertex int = -1
var Conn* amqp.Connection
var Ch* amqp.Channel
var Q amqp.Queue
var TYPE string = "topic"
var MsgQ <-chan amqp.Delivery

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func InitVertex(buffer int, vertex int, vertexType int) {
	var err error
	Buffer = buffer
	Vertex = vertex
	Host = fmt.Sprintf("amqp://guest:guest@buffer%d:5672/", buffer)
	Conn, err = amqp.Dial(Host)
	failOnError(err, "Failed to Connect to RabbitMQ")
	//defer Conn.Close()

	Ch, err = Conn.Channel()
	failOnError(err, "Failed to open a channel")
	//defer Ch.Close()

	ExchangeName = fmt.Sprintf("E%d", buffer)
	err = Ch.ExchangeDeclare(
		ExchangeName,	// name
		TYPE,		// type
		true,		// durable
		false,		// auto-deleted
		false,		// internal
		false,		// no-wait
		nil,		// arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Subscriber Code
	if vertexType == 1 {
		queueName = fmt.Sprintf("Q%d%d", buffer, vertex)
		Key = fmt.Sprintf("%d.%d", buffer, vertex)
		Q, err = Ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")
		/*
		if(err != nil){
			Q, err = Ch.QueueDeclare(
				queueName,	// name
				false,		// durable
				false,		// delete when unused
				true,		// exclusive
				false,		// no-wait
				nil,		// arguments
			)
			failOnError(err, "Failed to declare a queue")
		}
		*/
		err = Ch.QueueBind(
				Q.Name,		// queue name
				Key,		// routing key
				ExchangeName,	// exchange
				false,
				nil)
		failOnError(err, "Failed to bind a queue")
		cname := fmt.Sprintf("C%d%d", Buffer, Vertex)
		MsgQ, err = Ch.Consume(
			Q.Name,	// queue
			cname,	// consumer
			false,	// auto ack
			false,	// exclusive
			false,	// no local
			false,	// no wait
			nil,	// args
		)
		failOnError(err, "Failed to register a consumer")
	}

}

func SendData(vertex int, datatype string, data []byte) (error){
	if(Buffer == -1 || Vertex == -1){
		return fmt.Errorf("Vertex Uninitialized")
	}
	var key string
	if vertex == 0 {
		key = fmt.Sprintf("%d.*", Buffer)
	} else{
		key = fmt.Sprintf("%d.%d", Buffer, vertex)
	}
	sendQ := fmt.Sprintf("Q%d%d", Buffer, vertex)
	que, err := Ch.QueueInspect(sendQ)
	if que.Name == "lol"{
		log.Printf("lol")
	}
	if err != nil{
		return fmt.Errorf("Destination queue does not exist...")
	}
	err = Ch.Publish(
		ExchangeName,	// exchange
		key,		// routing key
		false,		// mandatory
		false,		// immediate
		amqp.Publishing{
			ContentType:	"application/binary",
			Type:		datatype,
			Body:		data,
		})
	failOnError(err, "Failed to publish a message")
	//log.Printf(" [x] Sent %s", )
	return nil
}


func ReceiveData(ack bool) (string, []byte, error){
	msg := <-MsgQ
	if ack == true{
		msg.Ack(true)
	}
	return msg.Type, msg.Body, nil
}

