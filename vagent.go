package vertex

import (
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
	Edge       int
	Vertexno   int
	Vertextype string
	Cmd        string
	Msgid      int
}

var pubadd11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "pub",
	Cmd:	"add",
	Msgid: 124,
}
var subadd11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "sub",
	Cmd:	"add",
	Msgid: 124,
}
var pubrem11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "pub",
	Cmd:	"rem",
	Msgid: 124,
}
var subrem11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "sub",
	Cmd:	"rem",
	Msgid: 124,
}
var ptest PipeData = PipeData{
	SendTo: "all",
	Datatype: "datatype",
	Data: []byte("data"),
}
var SubVertex = []VertexInfo{}
var PubVertex = []VertexInfo{}
var stopp = make(chan time.Duration)
var donep = make(chan bool)
var stops = make(chan time.Duration)
var stopg = make(chan time.Duration)
var dones = make(chan bool)
var doneg = make(chan bool)

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
			select {
				case s := <-stopp:
					for {
						logger.Println(
						"Waiting to Publish",
						)
						time.Sleep(s*time.Second)
						if <-donep {
							break
						}
					}
					continue
				default: {
					break
				}
			}
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
				PubVertex = removeVertexInfo(i, PubVertex)
				logger.Printf(
					`Send to edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
				break
			}
			fmt.Println("Sending to Edge", pi)
			time.Sleep(10*time.Second)

		}
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
		for i, vi := range SubVertex {
			select {
				case s := <-stops:
					for {
						logger.Println(
						"Waiting to Subscribe",
						)
						time.Sleep(s*time.Second)
						if <-dones {
							break
						}
					}
					continue
				default: {
					break
				}
			}
			var p PipeData
			var err error
			p.Datatype, p.Data, err = ReceiveDataEdge(vi, true)
			if err != nil {
				SubVertex = removeVertexInfo(i, SubVertex)
				logger.Printf(
					`Receive from edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
				break
			}
			WriteToPipe(IN, p)
			fmt.Println("Received from Edge", p)
			time.Sleep(10*time.Second)
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


func Vamain() {
	logInit()
	go ListenToEdge()
	go TransmitToEdge()
	//LaunchApp("/pkg/app.go")
	go ListenToController()
	for {
		SendToVagent(pubadd11msg)
		SendToVagent(subadd11msg)
		time.Sleep(10*time.Second)
		SendToVagent(pubrem11msg)
		SendToVagent(subrem11msg)
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


func removeVertexInfo(vi int, vertexSlice []VertexInfo) ([]VertexInfo){
	stops <- 5
	stopp <- 5
	stopg <- 5
	vlen := len(vertexSlice)
	if vlen == 0 {
		return []VertexInfo{}
	}
	vertexSlice[vi] = vertexSlice[vlen-1]
	vertexSlice[vlen-1] = VertexInfo{}
	vertexSlice = vertexSlice[:vlen-1]
	dones <- true
	donep <- true
	doneg <- true
	fmt.Println(vertexSlice)
	return vertexSlice
}


func getVertexInfo(cmsg ControlMsg, vslice []VertexInfo) (int, VertexInfo, error) {
	select {
		case s := <-stopg:
			for {
				logger.Println(
				"Waiting to Remove",
				)
				time.Sleep(s*time.Second)
				if <-doneg {
					break
				}
			}
		default: {
			break
		}
	}
	if len(vslice) == 0 {
		return -1, VertexInfo{}, fmt.Errorf("Slice empty")
	}
	for i, vi := range(vslice){
		if vi.edge == cmsg.Edge && vi.vertexno == cmsg.Vertexno {
			return i, vi, nil
		}
	}
	return -1, VertexInfo{}, fmt.Errorf("Not Found")
}


func addConnection(cmsg ControlMsg) {
	if cmsg.Vertextype == "pub" {
		i, vi, _ := getVertexInfo(cmsg, PubVertex)
		if i == -1 {
			vi = InitVertex(
				cmsg.Edge,
				cmsg.Vertexno,
				"pub",
			)
			pub <- vi
		} else {
			logger.Println("Vertex already exists")
		}
	}
	if cmsg.Vertextype == "sub" {
		i, vi, _ := getVertexInfo(cmsg, SubVertex)
		if i == -1 {
			vi = InitVertex(
				cmsg.Edge,
				cmsg.Vertexno,
				"sub",
			)
			sub <- vi
		} else {
			logger.Println("Vertex already exists")
		}
	}
	fmt.Println(SubVertex)
	fmt.Println(PubVertex)
}

func remConnection(cmsg ControlMsg){
	i, _, _ := getVertexInfo(cmsg, SubVertex)
	if i != -1 {
		SubVertex = removeVertexInfo(i, SubVertex)
	} else {
		logger.Println("Vertex not found SubVertex")
	}
	i, _, _ = getVertexInfo(cmsg, PubVertex)
	if i != -1 {
		SubVertex = removeVertexInfo(i, PubVertex)
	} else {
		logger.Println("Vertex not found in PubVertex")
	}

}

func handleController(conn net.Conn) {
	dec := json.NewDecoder(conn)
	var cmsg ControlMsg
	err := dec.Decode(&cmsg)
	if err != nil {
		logger.Println(err)
	}
	if cmsg.Cmd == "add" {
		addConnection(cmsg)
	}
	if cmsg.Cmd == "rem" {
		remConnection(cmsg)
	}
	defer conn.Close()

}
