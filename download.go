package main

import (
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type downloadPlugin struct{}

type cliError struct {
	err    error
	errMsg string
}

var (
	connection           plugin.CliConnection
	rootWorkingDirectory string
	appName              string
	useExec              bool
	failedDownloads      []string
)

var wg sync.WaitGroup

func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	start := time.Now()
	maxRoutines := 200
	runtime.GOMAXPROCS(maxRoutines)
	connection = cliConnection
	useExec = true
	overWrite := false

	// Ensure that we called the command download
	if args[0] == "download" {

		if len(args) < 2 {
			fmt.Println("\nError: Missing App Name")
			os.Exit(1)
		}

		workingDir, err := os.Getwd()
		check(cliError{err: err, errMsg: "Called by: Run"})
		appName = args[1]
		rootWorkingDirectory = workingDir + "/" + appName + "-download/"

		if exists(rootWorkingDirectory) && overWrite == true {
			fmt.Println("\nError: destination path", rootWorkingDirectory, "already exists and is not an empty directory. delete it or use 'cf download APP_NAME -f'")
			return
		}

		startingPath := "/"
		if len(args) == 3 {
			startingPath = args[2]
			if !strings.HasSuffix(startingPath, "/") {
				startingPath += "/"
			}
		}

		var files, dirs []string
		if useExec {
			files, dirs = execParseDir(startingPath)
		} else {
			files, dirs = parseDir(startingPath)
		}

		wg.Add(1)
		download(files, dirs, startingPath, rootWorkingDirectory)

		wg.Wait()
		if len(failedDownloads) == 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "file was not downloaded (inaccessible or corrupt):")
			printSlice(failedDownloads)
		} else if len(failedDownloads) > 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "files were not downloaded (inaccessible or corrupt):")
			printSlice(failedDownloads)
		}

		elapsed := time.Since(start)
		fmt.Printf("\nDownload time: %s\n", elapsed)

		msg := ansi.Color(appName+" Successfully Downloaded!", "green+b")
		fmt.Println(msg)
	}
}

func parseDir(readPath string) ([]string, []string) {
	dirSlice, err := connection.CliCommandWithoutTerminalOutput("files", appName, readPath)
	if strings.Contains(dirSlice[1], "not found") {
		errormsg := ansi.Color("Error: "+appName+" app not found (check space and org)", "red+b")
		fmt.Println(errormsg)
	}
	//printSlice(dir)
	dir := dirSlice[3]

	if strings.Contains(dir, "No files found") {
		return nil, nil
	} else {
		if err != nil {
			printSlice(dirSlice)
			check(cliError{err: err, errMsg: "Called by: ParseDir [cf files " + appName + " " + readPath + "]"})
		}
	}

	filesSlice := strings.Fields(dir)
	var files, dirs []string
	for i := 0; i < len(filesSlice); i += 2 {
		if strings.HasSuffix(filesSlice[i], "/") {
			dirs = append(dirs, filesSlice[i])
		} else {
			files = append(files, filesSlice[i])
		}

	}
	return files, dirs
}

func execParseDir(readPath string) ([]string, []string) {
	cmd := exec.Command("cf", "files", appName, readPath)

	output, err := cmd.CombinedOutput()

	dirSlice := strings.SplitAfterN(string(output), "\n", 3)
	if strings.Contains(dirSlice[1], "not found") {
		errmsg := ansi.Color("Error: "+appName+" app not found (check space and org)", "red+b")
		fmt.Println(errmsg)
	}

	dir := dirSlice[2]

	if strings.Contains(dir, "No files found") {
		return nil, nil
	} else {
		check(cliError{err: err, errMsg: "Called by: ExecParseDir [cf files " + appName + " " + readPath + "]"})
	}

	filesSlice := strings.Fields(dir)
	var files, dirs []string
	for i := 0; i < len(filesSlice); i += 2 {
		if strings.HasSuffix(filesSlice[i], "/") {
			dirs = append(dirs, filesSlice[i])
		} else {
			files = append(files, filesSlice[i])
		}

	}
	return files, dirs
}

func downloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error {
	defer fileDownloadGroup.Done()
	//fmt.Println("\ncf files", appName, readPath)

	cmd := exec.Command("cf", "files", appName, readPath)
	output, err := cmd.CombinedOutput()
	file := strings.SplitAfterN(string(output), "\n", 3)
	fileAsString := file[2]
	if strings.Contains(file[1], "FAILED") && strings.Contains(file[2], "status code: 404") {
		errormsg := ansi.Color("Server Error: '"+readPath+"' not downloaded", "red")
		failedDownloads = append(failedDownloads, errormsg)
		fmt.Println(errormsg)
		return nil
	} else {
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}

	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	check(cliError{err: err, errMsg: "Called by: downloadFile 2"})
	fmt.Printf("Wrote file: %s\n", readPath)
	return nil
}

func download(files, dirs []string, readPath, writePath string) error {
	defer wg.Done()

	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(cliError{err: err, errMsg: "Called by: download"})

	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val

		wg.Add(1)
		go downloadFile(fileRPath, fileWPath, &wg)
	}

	for _, val := range dirs {
		dirWPath := writePath + val
		dirRPath := readPath + val
		err := os.MkdirAll(dirWPath, 0755)
		check(cliError{err: err, errMsg: "Called by: download"})

		if useExec {
			files, dirs = execParseDir(dirRPath)
		} else {
			files, dirs = parseDir(dirRPath)
		}

		wg.Add(1)
		go download(files, dirs, dirRPath, dirWPath)
	}
	return nil
}

func (c *downloadPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "File Downloader",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "download",
				HelpText: "Download contents of targeted directory",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "download\n   cf download",
				},
			},
		},
	}
}

// error check function
func check(e cliError) {
	if e.err != nil {
		fmt.Println("\nError: ", e.err)
		if e.errMsg != "" {
			fmt.Println("Message: ", e.errMsg)
		}
		os.Exit(1)
	}
}

func printSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index+1, ": ", val)
	}
	return nil
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	check(cliError{err: err, errMsg: "Called by: exists"})
	return false
}

func main() {
	plugin.Start(new(downloadPlugin))

}
