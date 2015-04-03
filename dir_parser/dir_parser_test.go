package dir_parser_test

import (
	"github.com/ibmjstart/cf-download/cmd_exec_fake"
	. "github.com/ibmjstart/cf-download/dir_parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//unit tests
var _ = Describe("DirParser", func() {
	var p Parser
	var cmdExec cmd_exec_fake.FakeCmdExec

	BeforeEach(func() {
		cmdExec = cmd_exec_fake.NewCmdExec()
		p = NewParser(cmdExec, "TestApp", "0", false, false)
	})
	Describe("Test getFailedDownloads()", func() {
		It("Should return empty []string", func() {
			fails := p.GetFailedDownloads()
			Ω(len(fails)).To(Equal(0))
			Ω(cap(fails)).To(Equal(0))
		})
	})

	Describe("Test ExecParseDir()", func() {
		It("Should return 8 files and 3 directories", func() {
			cmdExec.SetOutput("Getting files for app smithInTheHouse in org jstart / space evans as email@us.ibm.com...\nOK\n\n.npmignore 136B\nLICENSE 1.1K\nREADME.md 5.3K\nReadme_zh-cn.md 28.4K\nbin/ -\ncomponent.json 282B\nindex.js 95B\njade-language.md 20.0K\njade.js 757.2K\njade.md 11.3K\nlib/ -\nnode_modules/ -\npackage.json 2.0K\nruntime.js 5.1K")
			files, directories := p.ExecParseDir("readPath")
			Ω(len(files)).To(Equal(11))
			Ω(files[0]).To(Equal(".npmignore"))
			Ω(files[1]).To(Equal("LICENSE"))
			Ω(files[2]).To(Equal("README.md"))
			Ω(files[3]).To(Equal("Readme_zh-cn.md"))
			Ω(files[4]).To(Equal("component.json"))
			Ω(files[5]).To(Equal("index.js"))
			Ω(files[6]).To(Equal("jade-language.md"))
			Ω(files[7]).To(Equal("jade.js"))
			Ω(files[8]).To(Equal("jade.md"))
			Ω(files[9]).To(Equal("package.json"))
			Ω(files[10]).To(Equal("runtime.js"))
			Ω(len(directories)).To(Equal(3))
			Ω(directories[0]).To(Equal("bin/"))
			Ω(directories[1]).To(Equal("lib/"))
			Ω(directories[2]).To(Equal("node_modules/"))
		})
	})
})
