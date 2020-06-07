package main

import (
	"fmt"
	"log"
	"os/exec"
)

func ReadData() (string, []byte, error) {
	pdata, err := ReadFromPipe("in")
	if err != nil {
		return "", nil, err
	}
	return pdata.Datatype, pdata.Data, nil
}

func WriteData(sendTo string, datatype string, data []byte) error {
	pdata := PipeData{
		SendTo:   sendTo,
		Datatype: datatype,
		Data:     data,
	}
	err := WriteToPipe("out", pdata)
	if err != nil {
		return fmt.Errorf("Write failed")
	}
	return nil
}

func LaunchApp(appname string) error {
	CreatePipe("in")
	CreatePipe("out")

	exe := exec.Command("go", "run", appname)
	err := exe.Start()
	if err != nil {
		return err
	}
	return nil
}

//func main() {
//	err := LaunchApp()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
