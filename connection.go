package vertex

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"github.com/hamzam15comp/vertex"

)

func main() {

	lsCmd := exec.Command("bash", "-c", "ls -a -l -h")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("> ls -a -l -h")
	fmt.Println(string(lsOut))
}
