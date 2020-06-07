package vertex

import (
//	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var logger *log.Logger
var pub = make(chan VertexInfo)
var sub = make(chan VertexInfo)

type ControlMsg struct {
	Edge       string
	Vertexno   string
	Vertextype string
	Cmd        string
	Msgid      string
}
var ptest PipeData = PipeData{
	SendTo: "all",
	Datatype: "datatype",
	Data: []byte("data"),
}
var SubVertex = []VertexInfo{}
var PubVertex = []VertexInfo{}


func TransmitToEdge(){
	for {
		select {
			case p := <-pub:
				PubVertex = append(PubVertex, p)
			default:
				if len(PubVertex) == 0 {
					continue
				}
		}
		for _, vi := range PubVertex {
			pi, perr := ReadFromPipe(OUT)
			if perr != nil {
				logger.Printf(
					`Read from pipe failed Trying again...`,
				)
				continue
			}
			logger.Printf("Received from pipe\n")
			serr := SendDataEdge(
				vi,
				pi.SendTo,
				pi.Datatype,
				pi.Data,
			)
			if serr != nil {
				removeVertexInfo(vi, "pub")
				logger.Printf(
					`Send to edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
			}

		}
	}

}

func removeVertexInfo(vi VertexInfo, aname string){
	if len(SubVertex) == 0 || (aname != "sub" && aname != "pub") {
		return
	}
	if aname == "sub" {


	} else {



	}
}

func ListenToEdge() {
	for {
		select {
			case s := <-sub:
				SubVertex = append(SubVertex, s)
			default:
				if len(SubVertex) == 0 {
					continue
				}
		}
		for _, vi := range SubVertex {
			var p PipeData
			p.Datatype, p.Data, err = ReceiveDataEdge(vi, true)
			if err != nil {
				removeVertexInfo(vi, "sub")
				logger.Printf(
					`Receive from edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
			}
			WriteToPipe(IN, p)
			logger.Println("Writing data %v to pipe", p)
		}
	}

}
func logInit() {
	f, err := os.OpenFile("errors.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	logger = log.New(f, "[INFO]", log.LstdFlags)
}

func ListenToController(){
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
func Vamain() {
	logInit()
	go ListenToEdge()
	go TransmitToEdge()
	//LaunchApp("/pkg/app.go")
	//go ListenToController()
	pub1 := InitVertex(1, 3, "pub")
	pub <- pub1
	sub1 := InitVertex(1, 2, "sub")
	sub <- sub1
	for {
		time.Sleep(10*time.Second)
	}
}

func handleConnection(conn net.Conn) {


	//bufferBytes, err := bufis.NewReader(conn).ReadBytes('\n')

	//if err != nil {
	//	logger.Println("client left..")
	//	conn.Close()
	//	return
	//}

	//message := string(bufferBytes)
	//clientAddr := conn.RemoteAddr().String()

	//conn.Write([]byte("you sent: " + response))

}
