// I use this file for testing code snippets. It is not used by any files.

package main

import (
	"fmt"
	"github.com/cf-download/cmd_exec_fake"
	"os"
)

var (
	cmdExec          cmd_exec_fake.FakeCmdExec
	currentDirectory string
)

func main() {
	cmdExec = cmd_exec_fake.NewCmdExec()
	cmdExec.SetFakeDir(true)
	currentDirectory, _ = os.Getwd()
	readPath := currentDirectory + "/testFiles"

	output, _ := cmdExec.GetFile("", readPath, "")
	fmt.Println(string(output))
}
