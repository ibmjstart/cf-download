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
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/ibmjstart/cf-download/cmd_exec"
	"github.com/ibmjstart/cf-download/dir_parser"
	"github.com/ibmjstart/cf-download/downloader"
	"github.com/ibmjstart/cf-download/filter"
	"github.com/mgutz/ansi"
)

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
 */
type downloadPlugin struct{}

// error struct that allows appending error messages
type cliError struct { //we don't need this struct, see line #307
	err    error
	errMsg string
}

// contains flag values
type flagVal struct {
	Omitp_flag        *string
	OverWritep_flag   *bool
	MaxRoutinesp_flag *int
	Instancep_flag    *int
	Verbosep_flag     *bool
}

var (
	rootWorkingDirectory string
	appName              string
	instance             string //these vars shouldn't be global, and are not necessary entirely, just use flagVal struct
	verbose              bool
	onWindows            bool
	omitp                bool
	filesDownloaded      int
	failedDownloads      []string
	parser               dir_parser.Parser
	dloader              downloader.Downloader
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
	if args[0] != "download" {
		os.Exit(0)
	}

	// start time for download timer
	start := time.Now()
	proceed := true  //no need for this var

	// disables ansi text color on windows
	onWindows = IsWindows()

	if len(args) < 2 {
		cmd := exec.Command("cf", "help", "download")
		output, err := cmd.CombinedOutput()
		check(cliError{err: err, errMsg: ""})
		fmt.Println("\nError: Missing App Name")
		fmt.Printf("%s", output)
		proceed = false //we can just exit here instead of line #109
	}

	copyOfArgs, proceed, flagVals := ParseFlags(args, proceed)

	if proceed == false {
		os.Exit(1)
	} else {
		// flag variables
		maxRoutines := *flagVals.MaxRoutinesp_flag //make MaxRoutinesp_flag a int instead of *int, and use MaxRoutinesp_flag directly instad of assigning to maxRoutines
		overWrite := *flagVals.OverWritep_flag //same as above
		instance = strconv.Itoa(*flagVals.Instancep_flag)//same as above
		verbose = *flagVals.Verbosep_flag //same as above
		runtime.GOMAXPROCS(maxRoutines) //? why limit cpu cores, this doesn't control go routine limit    // set number of go routines
		filterList := filter.GetFilterList(*flagVals.Omitp_flag, verbose) // get list of things to not download

		workingDir, err := os.Getwd()
		check(cliError{err: err, errMsg: "Called by: Getwd"})
		rootWorkingDirectory, startingPath := GetDirectoryContext(workingDir, copyOfArgs)

		// ensure cf_trace is disabled, otherwise parsing breaks
		if os.Getenv("CF_TRACE") == "true" {
			fmt.Println("\nError: environment variable CF_TRACE is set to true. This prevents download from succeeding.")
			return
		}

		// prevent overwriting files
		if Exists(rootWorkingDirectory) && overWrite == false {
			fmt.Println("\nError: destination path", rootWorkingDirectory, "already Exists and is not an empty directory.\n\nDelete it or use 'cf download APP_NAME --overwrite'")
			return
		}

		cmdExec := cmd_exec.NewCmdExec()
		parser = dir_parser.NewParser(cmdExec, appName, instance, onWindows, verbose)
		dloader = downloader.NewDownloader(cmdExec, &wg, appName, instance, rootWorkingDirectory, verbose, onWindows)

		// parse the directory
		files, dirs := parser.ExecParseDir(startingPath)

		// stop consoleWriter
		quit := make(chan int)

		// disable consoleWriter if verbose
		if verbose == false {
			go consoleWriter(quit)
		}

		// Start the download
		wg.Add(1)
		dloader.Download(files, dirs, startingPath, rootWorkingDirectory, filterList)

		// Wait for download goRoutines
		wg.Wait()
		fmt.Printf("Files completed: %d", filesDownloaded)
		// stop console writer
		if verbose == false {
			quit <- 0
		}

		getFailedDownloads()
		PrintCompletionInfo(start)

	}
}

