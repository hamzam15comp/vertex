package vertex

import (
	"fmt"
)

var IN string = "/in"
var OUT string = "/out"

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

func LaunchApp() error {
	CreatePipe(IN)
	CreatePipe(OUT)
	return nil
}

//func main() {
//	err := LaunchApp()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
