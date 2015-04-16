package downloader_test

import (
	"errors"
	"github.com/ibmjstart/cf-download/cmd_exec/cmd_exec_fake"
	. "github.com/ibmjstart/cf-download/downloader"
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

	AfterEach(func() {
		// turn off the directory faker
		cmdExec.SetFakeDir(false)

		os.RemoveAll("testFiles/test1.txt")
		os.RemoveAll("testFiles/test2.txt")
		os.RemoveAll(currentDirectory + "/test-download")
	})

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
				writePath = currentDirectory + "/testFiles/test2.txt"
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

	Describe("Test checkDownload Function", func() {
		Context("when we recieve permission error", func() {
			It("Should return server error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "FAILED"

				err := d.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(Equal(errors.New("download failed")))
			})
		})

		Context("when we recieve 502 error", func() {
			It("Should return server error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "status code: 502"

				// Throw away Stdout
				oldStdout := os.Stdout
				os.Stdout = nil

				err := d.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)

				// restore Stdout
				os.Stdout = oldStdout
				Expect(err.Error()).To(Equal("502"))
			})
		})

		Context("when we recieve 400 error", func() {
			It("Should return server error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "status code: 400"

				// Throw away Stdout
				oldStdout := os.Stdout
				os.Stdout = nil

				err := d.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)

				// restore Stdout
				os.Stdout = oldStdout
				Expect(err.Error()).To(Equal("400"))
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
		It("Should have 3 failed download from previous CheckDownload Test", func() {
			fails := d.GetFailedDownloads()
			Ω(len(fails)).To(Equal(3))
		})
	})

	/*
	*	The Following test call the download function on a test directory (testFiles) which exists in cf-download/downloader/
	*	The directory will be loaded into the download function and written to a new folder called test-download. The test suite
	*	will verify that the directory was loaded and written correctly. Once everything is verified then the test-download directory
	*	will be deleted as to not interfere with the next test. This not only test the entire download functionality but also the
	*	parsing and all flags.
	 */

	Describe("Test Download() Function", func() {
		Context("download the entire directory (no filter)", func() {
			It("should download and write the files to the testFiles Directory", func() {
				d = NewDownloader(cmdExec, &wg, "appName", "0", "rootWorkingDirectory", false, false)
				readPath := currentDirectory
				writePath := currentDirectory + "/test-download"

				// delete the test folder (if exists) before testing
				os.RemoveAll(writePath)

				cmdExec.SetFakeDir(true)

				files := []string{}
				dirs := []string{"/testFiles/"}
				filterList := []string{currentDirectory + "/testFiles/.DS_Store", currentDirectory + "/testFiles/app_content/.DS_Store"}

				wg.Add(1)
				go d.Download(files, dirs, readPath, writePath, filterList)
				wg.Wait()

				// test root structure
				rootInfo, _ := os.Stat(writePath + "/testFiles/")
				Ω(rootInfo.IsDir()).To(BeTrue())

				rootFile, _ := os.Open(writePath + "/testFiles/")
				rootContents, _ := rootFile.Readdir(0)

				Ω(rootContents[0].Name()).To(Equal("app_content"))
				Ω(rootContents[1].Name()).To(Equal("ignore.go"))
				Ω(rootContents[2].Name()).To(Equal("ignoreDir"))
				Ω(rootContents[3].Name()).To(Equal("notignored.go"))

				// test the contents of the app_contents directory
				Ω(rootContents[0].IsDir()).To(BeTrue())
				appContentFolder, _ := os.Open(writePath + "/testFiles/" + rootContents[0].Name())
				appContents, _ := appContentFolder.Readdir(0)
				Ω(appContents[0].Name()).To(Equal("app.go"))
				Ω(appContents[1].Name()).To(Equal("server.go"))

			})
		})
	})

	Describe("Test Download() Function", func() {
		Context("download the fake directory filtering out ignore.go and ignoreDir", func() {
			It("should download and write the files to the testFiles Directory", func() {
				d = NewDownloader(cmdExec, &wg, "appName", "0", "rootWorkingDirectory", false, false)
				readPath := currentDirectory
				writePath := currentDirectory + "/test-download"

				// delete the test folder (if exists) before testing
				os.RemoveAll(writePath)

				cmdExec.SetFakeDir(true)

				files := []string{}
				dirs := []string{"/testFiles/"}
				filterList := []string{currentDirectory + "/testFiles/ignore.go", currentDirectory + "/testFiles/ignoreDir", currentDirectory + "/testFiles/.DS_Store", currentDirectory + "/testFiles/app_content/.DS_Store"}

				wg.Add(1)
				go d.Download(files, dirs, readPath, writePath, filterList)
				wg.Wait()

				// test root structure
				rootInfo, _ := os.Stat(writePath + "/testFiles/")
				Ω(rootInfo.IsDir()).To(BeTrue())

				rootFile, _ := os.Open(writePath + "/testFiles/")
				rootContents, _ := rootFile.Readdir(0)
				Ω(rootContents[0].Name()).To(Equal("app_content"))
				Ω(rootContents[1].Name()).To(Equal("notignored.go"))

				// test the contents of the app_contents directory
				Ω(rootContents[0].IsDir()).To(BeTrue())
				appContentFolder, _ := os.Open(writePath + "/testFiles/" + rootContents[0].Name())
				appContents, _ := appContentFolder.Readdir(0)
				Ω(appContents[0].Name()).To(Equal("app.go"))
				Ω(appContents[1].Name()).To(Equal("server.go"))

				// delete the folder after testing
				os.RemoveAll(writePath)

			})
		})
	})

})
