package downloader_test

import (
	"errors"
	"fmt"
	"github.com/cf-download/cmd_exec_fake"
	. "github.com/cf-download/downloader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"sync"
)

var _ = Describe("Downloader tests", func() {
	var (
		wg               sync.WaitGroup
		d                Downloader
		cmdExec          cmd_exec_fake.FakeCmdExec
		currentDirectory string
	)

	currentDirectory, _ = os.Getwd()
	os.MkdirAll(currentDirectory+"/testFiles/", 0755)

	cmdExec = cmd_exec_fake.NewCmdExec()
	d = NewDownloader(cmdExec, &wg, "appName", "0", "rootWorkingDirectory", false, false)

	// downloadfile also tests the following functions
	// WriteFile(), CheckDownload()
	Describe("Test DownloadFile function", func() {
		Context("download and create a file containing only 'helloWorld'", func() {
			It("create test1.txt", func() {
				writePath := currentDirectory + "/testFiles/test1.txt"
				cmdExec.SetOutput("Getting files for app payToWin in org jstart / space koldus as email@us.ibm.com...\nOK\nHello World")
				wg.Add(1)
				go d.DownloadFile("", writePath, &wg)
				wg.Wait()

				fileContents, err := ioutil.ReadFile(writePath)
				Ω(err).To(BeNil())
				Ω(string(fileContents)).To(Equal("Hello World"))
			})
		})
	})

	Describe("Test DownloadFile function", func() {
		Context("download and create a file containing only 'helloWorld'", func() {
			It("create test2.txt", func() {
				writePath := currentDirectory + "/testFiles/test2.txt"
				cmdExec.SetOutput("Getting files for app payToWin in org jstart / space koldus as email@us.ibm.com...\nOK\nLorem ipsum is a pseudo-Latin text used in web design, typography, layout, and printing in place of English to emphasise design elements over content. It's also called placeholder (or filler) text. It's a convenient tool for mock-ups. It helps to outline the visual elements of a document or presentation, eg typography, font, or layout. Lorem ipsum is mostly a part of a Latin text by the classical author and philosopher Cicero. Its words and letters have been changed by addition or removal, so to deliberately render its content nonsensical; it's not genuine, correct, or comprehensible Latin anymore. While lorem ipsum's still resembles classical Latin, it actually has no meaning whatsoever. As Cicero's text doesn't contain the letters K, W, or Z, alien to latin, these, and others are often inserted randomly to mimic the typographic appearence of European languages, as are digraphs not to be found in the original.")
				wg.Add(1)
				go d.DownloadFile("", writePath, &wg)
				wg.Wait()

				fileInfo, err := os.Stat(writePath)
				Ω(err).To(BeNil())
				Ω(fileInfo.Name()).To(Equal("test2.txt"))
				Ω(fileInfo.Size()).To(BeEquivalentTo(924))
				Ω(fileInfo.IsDir()).To(BeFalse())
			})
		})
	})

	// this test will test all functions in the downloader package
	Describe("Test Download Function", func() {
		Context("download the fake directory found in fake_cmd_exec", func() {
			It("should download and write the files to the testFiles Directory", func() {
				readPath := currentDirectory
				writePath := currentDirectory + "/test-download"
				cmdExec.SetOutput("Getting files for app payToWin in org jstart / space koldus as email@us.ibm.com...\nOK\nLorem ipsum is a pseudo-Latin text used in web design, typography, layout, and printing in place of English to emphasise design elements over content. It's also called placeholder (or filler) text. It's a convenient tool for mock-ups. It helps to outline the visual elements of a document or presentation, eg typography, font, or layout. Lorem ipsum is mostly a part of a Latin text by the classical author and philosopher Cicero. Its words and letters have been changed by addition or removal, so to deliberately render its content nonsensical; it's not genuine, correct, or comprehensible Latin anymore. While lorem ipsum's still resembles classical Latin, it actually has no meaning whatsoever. As Cicero's text doesn't contain the letters K, W, or Z, alien to latin, these, and others are often inserted randomly to mimic the typographic appearence of European languages, as are digraphs not to be found in the original.")
				cmdExec.SetFakeDir(true)

				//remeber to turn off fake dir flag
				defer cmdExec.SetFakeDir(false)

				files := []string{}
				dirs := []string{"testFiles"}
				filterList := []string{"ignore.go"}

				fmt.Println("READPATH: ", currentDirectory)
				//wg.Add(1)
				go d.Download(files, dirs, readPath, writePath, filterList)
				//wg.Wait()

				Ω(len(dirs)).To(Equal(1))
			})
		})
	})

	Describe("Test checkDownload Function", func() {
		Context("when we recieve server error", func() {
			It("Should return server error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "FAILED"

				err := d.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(Equal(errors.New("download failed")))
			})
		})

		Context("when we recieve no error", func() {
			It("Should return no error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "OK"

				err := d.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("Test GetFilesDownloadedCount Function", func() {
		It("after downloading 2 files", func() {
			count := d.GetFilesDownloadedCount()
			Ω(count).To(Equal(2))
		})
	})

	Describe("Test getFailedDownloads()", func() {
		It("Should have 1 failed download from previous CheckDownload Test", func() {
			fails := d.GetFailedDownloads()
			Ω(len(fails)).To(Equal(1))
			Ω(cap(fails)).To(Equal(1))
		})
	})
})

// prints slices in readable format
func PrintSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index, ": ", val)
	}
	return nil
}

/*
	Describe("", func() {
		Context("", func() {
			It("", func() {
				Ω().To(Equal())
			})
		})
	})
*/
