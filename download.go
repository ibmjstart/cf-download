package main

import (
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"os"
	"strings"
)

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
*
 */
type downloadPlugin struct{}

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

func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	connection = cliConnection

	// Ensure that we called the command download
	if args[0] == "download" {

		if len(args) != 2 {
			fmt.Println("\nError: Missing App Name")
			os.Exit(1)
		}

		workingDir, err := os.Getwd()
		check(err)

		appName = args[1]
		rootWorkingDirectory = workingDir + "/" + appName + "-download/"

		output, err := getDirString(appName)

		check(err)

		// Print the output returned from the CLI command.
		files, dirs := parseDir(output)

		fmt.Println("---------- Files ----------")
		for index, val := range files {
			fmt.Println("#", index, " value: ", val)
		}
		fmt.Println("")

		fmt.Println("---------- Directories ----------")
		for index, val := range dirs {
			fmt.Println("#", index, " value: ", val)
		}
		fmt.Println("")

		fmt.Println("Starting file download!")
		download(files, dirs, "/", rootWorkingDirectory)
	}
}

func getDirString(appName string) (string, error) {
	output, err := connection.CliCommandWithoutTerminalOutput("files", appName)
	check(err)

	return output[3], nil
}

func parseDir(dir string) ([]string, []string) {
	// PRE: takes in a string of the directory and filesizes
	// POST: FCTVAL==[]string of only filenames and folders without filesizes
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
	fmt.Printf("Writing directory: %s\n", readPath)

	file, err := connection.CliCommandWithoutTerminalOutput("files", appName, readPath)
	check(err)

	fileAsString := file[3]

	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	check(err)
	return nil
}

func downloadDir(path string) error {

	return nil
}

func download(files, dirs []string, readPath, writePath string) error {

	if files == nil && dirs == nil {
		msg := ansi.Color("File Successfully Downloaded!", "green+b")
		defer fmt.Println(msg)
		os.Exit(0)
	}
	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(err)

	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val
		downloadFile(fileRPath, fileWPath)

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
func check(e error) {
	if e != nil {
		fmt.Println("\nError: ", e)
		os.Exit(1)
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
