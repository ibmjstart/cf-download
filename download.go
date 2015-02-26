package main

import (
	"fmt"
	"github.com/mgutz/ansi"
	"github.com/cloudfoundry/cli/plugin"
	"io/ioutil"
	"net/http"
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
func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[0] == "download" {
		if(len(args) != 2){
			fmt.Println("\nError: Missing App Name")
        	os.Exit(1)
		}

		appName := args[1]

		output, err := getDirString(cliConnection, appName)
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
		pull()
	}
}

func getDirString(cliConnection plugin.CliConnection, appName string) (string, error) {
	appGuidSlice, err := cliConnection.CliCommandWithoutTerminalOutput("app", appName, "--guid")
	check(err)
	appGuid := strings.TrimSpace(appGuidSlice[0])

	url := fmt.Sprintf("/v2/apps/%s/instances/%d/files/%s", appGuid, 0, "/app")
	fmt.Println(url)
	curlRequest := []string{"curl", url}
	output, err := cliConnection.CliCommandWithoutTerminalOutput(curlRequest...)
	check(err)

	return output[0], nil
}

func parseDir(dir string) ([]string, []string) {
// PRE: takes in a string of the directory and filesizes
// POST: FCTVAL==[]string of only filenames and folders without filesizes
    filesSlice := strings.Fields(dir)
    var files, dirs []string
    for i := 0; i < len(filesSlice); i+=2 {
    	if strings.HasSuffix(filesSlice[i], "/") {
    		dirs = append(dirs, filesSlice[i])
    	} else {
			files = append(files, filesSlice[i])
    	}
            
    }
    return files, dirs
}

func pull() {
	response, err := http.Get("http://golang.org/")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)
		check(err)

		dir, err := os.Getwd()
		check(err)

		dir += "/temp/"
		//create dir if does not exist
		err = os.MkdirAll(dir, 0755)
		check(err)

		// write file to dir
		filename := "google.html"
		dir += filename
		fmt.Printf("Writing directory: %s\n", dir)
	    err = ioutil.WriteFile(dir, contents, 0644)
	    check(err)

	    msg := ansi.Color("File Successfully Downloaded!", "green+b")
	    defer fmt.Println(msg)
	}
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



