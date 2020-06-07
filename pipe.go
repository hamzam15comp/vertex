package pipe 

import (
	"bytes"
	"io"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"os"
	"path/filepath"
	"syscall"

)

type PipeData struct {
	SendTo   string	`json:"sendto"`
	Datatype string `json:"datatype"`
	Data     []byte `json:"data"`
}

func createPipe(pipeName string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	namedPipe := filepath.Join(pwd, pipeName)
	syscall.Mkfifo(namedPipe, 0600)
}
func readFromPipe(pipeName string) PipeData {
	var buff bytes.Buffer
	var p PipeData
	input, _ := os.OpenFile(
		pipeName,
		os.O_RDONLY,
		os.ModeNamedPipe)
	io.Copy(&buff, input)
	b := buff.Bytes()
	err := json.Unmarshal(b, &p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func writeToPipe(pipeName string, pdata PipeData) {
	time.Sleep(5*time.Second)
	output, _ := os.OpenFile(pipeName, os.O_RDWR, 0600)
	b, err := json.Marshal(pdata)
	if err != nil {
		log.Fatal(err)
	}
	output.Write(b)
}

//func main() {
//	createPipe("in")
//	pi := readFromPipe("in")
//	fmt.Printf("%v", pi)
//}
