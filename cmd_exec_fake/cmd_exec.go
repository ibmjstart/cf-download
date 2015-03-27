package cmd_exec_fake

import (
	"fmt"
	"io/ioutil"
	"os"
)

type FakeCmdExec interface {
	GetFile(appName, readPath, instance string) ([]byte, error)
	SetOutput(output string)
	SetFakeDir(flag bool)
}

type cmdExec struct {
	output     string
	useFakeDir bool
}

func NewCmdExec() FakeCmdExec {
	return &cmdExec{}
}

func (c *cmdExec) SetOutput(output string) {
	c.output = output
}

func (c *cmdExec) SetFakeDir(flag bool) {
	c.useFakeDir = flag
}

func (c *cmdExec) GetFile(appName, readPath, instance string) ([]byte, error) {
	var output []byte
	if c.useFakeDir == false {
		return []byte(c.output), nil
	}

	// Needs to be appended to every response (for parser)
	startString := "Getting files for app payToWin in org jstart / space koldus as email@us.ibm.com...\nOK\n"

	fileInfo, _ := os.Stat(readPath)
	fmt.Println("ReadPath: ", readPath)
	fmt.Println("FileInfo: ", fileInfo)
	if fileInfo.IsDir() {
		file, _ := os.Open(readPath)
		dirFiles, _ := file.Readdir(0)
		var dirString string
		for _, val := range dirFiles {
			if val.IsDir() {
				dirString += "\n" + val.Name() + "/		-"
			} else {
				dirString += "\n" + val.Name() + " 		1B"
			}
		}

		output = []byte(startString + dirString)

	} else {
		fileContents, _ := ioutil.ReadFile(readPath)
		output = []byte(startString + string(fileContents))
	}
	return output, nil
}
