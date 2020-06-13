package vertex

import (
//	"bufio"
	"fmt"
	"encoding/json"
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
var testcmsg ControlMsg = ControlMsg{
	Edge: "edge1",
	Vertexno: "1",
	Vertextype: "pub",
	Cmd:	"add",
	Msgid: "124",
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
		for i, vi := range PubVertex {
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
				fmt.Println(i)
				//removeVertexInfo(i, PubVertex)
				logger.Printf(
					`Send to edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
			}

		}
	}

}

func removeVertexInfo(vi int, vertexSlice []byte) ([]byte){ //[]VertexInfo){
	vlen := len(vertexSlice)
	if vlen == 0 {
		return
	}
	vertexSlice[vi] = vertexSlice[vlen-1]
	vertexSlice[vlen-1] = 0 
	vertexSlice = vertexSlice[:vlen-1]
	return vertexSlice
}


func FindInSlice(slice []VertexInfo, val VertexInfo) (int, bool) {
        for i, item := range slice {
                if item == val {
                        return i, true
                }
        }
        return -1, false
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
		for i, vi := range SubVertex {
			var p PipeData
			var err error
			p.Datatype, p.Data, err = ReceiveDataEdge(vi, true)
			if err != nil {
				fmt.Println(i)
				//removeVertexInfo(i, SubVertex)
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
		logger.Println("tcp server listener error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Println("tcp server accept error", err)
		}

		go handleController(conn)
	}
}

var exslice = []byte{'a', 'b', 'c', 'd', 'e'}

func Vamain() {
	logInit()
	go ListenToEdge()
	go TransmitToEdge()
	//LaunchApp("/pkg/app.go")
	go ListenToController()
	for {
		time.Sleep(10*time.Second)
		SendToVagent(testcmsg)
		fmt.Println(exslice)
		exslice = removeVertexInfo(2, exslice)
		fmt.Println(exslice)
	}
}
func SendToVagent(cmsg ControlMsg) {
	host := "localhost:7000"
	con, err := net.Dial("tcp", host)
	if err != nil {
		logger.Println(err)
	}
	enc := json.NewEncoder(con)
	err = enc.Encode(cmsg)
	if err != nil {
		logger.Println(err)
	}
	defer con.Close()
}
func addConnection(cmsg ControlMsg) {

}

func delConnection(cmsg ControlMsg){

}
func handleController(conn net.Conn) {
	dec := json.NewDecoder(conn)
	var cmsg ControlMsg
	err := dec.Decode(&cmsg)
	if err != nil {
		logger.Println(err)
	}
	fmt.Println(cmsg)
	defer conn.Close()

}
