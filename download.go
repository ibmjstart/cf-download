/*
* IBM jStart team cf download cli Plugin
* A plugin for downloading contents of a running app's file directory
*
* Authors: Miguel Clement, Jake Eden
* Date: 3/5/2015
*
* for cross platform compiling use gox (https://github.com/mitchellh/gox)
* gox compile command: gox -output="binaries/{{.OS}}/{{.Arch}}/cf-download" -os="linux darwin windows"
 */

package main

import (
	"flag"
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
 */
type downloadPlugin struct{}

// error struct that allows appending error messages
type cliError struct {
	err    error
	errMsg string
}

var (
	rootWorkingDirectory string   //
	appName              string   //
	instance             string   //
	verbose              bool     //
	failedDownloads      []string //
	filesDownloaded      int      //
	onWindows            bool     //
)

// global wait group for all download threads
var wg sync.WaitGroup

/*
*	This function must be implemented by any plugin because it is part of the
*	plugin interface defined by the core CLI.
*
*	Run(....) is the entry point when the core CLI is invoking a command defined
*	by a plugin. The first parameter, plugin.CliConnection, is a struct that can
*	be used to invoke cli commands. The second paramter, args, is a slice of
*	strings. args[0] will be the name of the command, and will be followed by
*	any additional arguments a cli user typed in.
*
*	Any error handling should be handled with the plugin itself (this means printing
*	user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
*	1 should the plugin exits nonzero.
 */
func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// start time for download timer
	start := time.Now()
	proceed := true

	// disables ansi text color on windows
	onWindows = isWindows()

	if len(args) < 2 {
		cmd := exec.Command("cf", "help", "download")
		output, err := cmd.CombinedOutput()
		check(cliError{err: err, errMsg: ""})
		fmt.Println("\nError: Missing App Name")
		fmt.Printf("%s", output)
		proceed = false
	}

	// Create flagSet f1
	f1 := flag.NewFlagSet("f1", flag.ContinueOnError)

	// Create flags
	omitp := f1.String("omit", "", "--omit path/to/some/file")
	overWritep := f1.Bool("overwrite", false, "--overwrite")
	maxRoutinesp := f1.Int("routines", 200, "--routines [numOfRoutines]")
	instancep := f1.Int("i", 0, "-i [instanceNum]")
	verbosep := f1.Bool("verbose", false, "--verbose")

	// need to copy args[] for later as they will be overwritten
	copyOfArgs := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		copyOfArgs[i] = args[i]
	}

	// flag package parses os.Args so we need to set it correctly
	if len(args) > 2 && !strings.HasPrefix(args[2], "-") { // if there is a path as in 'cf download path' vs. 'cf download'
		os.Args = append(os.Args[:1], args[3:]...)
	} else {
		os.Args = append(os.Args[:1], args[2:]...)
	}

	// check for misplaced flags
	appName = copyOfArgs[1]
	if strings.HasPrefix(appName, "-") || strings.HasPrefix(appName, "--") {
		fmt.Println("\nError: App name begins with '-' or '--'. correct flag usage: 'cf download APP_NAME [--flags]'")
		proceed = false
	}

	// Check for parsing errors
	if err := f1.Parse(os.Args[1:]); err != nil {
		fmt.Println("\nError: ", err, "\n")
		cmd := exec.Command("cf", "help", "download")
		output, err := cmd.CombinedOutput()
		check(cliError{err: err, errMsg: ""})
		fmt.Printf("%s", output)
		proceed = false
	}

	if proceed == false {
		os.Exit(1)
	} else {
		// flag variables
		maxRoutines := *maxRoutinesp
		overWrite := *overWritep
		instance = strconv.Itoa(*instancep)
		verbose = *verbosep

		runtime.GOMAXPROCS(maxRoutines) // set number of go routines

		filterList := getFilterList(*omitp) // get list of things to not download

		workingDir, err := os.Getwd()
		check(cliError{err: err, errMsg: "Called by: Run"})
		rootWorkingDirectory = workingDir + "/" + appName + "-download/"

		// ensure cf_trace is disabled, otherwise parsing breaks
		if os.Getenv("CF_TRACE") == "true" {
			fmt.Println("\nError: environment variable CF_TRACE is set to true. This prevents download from succeeding.")
			return
		}

		// prevent overwriting files
		if exists(rootWorkingDirectory) && overWrite == false {
			fmt.Println("\nError: destination path", rootWorkingDirectory, "already exists and is not an empty directory.\n\nDelete it or use 'cf download APP_NAME --overwrite'")
			return
		}

		// append path if provided as arguement
		startingPath := "/"
		if len(args) > 2 && !strings.HasPrefix(copyOfArgs[2], "-") {
			startingPath = copyOfArgs[2]
			if !strings.HasSuffix(startingPath, "/") {
				startingPath += "/"
			}
			if strings.HasPrefix(startingPath, "/") {
				startingPath = strings.TrimPrefix(startingPath, "/")
			}
			rootWorkingDirectory += startingPath
			if !strings.HasPrefix(startingPath, "/") {
				startingPath = "/" + startingPath
			}
		}

		// parse the directory
		files, dirs := execParseDir(startingPath)

		if !verbose {
			fmt.Printf("Files completed: %d", filesDownloaded)
		}

		// stop consoleWriter
		quit := make(chan int)
		// disable consoleWriter if verbose
		if verbose == false {
			go consoleWriter(quit)
		}

		// Start the download
		wg.Add(1)
		download(files, dirs, startingPath, rootWorkingDirectory, filterList)

		// Wait for download goRoutines
		wg.Wait()

		// stop console writer
		if verbose == false {
			quit <- 0
		}

		// let user know if any files were inaccessible
		if len(failedDownloads) == 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "file or directory was not downloaded (permissions issue or corrupt):")
			printSlice(failedDownloads)
		} else if len(failedDownloads) > 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "files or directories were not downloaded (permissions issue or corrupt):")
			printSlice(failedDownloads)
		}

		// display runtime
		elapsed := time.Since(start)
		elapsedString := strings.Split(elapsed.String(), ".")[0]
		elapsedString = strings.TrimSuffix(elapsedString, ".") + "s"
		fmt.Println("\nDownload time: " + elapsedString)

		msg := ansi.Color(appName+" Successfully Downloaded!", "green+b")
		if onWindows == true {
			msg = "Successfully Downloaded!"
		}
		fmt.Println(msg)
	}
}

