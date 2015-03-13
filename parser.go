package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mgutz/ansi"
)

/*
*	execParseDir() uses os/exec to shell out commands to cf files with the given readPath. The returned
*	text contains file and directory structure which is then parsed into two slices, dirs and files. dirs
*	contains the names of directories in readPath, files contians the file names. dirs and files are returned
* 	to be downloaded by download() and downloadFile() respectively.
 */
func execParseDir(readPath string) ([]string, []string) {
	// make the cf files call using exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()
	dirSlice := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid or missing app
	if strings.Contains(dirSlice[1], "not found") {
		errmsg := ansi.Color("Error: "+appName+" app not found (check space and org)", "red+b")
		if onWindows == true {
			errmsg = "Error: " + appName + " app not found (check space and org)"
		}
		fmt.Println(errmsg)
	}

	// this usually gets called when an app is not running and you attempt to download it.
	dir := dirSlice[2]
	if strings.Contains(dir, "error code: 190001") {
		errmsg := ansi.Color("App not found, or the app is in stopped state", "red+b")
		if onWindows == true {
			errmsg = "App not found, possibly not yet running"
		}
		fmt.Println(errmsg)
		check(cliError{err: err, errMsg: "App not found"})
	}

	// handle an empty directory
	if strings.Contains(dir, "No files found") {
		return nil, nil
	} else {
		//check(cliError{err: err, errMsg:"Directory or file not found. Check filename or path on command line"})
		check(cliError{err: err, errMsg: "Called by: ExecParseDir [cf files " + appName + " " + readPath + "]"})
	}

	// directory inaccessible due to lack of permissions
	if strings.Contains(dirSlice[1], "FAILED") {
		errmsg := ansi.Color(" Server Error: '"+readPath+"' not downloaded", "yellow")
		if onWindows == true {
			errmsg = " Server Error: '" + readPath + "' not downloaded"
		}
		failedDownloads = append(failedDownloads, errmsg)
		if verbose {
			fmt.Println(errmsg)
		}
		return nil, nil
	} else {
		// check for other errors
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}

	// parse the returned output into files and dirs slices
	filesSlice := strings.Fields(dir)
	var files, dirs []string
	var name string
	for i := 0; i < len(filesSlice); i++ {
		if strings.HasSuffix(filesSlice[i], "/") {
			name += filesSlice[i]
			dirs = append(dirs, name)
			name = ""
		} else if isDelimiter(filesSlice[i]) {
			if len(name) > 0 {
				name = strings.TrimSuffix(name, " ")
				files = append(files, name)
			}
			name = ""
		} else {
			name += filesSlice[i] + " "
		}
	}
	return files, dirs
}

func isDelimiter(str string) bool {
	match, _ := regexp.MatchString("^[0-9]([0-9]|.)*(G|M|B|K)$", str)
	if match == true || str == "-" {
		return true
	}
	return false
}
