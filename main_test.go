package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/ibmjstart/cf-download"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// unit tests of individual functions
var _ = Describe("CfDownload", func() {
	var args []string

	BeforeEach(func() {
		args = make([]string, 7)
	})

	Describe("Test Flag functionality", func() {

		Context("Check if overWrite flag works", func() {
			It("Should set the overwrite_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "app/files/htdocs"
				args[3] = "--overwrite"

				flagVals := ParseFlags(args)
				Expect(flagVals.OverWrite_flag).To(BeTrue())
				Expect(flagVals.File_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if file flag works", func() {
			It("Should set the file_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--file"

				flagVals := ParseFlags(args)
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.File_flag).To(BeTrue())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if verbose flag works", func() {
			It("Should set the verbose_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--verbose"

				flagVals := ParseFlags(args)
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.File_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeTrue())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if instance (i) flag works", func() {
			It("Should set the instance_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--i"
				args[3] = "3"

				flagVals := ParseFlags(args)
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.File_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("3"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if omit flag works", func() {
			It("Should set the omit_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--omit"
				args[3] = "app/node_modules"

				flagVals := ParseFlags(args)
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.File_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal("app/node_modules"))
			})
		})

	})

	Describe("test directoryContext parsing", func() {

		It("Should return correct strings", func() {
			args[0] = "download"
			args[1] = "app_name"
			args[2] = "app/src/node"
			args[3] = "--verbose"
			currentDirectory, _ := os.Getwd()
			currentDirectory = filepath.ToSlash(currentDirectory)
			rootWD, startingPath := GetDirectoryContext(currentDirectory, args, false)

			correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

			Expect(correctSuffix).To(BeTrue())
			Expect(startingPath).To(Equal("/app/src/node/"))
		})

		It("should still return /app/src/node/ for startingPath (INPUT has leading and trailing slash)", func() {
			args[0] = "download"
			args[1] = "app_name"
			args[2] = "/app/src/node/"
			args[3] = "--verbose"
			currentDirectory, _ := os.Getwd()
			currentDirectory = filepath.ToSlash(currentDirectory)
			rootWD, startingPath := GetDirectoryContext(currentDirectory, args, false)

			correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

			Expect(correctSuffix).To(BeTrue())
			Expect(startingPath).To(Equal("/app/src/node/"))
		})

		It("should still return /app/src/node/ for startingPath (INPUT only has trailing slash)", func() {
			args[0] = "download"
			args[1] = "app_name"
			args[2] = "app/src/node/"
			args[3] = "--verbose"
			currentDirectory, _ := os.Getwd()
			currentDirectory = filepath.ToSlash(currentDirectory)
			rootWD, startingPath := GetDirectoryContext(currentDirectory, args, false)

			correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

			Expect(correctSuffix).To(BeTrue())
			Expect(startingPath).To(Equal("/app/src/node/"))
		})

		It("should still return /app/src/node/ for startingPath (INPUT only has leading slash)", func() {
			args[0] = "download"
			args[1] = "app_name"
			args[2] = "/app/src/node"
			args[3] = "--verbose"
			currentDirectory, _ := os.Getwd()
			currentDirectory = filepath.ToSlash(currentDirectory)
			rootWD, startingPath := GetDirectoryContext(currentDirectory, args, false)

			correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

			Expect(correctSuffix).To(BeTrue())
			Expect(startingPath).To(Equal("/app/src/node/"))
		})

		It("should return /app/src/file.html for startingPath (--file flag specified)", func() {
			args[0] = "download"
			args[1] = "app_name"
			args[2] = "/app/src/file.html"
			args[3] = "--file"
			currentDirectory, _ := os.Getwd()
			currentDirectory = filepath.ToSlash(currentDirectory)
			rootWD, startingPath := GetDirectoryContext(currentDirectory, args, true)

			correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/file.html")

			Expect(correctSuffix).To(BeTrue())
			Expect(startingPath).To(Equal("/app/src/file.html"))
		})

	})

	Describe("test error catching in run() [MUST HAVE PLUGIN INSTALLED TO PASS]", func() {
		Context("when appname begins with -- or -", func() {
			It("Should print error, because user has flags before appname", func() {
				cmd := exec.Command("cf", "download", "--appname")
				output, _ := cmd.CombinedOutput()
				Expect(strings.Contains(string(output), "Error: App name begins with '-' or '--'. correct flag usage: 'cf download APP_NAME [--flags]'")).To(BeTrue())
			})

			It("Should print error, because user not specified an appName", func() {
				cmd := exec.Command("cf", "download")
				output, _ := cmd.CombinedOutput()
				Expect(strings.Contains(string(output), "Error: Missing App Name")).To(BeTrue())
			})

			It("Should print error, test overwrite flag functionality", func() {
				// create directory that needs to be overwritten
				os.Mkdir("test-download", 755)

				cmd := exec.Command("cf", "download", "test")
				output, _ := cmd.CombinedOutput()
				Expect(strings.Contains(string(output), "already Exists and is not an empty directory.")).To(BeTrue())

				// clean up
				os.RemoveAll("test-download")
			})

			It("Should print error, instance flag not int", func() {
				cmd := exec.Command("cf", "download", "test", "-i", "hello")
				output, _ := cmd.CombinedOutput()
				Expect(strings.Contains(string(output), "Error:  invalid value ")).To(BeTrue())
			})

			It("Should print error, invalid flag", func() {
				cmd := exec.Command("cf", "download", "test", "-ooverwrite")
				output, _ := cmd.CombinedOutput()
				Expect(strings.Contains(string(output), "Error:  flag provided but not defined: -ooverwrite")).To(BeTrue())
			})
		})
	})

})