func getFilterList(omitString string) []string {
	// POST: FCTVAL== slice of strings (paths and files) to filter
	var filterList []string // filtered list to be returned

	// Add .cfignore files to filterList
	content, err := ioutil.ReadFile(".cfignore")
	if err != nil && verbose {
		fmt.Println("[ Info: ", err, "]")
	} else {
		lines := strings.Split(string(content), "\n") // get each line in .cfignore

		// Remove any leading forward slashes
		for i := 0; i < len(lines); i++ {
			lines[i] = strings.TrimPrefix(lines[i], "/")
		}

		filterList = append(filterList, lines[0:]...)

		// remove empty strings that we got from the last line
		if len(filterList) > 0 && filterList[len(filterList)-1] == "" {
			filterList = filterList[:len(filterList)-1]
		}
	}

	// Add the path from the --omit param to filterList
	if omitString != "" {

		allOmits := strings.Split(omitString, ";")

		// Parse for each path and remove leading forward slashes
		for i := 0; i < len(allOmits); i++ {
			allOmits[i] = strings.TrimSpace(allOmits[i])
			allOmits[i] = strings.TrimPrefix(allOmits[i], "/")
		}
		filterList = append(filterList, allOmits[0:]...)
	}

	var returnList []string // filtered strings to be returned

	// Remove any trailing forward slashes in the filterList[ex: app/ becomes app]
	for i, _ := range filterList {
		filterList[i] = strings.TrimSuffix(filterList[i], "/")
		filterList[i] = "/" + filterList[i]

		// don't include any empty strings, which only have a forward slash
		if strings.TrimSpace(filterList[i]) != "/" {
			returnList = append(returnList, filterList[i])
		}
	}

	return returnList
}

/*
*	consoleWriter prints the current number of files downloaded. It is polled every 350 milleseconds
* 	disabled if using verbose flag.
 */
