package vertex

import (
	"fmt"
)

var IN string = "in"
var OUT string = "out"

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
	logger.Println(pdata)
	return nil
}

func LaunchApp() error {
	perr := CreatePipe(IN)
	if perr != nil {
		return fmt.Errorf("Failed to create in")
	}
	perr = CreatePipe(OUT)
	if perr != nil {
		return fmt.Errorf("Failed to create out")
	}
	return nil
}

//func main() {
//	err := LaunchApp()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
