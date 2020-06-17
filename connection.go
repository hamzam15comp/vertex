package vertex

import (
	"fmt"
	"os/exec"
)

var IN string = "/in"
var OUT string = "/out"

func ReadData() (string, []byte, error) {
	pdata, err := ReadFromPipe(IN)
	if err != nil {
		CreatePipe(IN)
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
	err := WriteToPipe(OUT, pdata)
	if err != nil {
		CreatePipe(OUT)
		return fmt.Errorf("Write failed")
	}
	return nil
}

func LaunchApp(appname string) error {
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
