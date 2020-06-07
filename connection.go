package vertex

import (
	"fmt"
	"os/exec"
)

var IN string = "/tmp/in"
var OUT string = "/tmp/out"

func ReadData() (string, []byte, error) {
	pdata, err := ReadFromPipe(IN)
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
	err := WriteToPipe(OUT, pdata)
	if err != nil {
		return fmt.Errorf("Write failed")
	}
	return nil
}

func LaunchApp(appname string) error {
	CreatePipe(IN)
	CreatePipe(OUT)

	exe := exec.Command("go", "run", appname)
	err := exe.Run()
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
