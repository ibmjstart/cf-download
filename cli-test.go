package main

import (
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	connection           plugin.CliConnection
)

// global wait group for all download threads
var wg sync.WaitGroup

func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// start time for download timer
	connection = cliConnection

	proceed := true

	if len(args) < 2 {
		cmd := exec.Command("cf", "help", "download")
		output, err := cmd.CombinedOutput()
		check(cliError{err: err, errMsg: ""})
		fmt.Println("\nError: Missing App Name")
		fmt.Printf("%s", output)
		proceed = false
	}

	// need to copy args[] for later as they will be overwritten
	copyOfArgs := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		copyOfArgs[i] = args[i]
	}

	appName = args[1]
	fmt.Println("appname: ", appName)

	if proceed == false {
		os.Exit(1)
	} else {

		workingDir, err := os.Getwd()
		check(cliError{err: err, errMsg: "Called by: Run"})
		rootWorkingDirectory = workingDir + "/" + appName + "-download/"

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

		// This code will write the contents of logs/staging_task.log to file in a loop.

		//********************************************************************************************************************
		// Test ************************************* USAGE: cf download APP_NAME ********************************************
		//********************************************************************************************************************

		err = os.MkdirAll(rootWorkingDirectory, 0755)
		check(cliError{err: err, errMsg: "Called by: download"})
		for i := 0; i < 10; i++ {

			// A lower wait time will cause errors (this is why we use exec)
			time.Sleep(100 * time.Millisecond)

			wg.Add(1)
			go func(writePath string, num int) {
				defer wg.Done()
				writePath += "log-" + strconv.Itoa(num)

				// modify the "logs/staging_task.log" to get different files
				file, err := connection.CliCommandWithoutTerminalOutput("files", appName, "logs/staging_task.log")
				check(cliError{err: err, errMsg: ""})
				fmt.Printf("Writing file: %s\n", writePath)
				fileAsString := file[3]

				err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
				check(cliError{err: err, errMsg: ""})
			}(rootWorkingDirectory, i)
		}

		//********************************************************************************************************************
		//********************************************************************************************************************
		//********************************************************************************************************************

		wg.Wait()
		fmt.Println("COMPLETE")
		return
	}

}

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
						"omit \"path/to/file\"": "Omit directories or files delimited by semicolons",
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
		fmt.Println(index, ": ", val)
	}
	return nil
}

func main() {
	plugin.Start(new(downloadPlugin))
}
