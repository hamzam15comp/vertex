package vertex

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

//VertexInfo struct to hold the state
type VertexInfo struct {
	edge         int
	vertexno     int
	vertexType   string
	queueName    string
	key          string
	host         string
	exchangeName string
	conn         *amqp.Connection
	channel      *amqp.Channel
	q            amqp.Queue
	msgQ         <-chan amqp.Delivery
	cname        string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		return
	}
}

//InitVertex initializes the vertex
func InitVertex(edge int, vertex int, vertexType string) VertexInfo {
	var v VertexInfo
	var err error
	v.vertexno = vertex
	v.vertexType = vertexType
	v.edge = edge

	v.host = fmt.Sprintf("amqp://guest:guest@edge%d:5672/", v.edge)
	v.conn, err = amqp.Dial(v.host)
	failOnError(err, "Failed to Connect to RabbitMQ")

	v.channel, err = v.conn.Channel()
	failOnError(err, "Failed to open a channel")
	v.exchangeName = fmt.Sprintf("E%d", v.edge)
	err = v.channel.ExchangeDeclare(
		v.exchangeName, // name
		"topic",        //type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare an topic exchange")
	err = v.channel.ExchangeDeclare(
		"exfanout", // name
		"fanout",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare an Fanout exchange")

	// Subscriber Code
	if vertexType == "sub" {
		args := amqp.Table {
			"x-max-length-bytes": int(1<<30),
		}
		v.queueName = fmt.Sprintf("Q%d%d", v.edge, v.vertexno)
		v.key = fmt.Sprintf("%d.%d", v.edge, v.vertexno)
		v.q, err = v.channel.QueueDeclare(
			v.queueName,
			true,
			false,
			false,
			false,
			args,
		)
		failOnError(err, "Failed to declare a queue")
		err = v.channel.QueueBind(
			v.q.Name,       // queue name
			v.key,          // routing key
			v.exchangeName, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
		err = v.channel.QueueBind(
			v.q.Name,   // queue name
			v.key,      // routing key
			"exfanout", // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
		v.cname = fmt.Sprintf("C%d%d", v.edge, v.vertexno)
		v.msgQ, err = v.channel.Consume(
			v.q.Name, // queue
			v.cname,  // consumer
			false,    // auto ack
			false,    // exclusive
			false,    // no local
			false,    // no wait
			nil,      // args
		)
		failOnError(err, "Failed to register a consumer")
	}
	return v
}

//SendDataEdge sends data
func SendDataEdge(
	v VertexInfo,
	sendTo string,
	datatype string,
	data []byte,
) error {
	var err error = nil
	if v.edge == -1 || v.vertexno == -1 {
		return fmt.Errorf("Vertex Uninitialized")
	}
	key := fmt.Sprintf("%d.%s", v.edge, sendTo)
	if sendTo == "all" {
		err = v.channel.Publish(
			"exfanout", // exchange
			v.q.Name,   // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "application/binary",
				Type:        datatype,
				Body:        data,
			})
		failOnError(err, "Failed to publish a message")
	} else {
		err = v.channel.Publish(
			v.exchangeName, // exchange
			key,            // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "application/binary",
				Type:        datatype,
				Body:        data,
			})
		failOnError(err, "Failed to publish a message")
	}
	return err
}

//ReceiveDataEdge receives data
func ReceiveDataEdge(v VertexInfo, ack bool) (string, []byte, error) {
	msg := <-v.msgQ
	msg.Ack(ack)
	return msg.Type, msg.Body, nil
}

/*
func main() {
	pub1 := InitVertex(1, 1, "pub")
	fmt.Printf("%v\n", pub1)
	sub1 := InitVertex(1, 2, "sub")
	fmt.Printf("%v\n", sub1)
	err := SendDataEdge(pub1, "all", "datatype", []byte("data"))
	fmt.Println(err)
	datatype, data, err := ReceiveDataEdge(sub1, true)
	fmt.Println(datatype, string(data), err)
	// for {
	// 	datatype, data, err := ReceiveDataEdge(true)
	// 	failOnError(err, "Errors")
	// 	log.Printf("%s\t%s", datatype, data)
	// }
}
*/
