package vertex

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

//vertexInfo struct to hold the state
type vertexInfo struct {
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
	dtype        string
	msgQ         <-chan amqp.Delivery
	cname        string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//InitVertex initializes the vertex
func InitVertex(edge int, vertex int, vertexType string) vertexInfo {
	var v vertexInfo
	var err error
	v.dtype = "fanout"
	v.vertexno = vertex
	v.edge = -1
	v.vertexType = vertexType
	v.edge = edge
	v.vertexno = vertex
	v.host = fmt.Sprintf("amqp://guest:guest@edge%d:5672/", v.edge)
	v.conn, err = amqp.Dial(v.host)
	failOnError(err, "Failed to Connect to RabbitMQ")
	//defer Conn.Close()

	v.channel, err = v.conn.Channel()
	failOnError(err, "Failed to open a channel")
	//defer Ch.Close()

	v.exchangeName = fmt.Sprintf("E%d", v.edge)
	err = v.channel.ExchangeDeclare(
		v.exchangeName, // name
		v.dtype,        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Subscriber Code
	if vertexType == "sub" {
		v.queueName = fmt.Sprintf("Q%d%d", v.edge, v.vertexno)
		v.key = fmt.Sprintf("%d.%d", v.edge, v.vertexno)
		v.q, err = v.channel.QueueDeclare(
			v.queueName,
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
		err = v.channel.QueueBind(
			v.q.Name,       // queue name
			v.key,          // routing key
			v.exchangeName, // exchange
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
func SendDataEdge(v vertexInfo, datatype string, data []byte) error {
	var err error
	if v.edge == -1 || v.vertexno == -1 {
		return fmt.Errorf("Vertex Uninitialized")
	}
	//var key string
	// if v.vertexno == 0 {
	// 	v.key = fmt.Sprintf("#.#")
	// } else {
	// 	v.key = fmt.Sprintf("%d.%d", v.edge, v.vertexno)
	// 	sendQ := fmt.Sprintf("Q%d%d", v.edge, v.vertexno)
	// 	_, err := v.channel.QueueInspect(sendQ)
	// 	if err != nil {
	// 		return fmt.Errorf("Destination queue does not exist")
	// 	}
	// }
	err = v.channel.Publish(
		v.exchangeName, // exchange
		v.q.Name,       // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/binary",
			Type:        datatype,
			Body:        data,
		})
	failOnError(err, "Failed to publish a message")
	//log.Printf(" [x] Sent %s", )
	return nil
}

//ReceiveDataEdge receives data
func ReceiveDataEdge(v vertexInfo, ack bool) (string, []byte, error) {
	msg := <-v.msgQ
	if ack == true {
		msg.Ack(true)
	}
	return msg.Type, msg.Body, nil
}

/*
func main() {
	pub1 := InitVertex(1, 1, "pub")
	fmt.Printf("%v\n", pub1)
	sub1 := InitVertex(1, 2, "sub")
	fmt.Printf("%v\n", sub1)
	err := SendDataEdge(pub1, "datatype", []byte("data"))
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
