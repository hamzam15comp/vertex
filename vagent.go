package vertex

import (
	"fmt"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"
	"sync"
	"strconv"
)

var logger *log.Logger

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
	Msgid: 1,
}
var subadd11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "sub",
	Cmd:	"add",
	Msgid: 2,
}
var pubrem11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "pub",
	Cmd:	"rem",
	Msgid: 3,
}
var subrem11msg ControlMsg = ControlMsg{
	Edge: 1,
	Vertexno: 1,
	Vertextype: "sub",
	Cmd:	"rem",
	Msgid: 4,
}
var ptest PipeData = PipeData{
	SendTo: "all",
	Datatype: "datatype",
	Data: []byte("data"),
}

//var SubVertex []VertexInfo
//var PubVertex []VertexInfo
var SubVertex = make([]VertexInfo, 16, 16)
var PubVertex = make([]VertexInfo, 16, 16)
var pub = make(chan []VertexInfo, 1)
var sub = make(chan []VertexInfo, 1)
var mux sync.Mutex

func checkAddRemovePub(vertexSlice []VertexInfo)([]VertexInfo) {
	mux.Lock()
	select {
		case p := <-pub:
			logger.Println("Updated PubVertex")
			mux.Unlock()
			return p
		default:
			break
	}
	mux.Unlock()
	return vertexSlice
}
func checkAddRemoveSub(vertexSlice []VertexInfo)([]VertexInfo) {
	mux.Lock()
	select {
		case s := <-sub:
			logger.Println("Updated SubVertex")
			mux.Unlock()
			return s
		default:
			break
	}
	mux.Unlock()
	return vertexSlice
}

func TransmitToEdge(){
	for {
		pi, perr := ReadFromPipe(OUT)
		if perr != nil {
			logger.Printf(
				`Read from pipe failed Trying again...`,
			)
			continue
		}
		logger.Printf("Received from pipe\n")
		PubVertex = checkAddRemovePub(PubVertex)
		for i, vi := range PubVertex {
			if vi.edge == 0 {
				continue
			}
			logger.Println(
				"TransmitToEdge: PubV",
				vi.edge,
				vi.vertexno,
			)
			serr := SendDataEdge(
				vi,
				pi.SendTo,
				pi.Datatype,
				pi.Data,
			)
			if serr != nil {
				continue
				PubVertex = removeVertexInfo(i, PubVertex)
				logger.Println(
					"Send to edge failed. Deleting connection",
					vi.edge,
					vi.vertexno,
				)
				break
			}
			//fmt.Println("Sending to Edge", pi)
		}
	}

}



func ListenToEdge() {
	for {
		SubVertex = checkAddRemoveSub(SubVertex)
		//mux.Lock()
		for i, vi := range SubVertex {
			if vi.edge == 0 {
				continue
			}
			logger.Println(
				"ListenToEdge: SubV",
				vi.edge,
				vi.vertexno,
			)
			var p PipeData
			var err error
			p.Datatype, p.Data, err = ReceiveDataEdge(vi, true)
			if err != nil {
				continue
				SubVertex = removeVertexInfo(i, SubVertex)
				logger.Printf(
					`Receive from edge %d failed.
					Closing and deleting connection`,
					vi.edge,
				)
				break
			}
			WriteToPipe(IN, p)
			//fmt.Println("Received from Edge", p)
			logger.Println("Writing data %v to pipe")
		}
		//mux.Unlock()
	}

}


func logInit() {
	f, err := os.OpenFile("errors.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	logger = log.New(f, "[INFO]", log.Lmicroseconds | log.LUTC)
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

		handleController(conn)
	}
}


func VertexAgentLaunch() {
	logInit()
	go ListenToEdge()
	go TransmitToEdge()
	go ListenToController()
	for {
		time.Sleep(1*time.Second)
	}
}


func SendToVagent(cmsg ControlMsg, hostname string, port int) {
	host := hostname + ":" + strconv.Itoa(port)
	con, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println(err)
	}
	enc := json.NewEncoder(con)
	err = enc.Encode(cmsg)
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
}


