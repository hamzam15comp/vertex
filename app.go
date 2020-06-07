package main

import (
	"fmt"
	"github.com/hamzam15comp/vertex"

)

func main(){

	datatype, data, err := vertex.ReadData()
	fmt.Println(datatype, data)

}
