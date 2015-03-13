package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/mgutz/ansi"
)

/*
*	given file and directory names, download() will download the files from
* 	'readPath' and write them to disk on the 'writepath'.
* 	the function calls it's self recursively for each directory as it travels down the tree.
* 	Each call runs on a seperate go routine and and calls a go routine for every
* 	file download.
 */
func download(files, dirs []string, readPath, writePath string, filterList []string) error {
	defer wg.Done()

	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(cliError{err: err, errMsg: "Called by: download"})

	// download each file
	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val

		if checkToFilter(fileRPath, filterList) {
			continue
		}

		wg.Add(1)
		go downloadFile(fileRPath, fileWPath, &wg)
	}

	// call download on every sub directory
	for _, val := range dirs {
		dirWPath := writePath + val
		dirRPath := readPath + val

		if checkToFilter(dirRPath, filterList) {
			continue
		}

		err := os.MkdirAll(dirWPath, 0755)
		check(cliError{err: err, errMsg: "Called by: download"})

		files, dirs = execParseDir(dirRPath)

		wg.Add(1)
		go download(files, dirs, dirRPath, dirWPath, filterList)
	}
	return nil
}

/*
*	downloadFile() takes a 'readPath' which corresponds to a file in the cf app. The file is
*	downloaded using the os/exec library to call cf files with the given readPath. The output is
*	written to a file at writePath.
 */
func downloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error {
	defer fileDownloadGroup.Done()

	output, err := getFile(readPath)
	err = writeFile(readPath, writePath, output, err)
	check(cliError{err: err, errMsg: "Called by: downloadFile 2"})

	return nil
}

func getFile(readPath string) ([]byte, error) {
	// call cf files using os/exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()

	return output, err
}

func writeFile(readPath, writePath string, output []byte, err error) error {
	file := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid files or download issues
	CheckDownload(readPath, file, err)

	if verbose {
		fmt.Printf("Writing file: %s\n", readPath)
	} else {
		// increment download counter for commandline display
		// see consoleWriter()
		filesDownloaded++
	}

	fileAsString := file[2]
	// write downloaded file to writePath
	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	return err
}

func CheckDownload(readPath string, file []string, err error) error {
	// check for invalid file error.
	// some files are inaccesible from the cf files (permission issues) this is rare but we need to
	// alert users if it occurs. It usually happens in vendor files.
	if strings.Contains(file[1], "FAILED") {
		errmsg := ansi.Color(" Server Error: '"+readPath+"' not downloaded", "yellow")
		if onWindows == true {
			errmsg = " Server Error: '" + readPath + "' not downloaded"
		}
		failedDownloads = append(failedDownloads, errmsg)
		if verbose {
			fmt.Println(errmsg)
		}
		return errors.New("download failed")
	} else {
		// check for other errors
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}
	return nil
}