func removeVertexInfo(vi int, vertexSlice []VertexInfo) ([]VertexInfo){
	vert := vertexSlice[vi]
	//NotifyController(
	//	ControlMsg{
	//		Edge: vert.edge,
	//		Vertexno: vert.vertexno,
	//		Vertextype: vert.vertextype,
	//		Cmd: "rem",
	//		Msgid: -2,
	//	}
	//)
	logger.Println(
		"Removing:",
		vert.vertexType,
		vert.edge,
		vert.vertexno,
	)
	vert.conn.Close()
	vert.channel.Close()
	vlen := len(vertexSlice)
	if vlen == 0 {
		return []VertexInfo{}
	}
	vertexSlice[vi] = vertexSlice[vlen-1]
	vertexSlice[vlen-1] = VertexInfo{}
	vertexSlice = vertexSlice[:vlen-1]
	return vertexSlice
}


func getVertexInfo(cmsg ControlMsg, vslice []VertexInfo) (int, VertexInfo, error) {
	for i, vi := range(vslice){
		if vi.edge == cmsg.Edge && vi.vertexno == cmsg.Vertexno {
			fmt.Println("GetVertexInfo returns: ", vi)
			return i, vi, nil
		}
	}
	return -1, VertexInfo{}, fmt.Errorf("Not Found")
}


func UpdateConnection(cmsg ControlMsg) {
	logger.Println("UpdateConn: Waiting for mutex", cmsg)
	mux.Lock()
	if cmsg.Cmd == "add" {
		if cmsg.Vertextype == "pub" {
			i, _ , _ := getVertexInfo(cmsg, PubVertex)
			if i == -1 {
				vi := InitVertex(
					cmsg.Edge,
					cmsg.Vertexno,
					"pub",
				)
				logger.Println(
					"Created Pubvertex: ",
					vi.edge,
					vi.vertexno,
				)
				PubVertex = append(PubVertex, vi)
				select {
					case <-pub:
						break
					default:
						break
				}
				pub <- PubVertex
			} else {
				logger.Println("Vertex already exists")
			}
		} else if cmsg.Vertextype == "sub" {
			i, _ , _ := getVertexInfo(cmsg, SubVertex)
			if i == -1 {
				vi := InitVertex(
					cmsg.Edge,
					cmsg.Vertexno,
					"sub",
				)
				logger.Println(
					"Created Subvertex: ",
					vi.edge,
					vi.vertexno,
				)
				SubVertex = append(SubVertex, vi)
				select {
					case <-sub:
						break
					default:
						break
				}
				sub <- SubVertex
			} else {
				logger.Println("Vertex already exists")
			}
		} else {
			logger.Println("Invalid vertextype")
		}
	} else if cmsg.Cmd == "rem" {

		if cmsg.Vertextype == "sub" {
			i, _, _ := getVertexInfo(cmsg, SubVertex)
			if i != -1 {
				SubVertex = removeVertexInfo(i, SubVertex)
				sub <- SubVertex
			} else {
				logger.Println("Vertex not found SubVertex")
			}
		} else if cmsg.Vertextype == "pub" {
			i, _, _ := getVertexInfo(cmsg, PubVertex)
			if i != -1 {
				PubVertex = removeVertexInfo(i, PubVertex)
				pub <- PubVertex
			} else {
				logger.Println("Vertex not found in PubVertex")
			}
		} else {
			logger.Println("Invalid vertextype")
		}

	} else {
		logger.Println("Invalid command")
	}
	mux.Unlock()
	logger.Println("UpdateConn: Released lock")
}

func handleController(conn net.Conn) {
	dec := json.NewDecoder(conn)
	var cmsg ControlMsg
	err := dec.Decode(&cmsg)
	if err != nil {
		logger.Println(err)
	}
	UpdateConnection(cmsg)
	defer conn.Close()

}
