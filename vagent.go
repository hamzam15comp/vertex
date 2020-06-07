package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/hamzam/vertex"
	"github.com/streadway/amqp"
)

var logger *log.Logger

type controlMsg struct {
	edge       string
	vertexno   string
	vertextype string
	cmd        string
	msgid      string
}

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

var subVertex = []vertexInfo{}
var pubVertex = []vertexInfo{}

func listenToEdge() {
	for {
		if len(subVertex) == 0 {
			continue
		}
		for vi := range subVertex {
			var p pipe.pipeData
			p.datatype, p.data, err = ReceiveData(vi, true)
			if err != nil {
				removeVertexInfo(vi, subVertex)
				logger.Printf(
					`Receive from edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
			}
			//sendtoPipe(p)
			logger.Println(p)
		}
	}

}

func main() {
	f, err := os.OpenFile("errors.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Println(err)
	}
	defer f.Close()

	logger = log.New(f, "[INFO]", log.LstdFlags)
	listener, err := net.Listen("tcp", "0.0.0.0:7000")
	if err != nil {
		logger.Fatal("tcp server listener error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Fatal("tcp server accept error", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	bufferBytes, err := bufio.NewReader(conn).ReadBytes('\n')

	if err != nil {
		logger.Println("client left..")
		conn.Close()
		return
	}

	message := string(bufferBytes)
	clientAddr := conn.RemoteAddr().String()
	response := fmt.Sprintf(message + " from " + clientAddr + "\n")

	logger.Println(response)

	conn.Write([]byte("you sent: " + response))

	handleConnection(conn)
}
