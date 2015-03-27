package cmd_exec_fake

type FakeCmdExec interface {
	GetFile(appName, readPath, instance string) ([]byte, error)
}

type cmdExec struct {
	output string
}

func NewCmdExec() FakeCmdExec {
	return &cmdExec{}
}

func (c *cmdExec) SetOutput(output string) {
	c.output = output
}

func (c *cmdExec) GetFile(appName, readPath, instance string) ([]byte, error) {
	// call cf files using os/exec
	return []byte(c.output), nil
}