func consoleWriter(quit chan int) {
	count := 0
	for {
		select {
		case <-quit:
			return
		default:
			switch count = (count + 1) % 4; count {
			case 0:
				fmt.Printf("\rFiles completed: %d \\ ", filesDownloaded)
			case 1:
				fmt.Printf("\rFiles completed: %d | ", filesDownloaded)
			case 2:
				fmt.Printf("\rFiles completed: %d / ", filesDownloaded)
			case 3:
				fmt.Printf("\rFiles completed: %d --", filesDownloaded)
			}
			time.Sleep(350 * time.Millisecond)
		}
	}
}

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
		errmsg := ansi.Color("App not found, possibly not yet running", "red+b")
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

/*
*	downloadFile() takes a 'readPath' which corresponds to a file in the cf app. The file is
*	downloaded using the os/exec library to call cf files with the given readPath. The output is
*	written to a file at writePath.
 */
func downloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error {
	defer fileDownloadGroup.Done()

	// call cf files using os/exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()
	file := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid file error.
	// some files are inaccesible from the cf files (permission issues) this is rare but we need to
	// alert users if it occurs. It usually happens in vendor files.
	fileAsString := file[2]
	if strings.Contains(file[1], "FAILED") {
		errmsg := ansi.Color(" Server Error: '"+readPath+"' not downloaded", "yellow")
		if onWindows == true {
			errmsg = " Server Error: '" + readPath + "' not downloaded"
		}
		failedDownloads = append(failedDownloads, errmsg)
		if verbose {
			fmt.Println(errmsg)
		}
		return nil
	} else {
		// check for other errors
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}
	if verbose {
		fmt.Printf("Writing file: %s\n", readPath)
	} else {
		// increment download counter for commandline display
		// see consoleWriter()
		filesDownloaded++
	}
	// write downloaded file to writePath
	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	check(cliError{err: err, errMsg: "Called by: downloadFile 2"})

	return nil
}

func checkToFilter(appPath string, filterList []string) bool {
	appPath = strings.TrimSuffix(appPath, "/")
	comparePath1 := strings.TrimPrefix(appPath, rootWorkingDirectory)

	for _, item := range filterList {
		if comparePath1 == item {
			return true
		}
	}

	return false
}

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
*	This function must be implemented as part of the	plugin interface
*	defined by the core CLI.
*
*	GetMetadata() returns a PluginMetadata struct. The first field, Name,
*	determines the name of the plugin which should generally be without spaces.
*	If there are spaces in the name a user will need to properly quote the name
*	during uninstall otherwise the name will be treated as seperate arguments.
*	The second value is a slice of Command structs. Our slice only contains one
*	Command Struct, but could contain any number of them. The first field Name
*	defines the command `cf basic-plugin-command` once installed into the CLI. The
*	second field, HelpText, is used by the core CLI to display help information
*	to the user in the core commands `cf help`, `cf`, or `cf -h`.
 */
func (c *downloadPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "download",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 1,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "download",
				HelpText: "Download contents of a running app's file directory",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "cf download APP_NAME [PATH] [--overwrite] [--verbose] [--omit ommited_paths] [--routines num_max_routines] [-i instance_num]",
					Options: map[string]string{
						"overwrite":             "Overwrite existing files",
						"verbose":               "Verbose output",
						"omit \"path/to/file\"": "Omit directories or files delimited by commas",
						"routines":              "Max number of concurrent subroutines (default 200)",
						"i":                     "Instance",
					},
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

// prints slices in readable format
func printSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index+1, ": ", val)
	}
	return nil
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

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

func isDelimiter(str string) bool {
	match, _ := regexp.MatchString("^[0-9]([0-9]|.)*(G|M|B|K)$", str)
	if match == true || str == "-" {
		return true
	}
	return false
}

/*
* Unlike most Go programs, the `Main()` function will not be used to run all of the
* commands provided in your plugin. Main will be used to initialize the plugin
* process, as well as any dependencies you might require for your
* plugin.
 */
func main() {
	// Any initialization for your plugin can be handled here

	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(downloadPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
