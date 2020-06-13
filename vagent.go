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

var pub21msg ControlMsg = ControlMsg{
	Edge: 2,
	Vertexno: 1,
	Vertextype: "pub",
	Cmd:	"add",
	Msgid: "124",
}
var sub11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "sub",
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
				PubVertex = removeVertexInfo(i, PubVertex)
				logger.Printf(
					`Send to edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
				break
			}
			fmt.Println(vi)
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
			fmt.Println(vi)
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

var exslice = []byte{'a', 'b', 'c', 'd', 'e'}

func Vamain() {
	logInit()
	go ListenToEdge()
	go TransmitToEdge()
	//LaunchApp("/pkg/app.go")
	go ListenToController()
	for {
		SendToVagent(pub21msg)
		SendToVagent(sub11msg)
		time.Sleep(10*time.Second)
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
	vlen := len(vertexSlice)
	if vlen == 0 {
		return []VertexInfo{}
	}
	vertexSlice[vi] = vertexSlice[vlen-1]
	vertexSlice[vlen-1] = nil
	vertexSlice = vertexSlice[:vlen-1]
	return vertexSlice
}


func getVertexInfo(
	cmsg ControlMsg,
	vslice VertexInfo[]) (int, VertexInfo, error) {
	if len(vslice) == 0 {
		return -1, VertexInfo{}, log.Errorf("Slice empty")
	}
	for i, vi := range(vslice){
		if vi.edge == cmd.edge && vi.vertexno == cmsg.Vertexno {
			return i, vi, nil
		}
	}
	return -1, VertexInfo{}, log.Errorf("Not Found")
}


func addConnection(cmsg ControlMsg) {
	if cmsg.Vertextype == "pub" {
		i, vi := getVertexInfo(cmsg, PubVertex)
		if i == -1 {
			vi = InitVertex(
				cmsg.edge,
				cmsg.Vertexno,
				"pub",
			)
			pub <- vi
		}
		else {
			logger.Println("Vertex already exists")
		}
	}
	if cmsg.Vertextype == "sub" {
		i, vi := getVertexInfo(cmsg, SubVertex)
		if i == -1 {
			vi = InitVertex(
				cmsg.edge,
				cmsg.Vertexno,
				"sub",
			)
			sub <- vi
		}
		else {
			logger.Println("Vertex already exists")
		}
	}
}

func remConnection(cmsg ControlMsg){
	i, _ := getVertexInfo(cmsg, SubVertex)
	if i != -1 {
		SubVertex = removeVertexInfo(i, SubVertex)
	}
	else {
		logger.Println("Vertex not found SubVertex")
	}
	i, _ := getVertexInfo(cmsg, PubVertex)
	if i != -1 {
		SubVertex = removeVertexInfo(i, PubVertex)
	}
	else {
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
