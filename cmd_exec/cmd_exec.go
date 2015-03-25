package cmd_exec

import "os/exec"

type CmdExec interface {
	GetFile(appName, readPath, instance string) ([]byte, error)
}

type cmdExec struct {
}

func NewCmdExec() CmdExec {
	return &cmdExec{}
}

func (c *cmdExec) GetFile(appName, readPath, instance string) ([]byte, error) {
	// call cf files using os/exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()

	return output, err
}
