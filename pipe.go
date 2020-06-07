package main 

import (
	"bytes"
	"encoding/json"
//	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

type PipeData struct {
	SendTo   string `json:"sendto"`
	Datatype string `json:"datatype"`
	Data     []byte `json:"data"`
}

func CreatePipe(pipeName string) error {
	pwd, pwderr := os.Getwd()
	if pwderr != nil {
		return pwderr
	}
	namedPipe := filepath.Join(pwd, pipeName)
	mkerr := syscall.Mkfifo(namedPipe, 0660)
	if mkerr != nil && !os.IsExist(mkerr) {
		return nil
	}
	f, operr := os.OpenFile(pipeName, os.O_RDWR, 0660)
	if operr != nil {
		return operr
	}
	// In case we're using a pre-made file, check that it's actually a FIFO
	fi, ferr := f.Stat()
	if ferr != nil {
		f.Close()
		return ferr
	}
	if fi.Mode()&os.ModeType != os.ModeNamedPipe {
		f.Close()
		return os.ErrExist
	}
	return nil
}

func ReadFromPipe(pipeName string) (PipeData, error) {
	var buff bytes.Buffer
	var p PipeData
	input, operr := os.OpenFile(
		pipeName,
		os.O_RDONLY,
		os.ModeNamedPipe)
	if operr != nil {
		return PipeData{}, operr
	}
	io.Copy(&buff, input)
	b := buff.Bytes()
	jerr := json.Unmarshal(b, &p)
	if jerr != nil {
		return PipeData{}, jerr
	}
	return p, nil
}

func WriteToPipe(pipeName string, pdata PipeData) error {
	output, operr := os.OpenFile(
		pipeName,
		os.O_RDWR,
		os.ModeNamedPipe)
	if operr != nil {
		return operr
	}
	b, jerr := json.Marshal(pdata)
	if jerr != nil {
		return jerr
	}
	_, wrerr := output.Write(b)
	if wrerr != nil {
		return wrerr
	}
	return nil
}

//func main() {
//	CreatePipe("in")
//	pi, perr := ReadFromPipe("in")
//	if perr != nil {
//		fmt.Println(perr)
//		return
//	}
//	fmt.Printf("%v", pi)
//}
