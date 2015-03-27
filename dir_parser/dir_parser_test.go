package dir_parser_test

import (
	"github.com/cf-download/cmd_exec_fake"
	. "github.com/cf-download/dir_parser"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

//unit tests
var _ = Describe("DirParser", func() {
	BeforeEach(func() {
		cmdExec := cmd_exec_fake.NewCmdExec()
		p := NewParser(cmdExec, "TestApp", "0", false, false)
		p.GetFailedDownloads
	})
})
