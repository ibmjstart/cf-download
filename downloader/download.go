package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ibmjstart/cf-download/cmd_exec"
	"github.com/ibmjstart/cf-download/dir_parser"
	"github.com/ibmjstart/cf-download/filter"
	"github.com/mgutz/ansi"
)

type Downloader interface {
	Download(files, dirs []string, readPath, writePath string, filterList []string) error
	DownloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error
	WriteFile(readPath, writePath string, output []byte, err error) error
	CheckDownload(readPath string, file []string, err error) error
	GetFilesDownloadedCount() int
	GetFailedDownloads() []string
}

type downloader struct {
	cmdExec                    cmd_exec.CmdExec
	rootWorkingDirectoryServer string
	appName                    string
	instance                   string
	verbose                    bool
	onWindows                  bool
	failedDownloads            []string
	filesDownloaded            int
	parser                     dir_parser.Parser
	wg                         *sync.WaitGroup
}

func NewDownloader(cmdExec cmd_exec.CmdExec, WG *sync.WaitGroup, appName, instance, rootWorkingDirectoryServer string, verbose, onWindows bool) *downloader {

	return &downloader{
		cmdExec:                    cmdExec,
		rootWorkingDirectoryServer: rootWorkingDirectoryServer,
		appName:                    appName,
		instance:                   instance,
		verbose:                    verbose,
		onWindows:                  onWindows,
		parser:                     dir_parser.NewParser(cmdExec, appName, instance, onWindows, verbose),
		wg:                         WG,
	}
}

// error struct that allows appending error messages
type cliError struct {
	err    error
	errMsg string
}

/*
*	given file and directory names, download() will download the files from
* 	'readPath' and write them to disk on the 'writepath'.
* 	the function calls it's self recursively for each directory as it travels down the tree.
* 	Each call runs on a seperate go routine and and calls a go routine for every
* 	file download.
 */
func (d *downloader) Download(files, dirs []string, readPath, writePath string, filterList []string) error {
	defer d.wg.Done()

	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(err, "Error D1: failed to create directory.")

	// download each file
	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val

		filePath := strings.TrimPrefix(strings.TrimSuffix(fileRPath, "/"), d.rootWorkingDirectoryServer)

		if filter.CheckToFilter(filePath, filterList) {
			continue
		}

		d.wg.Add(1)
		go d.DownloadFile(fileRPath, fileWPath, d.wg)
	}

	// call download on every sub directory
	for _, val := range dirs {
		dirWPath := writePath + filepath.FromSlash(val)
		dirRPath := readPath + val
		dirPath := strings.TrimPrefix(strings.TrimSuffix(dirRPath, "/"), d.rootWorkingDirectoryServer)

		if filter.CheckToFilter(dirPath, filterList) {
			continue
		}

		err := os.MkdirAll(dirWPath, 0755)
		check(err, "Error D2: failed to create directory.")

		files, dirs = d.parser.ExecParseDir(dirRPath)

		d.wg.Add(1)
		go d.Download(files, dirs, dirRPath, dirWPath, filterList)
	}
	return nil
}

/*
*	downloadFile() takes a 'readPath' which corresponds to a file in the cf app. The file is
*	downloaded using the cmd_exec package which uses the os/exec library to call cf files with the given readPath. The output is
*	written to a file at writePath.
 */
func (d *downloader) DownloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error {
	defer fileDownloadGroup.Done()

	output, err := d.cmdExec.GetFile(d.appName, readPath, d.instance)

	d.WriteFile(readPath, writePath, output, err)

	return nil
}

func (d *downloader) WriteFile(readPath, writePath string, output []byte, err error) error {
	file := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid files or download issues
	downloadErr := d.CheckDownload(readPath, file, err)

	if downloadErr == nil {
		fileAsString := file[2]

		// there is currently an issue open to change the behavior for empty files
		// https://github.com/cloudfoundry/cli/issues/869
		if strings.Contains(fileAsString, "No files found") {
			fileAsString = ""
		}

		if d.verbose {
			fmt.Printf("Writing file: %s\n", readPath)
		}

		// write downloaded file to writePath
		err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)

		for i := 1; i <= 32 && err != nil; i *= 2 {
			time.Sleep(time.Duration(i) * time.Second)
			err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
		}

		if err == nil {
			// increment download counter for commandline display
			// see consoleWriter() in main.go
			d.filesDownloaded++
		} else {
			errMsg := createMessage(" Write Error: '"+readPath+"' encountered error while writing to local file", "yellow", d.onWindows)
			d.failedDownloads = append(d.failedDownloads, errMsg)
		}

	}

	return err
}

func (d *downloader) CheckDownload(readPath string, file []string, err error) error {
	if len(file) >= 2 && strings.Contains(file[1], "OK") {
		return nil
	} else {
		errMsg := createMessage(" Server Error: '"+readPath+"' not downloaded", "yellow", d.onWindows)

		d.failedDownloads = append(d.failedDownloads, errMsg)

		if d.verbose {
			fmt.Println(errMsg)
			// check for other errors
			if err != nil && len(file) >= 2 {
				PrintSlice(file)
			}
		}
		return errors.New("download failed")
	}
}

func (d *downloader) GetFilesDownloadedCount() int {
	return d.filesDownloaded
}

func (d *downloader) GetFailedDownloads() []string {
	return d.failedDownloads
}

// error check function
func check(e error, errMsg string) {
	if e != nil {
		fmt.Println("\nError: ", e)
		if errMsg != "" {
			fmt.Println("Message: ", errMsg)
		}
		os.Exit(1)
	}
}

// prints slices in readable format
func PrintSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index, ": ", val)
	}
	return nil
}

func createMessage(message, color string, onWindows bool) string {
	errmsg := ansi.Color(message, color)
	if onWindows == true {
		errmsg = message
	}

	return errmsg
}
