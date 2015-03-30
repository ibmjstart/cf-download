package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/cf-download/cmd_exec"
	"github.com/cf-download/dir_parser"
	"github.com/cf-download/filter"
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
	cmdExec              cmd_exec.CmdExec
	rootWorkingDirectory string
	appName              string
	instance             string
	verbose              bool
	onWindows            bool
	failedDownloads      []string
	filesDownloaded      int
	parser               dir_parser.Parser
	wg                   *sync.WaitGroup
}

func NewDownloader(cmdExec cmd_exec.CmdExec, WG *sync.WaitGroup, appName, instance, rootWorkingDirectory string, verbose, onWindows bool) *downloader {

	return &downloader{
		cmdExec:              cmdExec,
		rootWorkingDirectory: rootWorkingDirectory,
		appName:              appName,
		instance:             instance,
		verbose:              verbose,
		onWindows:            onWindows,
		parser:               dir_parser.NewParser(cmdExec, appName, instance, onWindows, verbose),
		wg:                   WG,
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
	check(cliError{err: err, errMsg: "Called by: download"})

	// download each file
	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val

		if filter.CheckToFilter(fileRPath, d.rootWorkingDirectory, filterList) {
			continue
		}

		d.wg.Add(1)
		go d.DownloadFile(fileRPath, fileWPath, d.wg)
	}

	// call download on every sub directory
	for _, val := range dirs {
		dirWPath := writePath + val
		dirRPath := readPath + val

		if filter.CheckToFilter(dirRPath, d.rootWorkingDirectory, filterList) {
			continue
		}

		err := os.MkdirAll(dirWPath, 0755)
		check(cliError{err: err, errMsg: "Called by: download"})

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
	err = d.WriteFile(readPath, writePath, output, err)
	check(cliError{err: err, errMsg: "Called by: downloadFile 2"})

	return nil
}

func (d *downloader) WriteFile(readPath, writePath string, output []byte, err error) error {
	file := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid files or download issues
	d.CheckDownload(readPath, file, err)

	if d.verbose {
		fmt.Printf("Writing file: %s\n", readPath)
	} else {
		// increment download counter for commandline display
		// see consoleWriter() in main.go
		d.filesDownloaded++
	}

	fileAsString := file[2]
	// write downloaded file to writePath
	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	return err
}

func (d *downloader) CheckDownload(readPath string, file []string, err error) error {
	// check for invalid file error.
	// some files are inaccesible from the cf files (permission issues) this is rare but we need to
	// alert users if it occurs. It usually happens in vendor files.
	if strings.Contains(file[1], "FAILED") {
		errmsg := ansi.Color(" Server Error: '"+readPath+"' not downloaded", "yellow")
		if d.onWindows == true {
			errmsg = " Server Error: '" + readPath + "' not downloaded"
		}
		d.failedDownloads = append(d.failedDownloads, errmsg)
		if d.verbose {
			fmt.Println(errmsg)
		}
		return errors.New("download failed")
	} else {
		// check for other errors
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}
	return nil
}

func (d *downloader) GetFilesDownloadedCount() int {
	return d.filesDownloaded
}

func (d *downloader) GetFailedDownloads() []string {
	return d.failedDownloads
}

func check(e cliError) {
	if e.err != nil {
		fmt.Println("\nError: ", e.err)
		if e.errMsg != "" {
			fmt.Println("Message: ", e.errMsg)
		}
		os.Exit(1)
	}
}