/*
*	-----------------------------------------------------------------------------------------------
* 	------------------------------------- Helper Functions ----------------------------------------
* 	-----------------------------------------------------------------------------------------------
 */

func getFailedDownloads() {
	failedDownloads = append(parser.GetFailedDownloads(), dloader.GetFailedDownloads()...)
}

func GetDirectoryContext(workingDir string, copyOfArgs []string) (string, string) {
	rootWorkingDirectory := workingDir + "/" + appName + "-download/"

	// append path if provided as arguement
	startingPath := "/"
	if len(copyOfArgs) > 2 && !strings.HasPrefix(copyOfArgs[2], "-") {
		startingPath = copyOfArgs[2]
		if !strings.HasSuffix(startingPath, "/") {
			startingPath += "/"
		}
		if strings.HasPrefix(startingPath, "/") { //don't need this block to Trim and append at line 193
			startingPath = strings.TrimPrefix(startingPath, "/")
		}
		rootWorkingDirectory += startingPath //lose if block above, here should be rootWorkingDirectory += strings.TrimPrefix(startingPath,"/")
		if !strings.HasPrefix(startingPath, "/") {
			startingPath = "/" + startingPath
		}
	}

	return rootWorkingDirectory, startingPath
}

func ParseFlags(args []string, proceed bool) ([]string, bool, flagVal) {

	// Create flagSet f1
	f1 := flag.NewFlagSet("f1", flag.ContinueOnError)

	// Create flags
	omitp := f1.String("omit", "", "--omit path/to/some/file") //don't need to assign to var, see line 243
	overWritep := f1.Bool("overwrite", false, "--overwrite")
	maxRoutinesp := f1.Int("routines", 200, "--routines [numOfRoutines]")
	instancep := f1.Int("i", 0, "-i [instanceNum]")
	verbosep := f1.Bool("verbose", false, "--verbose") 

	// need to copy args[] for later as they will be overwritten
	copyOfArgs := make([]string, len(args))  //I don't see how args will be overwritten, copying args not necessary
	for i := 0; i < len(args); i++ {
		copyOfArgs[i] = args[i]
	}
	//we can just f1.Parse(args[1:]), no need to deal with os.Args
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

	flagVals := flagVal{
		Omitp_flag:        omitp, //inline init here. Omitp_flag f1.String("omit", "", "--omit path/to/some/file")
		OverWritep_flag:   overWritep,
		MaxRoutinesp_flag: maxRoutinesp,
		Instancep_flag:    instancep,
		Verbosep_flag:     verbosep,
	}

	return copyOfArgs, proceed, flagVals
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
			filesDownloaded := dloader.GetFilesDownloadedCount()
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

// prints all the info you see at program finish
func PrintCompletionInfo(start time.Time) {
	// let user know if any files were inaccessible
	if len(failedDownloads) == 1 {
		fmt.Println("")
		fmt.Println(len(failedDownloads), "file or directory was not downloaded (permissions issue or corrupt):")
		PrintSlice(failedDownloads)
	} else if len(failedDownloads) > 1 {
		fmt.Println("")
		fmt.Println(len(failedDownloads), "files or directories were not downloaded (permissions issue or corrupt):")
		PrintSlice(failedDownloads)
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

// error check function
func check(e cliError) { //not need to make cliError{}, what about "func check(e err, errMsg string){}"
	if e.err != nil {
		fmt.Println("\nError: ", e.err)
		if e.errMsg != "" {
			fmt.Println("Message: ", e.errMsg)
		}
		os.Exit(1)
	}
}

// prints slices in readable format
func PrintSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index+1, ": ", val)
	}
	return nil
}

func IsWindows() bool { //what about runtime.GOOS , more reliable than checking for PathSeparator
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// Exists returns whether the given file or directory Exists or not
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	check(cliError{err: err, errMsg: "Called by: Exists"})
	return false
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
		Name: "download", //maybe a more descriptive name for the plugin?
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 2,
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
						"omit \"path/to/file\"": "Omit directories or files delimited by semicolons",
						"routines":              "Max number of concurrent subroutines (default 200)",
						"i":                     "Instance",
					},
				},
			},
		},
	}
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
