package dir_parser

import (
	"fmt"
	"github.com/ibmjstart/cf-download/cmd_exec"
	"github.com/mgutz/ansi"
	"os"
	"regexp"
	"strings"
	"time"
)

type Parser interface {
	ExecParseDir(readPath string) ([]string, []string)
	GetFailedDownloads() []string
	GetDirectory(readPath string) (string, string)
}

type parser struct {
	cmdExec         cmd_exec.CmdExec
	appName         string
	instance        string
	onWindows       bool
	verbose         bool
	failedDownloads []string
}

func NewParser(cmdExec cmd_exec.CmdExec, appName, instance string, onWindows, verbose bool) *parser {
	return &parser{
		cmdExec:   cmdExec,
		appName:   appName,
		instance:  instance,
		onWindows: onWindows,
		verbose:   verbose,
	}
}

/*
*	execParseDir() uses os/exec to shell out commands to cf files with the given readPath. The returned
*	text contains file and directory structure which is then parsed into two slices, dirs and files. dirs
*	contains the names of directories in readPath, files contians the file names. dirs and files are returned
* 	to be downloaded by download() and downloadFile() respectively.
 */
func (p *parser) ExecParseDir(readPath string) ([]string, []string) {
	dir, status := p.GetDirectory(readPath)

	if status == "OK" {
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
	} else {
		//error occured, error message displayed by GetDirectory()
		if len(dir) > 0 {
			fmt.Println(dir)
		}
		if readPath == "/" {
			os.Exit(1)
		}
	}

	return nil, nil
}

/*
*	getDirectory will return the directory as a string ready for parsing.
*	There is a status code returned as well, this is not necessary but helps with testing.
 */
func (p *parser) GetDirectory(readPath string) (string, string) {

	// make the cf files call using exec
	output, err := p.cmdExec.GetFile(p.appName, readPath, p.instance)
	dirSlice := strings.SplitAfterN(string(output), "\n", 3)

	// if cf files fails to get directory, retry (this code is not covered in tests)
	if len(dirSlice) < 2 {
		iterations := 0
		for len(dirSlice) < 2 && iterations < 10 {
			time.Sleep(3 * time.Second)
			output, err = p.cmdExec.GetFile(p.appName, readPath, p.instance)
			dirSlice = strings.SplitAfterN(string(output), "\n", 3)
			iterations++
		}
	}

	if len(dirSlice) >= 2 && strings.Contains(dirSlice[1], "OK") {
		if strings.Compare(dirSlice[2], "\nNo files found\n") == 0 {
			return "", "noFiles"
		}
		return dirSlice[2], "OK"
	} else {
		message := createMessage(" Server Error: '"+readPath+"' not downloaded", "yellow", p.onWindows)

		p.failedDownloads = append(p.failedDownloads, message)

		if p.verbose {
			fmt.Println(message)
			// check for other errors
			if err != nil {
				PrintSlice(dirSlice)
			}
		}

		return string(output), "Failed"
	}
}

func (p *parser) GetFailedDownloads() []string {
	return p.failedDownloads
}

func isDelimiter(str string) bool {
	match, _ := regexp.MatchString("^[0-9]([0-9]|.)*(G|M|B|K)$", str)
	if match == true || str == "-" {
		return true
	}
	return false
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
