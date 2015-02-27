package main

import (
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
*
 */
type downloadPlugin struct{}

type file struct {
	writePath, content string
}

type cliError struct {
	err    error
	errMsg string
}

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

var (
	connection           plugin.CliConnection
	rootWorkingDirectory string
	appName              string
)

var master = make(chan string)
var wg sync.WaitGroup

func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	connection = cliConnection

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

		startingPath := "/"
		if len(args) == 3 {
			startingPath = args[2]
			if !strings.HasSuffix(startingPath, "/") {
				startingPath += "/"
			}
		}

		files, dirs := parseDir("/app/public/css/")

		fmt.Println("Starting file download!")
		wg.Add(1)
		go download(files, dirs, "/app/public/css/", rootWorkingDirectory)
		files, dirs = parseDir("/app/public/css/")
		wg.Add(1)
		go download(files, dirs, "/app/public/images/", rootWorkingDirectory)
		files, dirs = parseDir("/app/public/images/")
		wg.Add(1)
		go download(files, dirs, "/app/public/js/", rootWorkingDirectory)
		files, dirs = parseDir("/app/public/js/")
		/*
			go download(files, dirs, startingPath, rootWorkingDirectory)
			msg := ansi.Color("File Successfully Downloaded!", "green+b")
			defer fmt.Println(msg)
		*/

		wg.Wait()
		fmt.Println("EXITING!!!!")
		time.Sleep(2 * time.Second)
	}
}

func parseDir(readPath string) ([]string, []string) {
	fmt.Println("\ncf files", appName, readPath)
	dirSlice, err := connection.CliCommandWithoutTerminalOutput("files", appName, readPath)
	if strings.Contains(dirSlice[1], "not found") {
		errmsg := ansi.Color("Error: "+appName+" app not found (check space and org)", "red+b")
		fmt.Println(errmsg)
	}
	dir := dirSlice[3]

	if strings.Contains(dir, "No files found") {
		return nil, nil
	} else {
		printSlice(dirSlice)
		check(cliError{err: err, errMsg: "Called by: parseDir [cf files " + appName + " " + readPath + "]"})
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

func downloadFile(readPath, writePath string) error {
	fmt.Println("\ncf files", appName, readPath)

	file, err := connection.CliCommandWithoutTerminalOutput("files", appName, readPath)

	if strings.Contains(file[2], "status code") {
		errmsg := ansi.Color("Server Error: "+readPath+"not downloaded", "red")
		fmt.Println(errmsg)
		return nil
	} else {
		check(cliError{err: err, errMsg: "Called by: downloadFile"})
	}

	fmt.Printf("Writing file: %s\n", readPath)
	fileAsString := file[3]

	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	check(cliError{err: err, errMsg: "Called by: downloadFile"})
	return nil
}

func download(files, dirs []string, readPath, writePath string) error {

	fmt.Println("ReadPath: ", readPath, "writePath: ", writePath)
	fmt.Println("---------- Files ----------")
	printSlice(files)
	fmt.Println("------- Directories -------")
	printSlice(dirs)

	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(cliError{err: err, errMsg: "Called by: download"})

	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val
		downloadFile(fileRPath, fileWPath)
	}

	for _, val := range dirs {
		dirWPath := writePath + val
		dirRPath := readPath + val
		/*//************ REMOVE ***************************************************** REMOVE
		if strings.Contains(val, "app") {
			continue
		}
		//************ REMOVE ***************************************************** REMOVE*/
		files, dirs = parseDir(dirRPath)

		wg.Add(1)
		go download(files, dirs, dirRPath, dirWPath)
	}
	wg.Done()

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
		fmt.Println("#", index, " value: ", val)
	}
	fmt.Println("")
	return nil
}

/*
* Unlike most Go programs, the `Main()` function will not be used to run all of the
* commands provided in your plugin. Main will be used to initialize the plugin
* process, as well as any dependencies you might require for your
* plugin.
 */
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "github.com/cloudfoundry/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(downloadPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
